#!/usr/bin/env bash

# Tests how new-failures.jq classifies one test's outcomes, given one outcome
# per runner call, in order, in each of the head, base, and retry runs, and how
# rerun.jq parses /rerun comments.

set -euo pipefail

DIR="$(CDPATH= cd "$(dirname "${0}")" && pwd)"
cd "${DIR}"

# name|head|base|head retry|base retry|expected kinds (failure, pass,
# flaky-failure, flaky-pass, not-run, misaligned, or none)
CASES=(
  "new failure|failed|passed|failed|passed|failure"
  "new failure without a retry|failed|passed|||failure"
  "new failure in an added suite|failed||failed||failure"
  "failure that passes when retried|failed|passed|passed|passed|flaky-failure"
  "failure that also fails at base when retried|failed|passed|failed|failed|flaky-failure"
  "failure whose retry couldn't run|failed|passed|not-run|passed|failure"
  "new failure in one of several calls|passed passed failed|passed passed passed|passed passed failed|passed passed passed|failure"
  "failure in one of several calls that doesn't repeat|passed passed failed|passed passed passed|passed passed passed|passed passed passed|flaky-failure"
  "failure in one call at head and another at base|failed passed|passed failed|||failure,pass"
  "new failure hidden by another call's flaky failure at base|passed passed failed|failed passed passed|passed passed failed|passed passed passed|failure,flaky-pass"
  "new failure in one call and a flaky one in another|failed failed|passed passed|failed passed|passed passed|failure"
  "results from calls that don't line up|failed passed|passed failed passed|||misaligned"
  "new failure in calls that don't line up|passed failed|passed passed passed|||failure,misaligned"
  "new pass|passed|failed|passed|failed|pass"
  "pass that fails at base only once|passed|failed|passed|passed|flaky-pass"
  "new pass in one of several calls|passed passed|passed failed|passed passed|passed failed|pass"
  "couldn't run at head|not-run|passed|||not-run"
  "couldn't run at head and failed when retried|not-run|passed|failed|passed|failure"
  "couldn't run at head and passed when retried|not-run|passed|passed|passed|none"
)

# Prints a runner report's tests for a test with the given outcomes, one for
# each call.
tests_json() {
  local outcomes="${1}"
  jq -cn --arg outcomes "${outcomes}" \
    '$outcomes | split(" ") | map(select(. != "")) | to_entries
    | map({name: "Suite/test", outcome: .value, knownFailing: false, call: .key})'
}

calls_json() {
  jq -cn '$ARGS.positional | map(split(" ") | map(select(. != "")) | length)
    | {head: .[0], base: .[1], headRetry: .[2], baseRetry: .[3]}' --args "${@}"
}

classify() {
  local head="${1}"
  local base="${2}"
  local head_retry="${3}"
  local base_retry="${4}"
  jq -n -L . \
    --arg implementation impl --arg target target --arg commit commit \
    --arg status ok --arg statusSide "" --argjson notes "[]" \
    --argjson knownFailingList false --arg retry done --arg notApplicable "" \
    --argjson head "[$(tests_json "${head}")]" \
    --argjson base "[$(tests_json "${base}")]" \
    --argjson headRetry "[$(tests_json "${head_retry}")]" \
    --argjson baseRetry "[$(tests_json "${base_retry}")]" \
    --argjson changes '[{"suites": []}]' \
    --argjson calls "$(calls_json "${head}" "${base}" "${head_retry}" "${base_retry}")" \
    -f new-failures.jq |
    jq -r '[
      (.newFailures[] | "failure"), (.newPasses[] | "pass"),
      (.flaky[] | "flaky-" + .change), (.notRun[] | "not-run"),
      (.notes[] | select(. == "calls-misaligned") | "misaligned")
    ] | if length == 0 then "none" else join(",") end'
}

# comment body|expected parse_command output, compact
COMMAND_CASES=(
  "/rerun connect-py https://github.com/connectrpc/connect-py/pull/463|{\"implementation\":\"connect-py\",\"pullRequest\":463}"
  "/rerun connect-py https://github.com/connectrpc/connect-py/pull/463/|{\"implementation\":\"connect-py\",\"pullRequest\":463}"
  "  /rerun   connect-py   https://github.com/connectrpc/connect-py/pull/463  \r|{\"implementation\":\"connect-py\",\"pullRequest\":463}"
  "/rerun connect-py https://github.com/connectrpc/connect-py/pull/463\nThe timeout fix.|{\"implementation\":\"connect-py\",\"pullRequest\":463}"
  "/rerun connect-py|{\"implementation\":\"connect-py\",\"pullRequest\":null}"
  "/rerun|{\"error\":\"usage\"}"
  "/rerun connect-py https://github.com/connectrpc/connect-py/pull/463 extra|{\"error\":\"usage\"}"
  "/rerun connect-cobol|{\"error\":\"unknown-implementation\"}"
  "/rerun connect-py https://github.com/connectrpc/connect-go/pull/463|{\"error\":\"invalid-pull-request\"}"
  "/rerun connect-py https://github.com/example/connect-py/pull/463|{\"error\":\"invalid-pull-request\"}"
  "/rerun connect-py https://github.com/connectrpc/connect-py/pull/463/files|{\"error\":\"invalid-pull-request\"}"
  "/rerun connect-py https://github.com/connectrpc/connect-py/pull/1234567890|{\"error\":\"invalid-pull-request\"}"
  "/rerunconnect-py|null"
  "Please /rerun connect-py|null"
  "\n/rerun connect-py|null"
)

check() {
  local name="${1}"
  local actual="${2}"
  local expected="${3}"
  if [ "${actual}" != "${expected}" ]; then
    echo "FAIL: ${name}: got ${actual}, want ${expected}" >&2
    failures=$((failures + 1))
  fi
}

# Reports for runs started by /rerun comments, as the run's conclusion, the
# artifacts it uploaded (none, pr, or results), whether its name is a /rerun
# run's, whether report.sh started commented, and text the comment must have
# afterwards, or "none" for no comment.
RERUN_REPORT_CASES=(
  "success|results|rerun|started|**No tests newly fail.**"
  "cancelled|results|rerun|started|**Cancelled in [this workflow run](https://github.com/o/r/actions/runs/7)**"
  "failure|none|rerun|started|**Failed in [this workflow run](https://github.com/o/r/actions/runs/7) before running the tests.**"
  "success|none|other|started|**Running in [this workflow run](https://github.com/o/r/actions/runs/7).**"
  "failure|none|rerun|not-started|none"
)

# A gh for report.sh that keeps pull request #4's comments in a file.
write_fake_gh() {
  local path="${1}"
  cat >"${path}" <<'EOF'
#!/usr/bin/env bash
comments="${FAKE_DIR}/comments.json"
case "${*}" in
  *"--method POST"*)
    jq -c --slurpfile new /dev/stdin '. + [{id: 2, user: {login: "github-actions[bot]"}, body: $new[0].body}]' \
      "${comments}" >"${comments}.new" ;;
  *"--method PATCH"*)
    jq -c --slurpfile new /dev/stdin 'map(.body = $new[0].body)' "${comments}" >"${comments}.new" ;;
  *"issues/comments/"*)
    echo '{"body": "/rerun connect-py https://github.com/connectrpc/connect-py/pull/463", "url": "https://github.com/o/r/pull/4#issuecomment-1"}'
    exit 0 ;;
  *"/issues/4/comments"*)
    jq "${@: -1}" "${comments}"
    exit 0 ;;
  *)
    exit 0 ;;
esac
mv "${comments}.new" "${comments}"
EOF
  chmod +x "${path}"
}

# Prints the comment on the pull request after reporting a /rerun run, or
# "none".
rerun_report() {
  local conclusion="${1}"
  local artifacts="${2}"
  local name="${3}"
  local started="${4}"
  local dir
  local run_name="Implementation Conformance"
  dir="$(mktemp -d)"
  mkdir -p "${dir}/bin" "${dir}/artifacts"
  write_fake_gh "${dir}/bin/gh"
  echo "[]" >"${dir}/comments.json"
  if [ "${name}" = "rerun" ]; then
    run_name="Implementation conformance for #4 (/rerun comment 1)"
  fi
  if [ "${artifacts}" != "none" ]; then
    mkdir -p "${dir}/artifacts/impl-conformance-pr"
    echo 4 >"${dir}/artifacts/impl-conformance-pr/pr-number"
  fi
  if [ "${artifacts}" = "results" ]; then
    mkdir -p "${dir}/artifacts/impl-conformance-changes" "${dir}/artifacts/impl-conformance-results-connect-py-client"
    echo '{"scope": "all", "suites": [], "deleted": [], "skipped": []}' \
      >"${dir}/artifacts/impl-conformance-changes/changes.json"
    echo '{"implementation": "connect-py", "target": "client", "commit": "abc1234", "status": "ok"}' \
      >"${dir}/artifacts/impl-conformance-results-connect-py-client/result.json"
  fi
  (
    export PATH="${dir}/bin:${PATH}" FAKE_DIR="${dir}" GH_TOKEN=token GITHUB_REPOSITORY=o/r RUN_ID=7 \
      RUN_URL=https://github.com/o/r/actions/runs/7 HEAD_SHA=0000000000000000000000000000000000000000 \
      EVENT=issue_comment RUN_NAME="${run_name}" CONCLUSION="${conclusion}" PR_NUMBER=4 COMMENT_ID=1
    if [ "${started}" = "started" ]; then
      ./report.sh started
    fi
    ./report.sh post "${dir}/artifacts" 2>/dev/null
  )
  jq -r 'if length == 0 then "none" else .[-1].body end' "${dir}/comments.json"
  rm -rf "${dir}"
}

failures=0
for test_case in "${CASES[@]}"; do
  IFS="|" read -r name head base head_retry base_retry expected <<<"${test_case}"
  check "${name}" "$(classify "${head}" "${base}" "${head_retry}" "${base_retry}")" "${expected}"
done
for test_case in "${COMMAND_CASES[@]}"; do
  IFS="|" read -r body expected <<<"${test_case}"
  check "command ${body}" \
    "$(jq -nc -L . --arg body "$(printf '%b' "${body}")" \
      'include "rerun"; $body | parse_command(["connect-go", "connect-py"])')" \
    "${expected}"
done
for test_case in "${RERUN_REPORT_CASES[@]}"; do
  IFS="|" read -r conclusion artifacts name started expected <<<"${test_case}"
  comment="$(rerun_report "${conclusion}" "${artifacts}" "${name}" "${started}")"
  if [ "${expected}" = "none" ]; then
    check "rerun report ${test_case}" "${comment}" "none"
  elif [[ "${comment}" != *"${expected}"* ]]; then
    check "rerun report ${test_case}" "${comment}" "a comment with ${expected}"
  fi
done
for mode in summary comment status; do
  if ! jq -n --arg mode "${mode}" --argjson results '[]' --argjson changes '[]' --argjson failedJobs '[]' \
    --arg sha "" --arg runUrl "" --arg reference "" -f report.jq >/dev/null; then
    echo "FAIL: report.jq in ${mode} mode" >&2
    failures=$((failures + 1))
  fi
done
total=$((${#CASES[@]} + ${#COMMAND_CASES[@]} + ${#RERUN_REPORT_CASES[@]} + 3))
if [ "${failures}" -gt 0 ]; then
  exit 1
fi
echo "All ${total} cases passed."
