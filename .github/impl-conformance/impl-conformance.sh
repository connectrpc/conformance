#!/usr/bin/env bash

# Runs a Connect implementation's conformance tests against the test suites
# changed in a pull request, with and without the pull request's changes, to
# find the tests that newly fail.
#
# Usage:
#   impl-conformance.sh changes <base commit> <output file>
#     Run in the conformance repository. Writes the changes between <base
#     commit> and HEAD as JSON: the suites to run, with their head and base
#     names; the deleted suites; whether every suite must run because code or
#     protos changed; and whether protos changed.
#   impl-conformance.sh install <path> [<version>]
#     Installs a connectconformance wrapper at <path>. Implementations call it
#     in place of the runner version they pin. With <version>, the wrapper's
#     --version prints that version instead of the runner's.
#   impl-conformance.sh install-go <path> <wrapper>
#     Installs a go command at <path> that runs <wrapper> for
#     "go run connectrpc.com/conformance/cmd/connectconformance@<version>" and
#     the go command on PATH for anything else.
#   impl-conformance.sh install-program <path> <base program> <head program>
#     Installs a wrapper at <path> that runs the program matching the runner in
#     use, for test programs built from the conformance repository.
#   impl-conformance.sh run <implementation> <target> <command...>
#     Runs <command> with the pull request's runner and suites, then with the
#     base commit's, without known-failing lists, then runs the tests whose
#     results differ again on each side to find flaky ones. Writes the tests
#     that newly fail or newly pass, and whether the implementation's
#     known-failing list matches them, the flaky tests, and the tests that
#     couldn't run to RESULTS_DIR/result.json and the job summary. Fails only
#     if <command> never called the wrapper, since then nothing ran.
#   impl-conformance.sh skipped <labels> <implementation...>
#     Prints the given implementations that the pull request's labels skip, as
#     a JSON array. <labels> is a JSON array of the label names.
#   impl-conformance.sh exec <runner args...>
#     Called by the wrapper. Runs the current runner limited to the selected
#     suites that apply to the call, without --known-failing, and writes a JSON
#     report. Saves the --known-failing patterns next to the report. Always
#     exits 0.
#
# Unlike most scripts, this does not cd to the project root, because <command>
# must run in the implementation's own working directory.

set -euo pipefail

: "${CONFORMANCE_DIR:=${RUNNER_TEMP:-}/conformance}"
: "${RESULTS_DIR:=${RUNNER_TEMP:-}/results}"
: "${GITHUB_STEP_SUMMARY:=/dev/stdout}"
# Set to "true" when the implementation's generated code comes from its pinned
# conformance release, so it can't reflect proto changes.
: "${PROTOS_PINNED:=false}"
# More newly failing or passing tests than this suggest a real regression. Tests
# that couldn't run don't count, since a stalled client leaves many of them.
: "${MAX_RETRIES:=500}"

SCRIPT_DIR="$(CDPATH= cd "$(dirname "${0}")" && pwd)"
SCRIPT_PATH="${SCRIPT_DIR}/$(basename "${0}")"
SUITES_PATH="internal/app/connectconformance/testsuites/data"
# Changes here can affect the outcome of any test, so every suite runs.
CODE_PATHS=(cmd internal proto ":(exclude)${SUITES_PATH}")
CHANGES_FILE="${CONFORMANCE_DIR}/changes.json"
PATTERNS_FILE="${CONFORMANCE_DIR}/run-patterns.txt"
# Holds "base" or "head", selecting the runner and programs the wrappers run.
SIDE_FILE="${CONFORMANCE_DIR}/side"
# Each wrapper call appends the path of its report, to prove the
# implementation used the wrapper.
CALLS_FILE="${CONFORMANCE_DIR}/wrapper-calls"
REPORTS_DIR="${CONFORMANCE_DIR}/reports"
# A pull request label that skips every implementation. Followed by a space and
# an implementation's name, it skips that one.
SKIP_LABEL="impl-conformance: skip"

# Prints a suite's name at a commit. Fails for names that can't be used as the
# first component of a test pattern.
suite_name() {
  local commit="${1}"
  local path="${2}"
  local name
  name="$(git show "${commit}:${path}" | yq '.name')"
  if [ -z "${name}" ] || [ "${name}" = "null" ] || [[ "${name}" =~ [/*[:cntrl:]] ]] ||
    [[ "${name}" =~ ^[[:space:]#] ]] || [[ "${name}" =~ [[:space:]]$ ]]; then
    echo "::error file=${path}::Suite name \"${name}\" at ${commit} can't be used in a test pattern." >&2
    return 1
  fi
  echo "${name}"
}

detect_changes() {
  local base_commit="${1}"
  local output_file="${2}"
  local diff_file
  local status
  local old_path
  local new_path
  local name
  local base_name
  local suites_json="[]"
  local deleted_json="[]"
  local scope="changed"
  local protos_changed="false"
  diff_file="$(mktemp)"
  git diff --name-status -M --diff-filter=ADMR "${base_commit}" HEAD -- "${SUITES_PATH}/*.yaml" >"${diff_file}"
  while IFS=$'\t' read -r status old_path new_path; do
    case "${status}" in
      A)
        name="$(suite_name HEAD "${old_path}")"
        base_name=""
        ;;
      M)
        name="$(suite_name HEAD "${old_path}")"
        base_name="$(suite_name "${base_commit}" "${old_path}")"
        ;;
      R*)
        name="$(suite_name HEAD "${new_path}")"
        base_name="$(suite_name "${base_commit}" "${old_path}")"
        ;;
      D)
        base_name="$(suite_name "${base_commit}" "${old_path}")"
        deleted_json="$(jq -c --arg name "${base_name}" '. + [$name]' <<<"${deleted_json}")"
        continue
        ;;
      *)
        continue
        ;;
    esac
    suites_json="$(jq -c --arg name "${name}" --arg baseName "${base_name}" \
      '. + [{name: $name, baseName: $baseName}]' <<<"${suites_json}")"
  done <"${diff_file}"
  # git diff exits 1 for changes; treating other failures as changes is safe.
  if ! git diff --quiet "${base_commit}" HEAD -- "${CODE_PATHS[@]}"; then
    scope="all"
  fi
  if ! git diff --quiet "${base_commit}" HEAD -- proto; then
    protos_changed="true"
  fi
  jq -n --arg scope "${scope}" --argjson suites "${suites_json}" \
    --argjson deleted "${deleted_json}" --argjson protosChanged "${protos_changed}" \
    '{scope: $scope, suites: $suites, deleted: $deleted, protosChanged: $protosChanged}' >"${output_file}"
  cat "${output_file}"
}

install_wrapper() {
  local wrapper_path="${1}"
  local reported_version="${2:-}"
  chmod +x "${CONFORMANCE_DIR}/base/connectconformance" "${CONFORMANCE_DIR}/head/connectconformance"
  echo "head" >"${SIDE_FILE}"
  touch "${PATTERNS_FILE}"
  mkdir -p "$(dirname "${wrapper_path}")" "${REPORTS_DIR}"
  # Paths are baked in because build tools like turbo can strip environment
  # variables before running the wrapper.
  cat >"${wrapper_path}" <<EOF
#!/usr/bin/env bash
CONFORMANCE_DIR="${CONFORMANCE_DIR}" REPORTED_VERSION="${reported_version}" exec "${SCRIPT_PATH}" exec "\${@}"
EOF
  chmod +x "${wrapper_path}"
}

install_go() {
  local go_path="${1}"
  local wrapper_path="${2}"
  local real_go
  real_go="$(command -v go)"
  if [ "${real_go}" = "${go_path}" ]; then
    echo "error: ${go_path} is already the first go command on PATH" >&2
    return 1
  fi
  mkdir -p "$(dirname "${go_path}")"
  cat >"${go_path}" <<EOF
#!/usr/bin/env bash
if [ "\${1:-}" = "run" ] && [[ "\${2:-}" == connectrpc.com/conformance/cmd/connectconformance@* ]]; then
  shift 2
  exec "${wrapper_path}" "\${@}"
fi
exec "${real_go}" "\${@}"
EOF
  chmod +x "${go_path}"
}

install_program() {
  local wrapper_path="${1}"
  local base_program="${2}"
  local head_program="${3}"
  mkdir -p "$(dirname "${wrapper_path}")"
  cat >"${wrapper_path}" <<EOF
#!/usr/bin/env bash
if [ "\$(cat "${SIDE_FILE}")" = "base" ]; then
  exec "${base_program}" "\${@}"
fi
exec "${head_program}" "\${@}"
EOF
  chmod +x "${wrapper_path}"
}

# Writes the patterns of --known-failing values to a file, reading the "@"
# files they name relative to the current directory, as the runner would.
save_known_failing() {
  local output_file="${1}"
  shift
  local value
  : >"${output_file}"
  for value in "${@}"; do
    if [ "${value#@}" = "${value}" ]; then
      echo "${value}" >>"${output_file}"
    elif ! { cat "${value#@}" && echo; } >>"${output_file}"; then
      echo "warning: couldn't read known-failing list ${value#@}" >&2
    fi
  done
}

# Starts the program under test with no input when a client-mode call didn't
# start it, because some harnesses, like connect-dart's, wait for it to connect.
start_idle_client() {
  local mode="${1}"
  shift
  if [ "${mode}" = "client" ] && [ "${1:-}" = "--" ] && [ "${#}" -gt 1 ]; then
    shift
    "${@}" </dev/null >/dev/null || true
  fi
}

exec_runner() {
  local side
  local runner_binary
  local report_file
  local patterns_file
  local unmatched_file
  local attempt
  local mode=""
  local args=()
  local known_failing=()
  side="$(cat "${SIDE_FILE}")"
  runner_binary="${CONFORMANCE_DIR}/${side}/connectconformance"
  # The wrapper selects tests and writes the report, so --run and --report-json
  # are dropped. --known-failing is saved to check results against, not passed.
  while [ "${#}" -gt 0 ] && [ "${1}" != "--" ]; do
    case "${1}" in
      --version)
        # connect-dart replaces a cached runner unless this output ends with
        # the version it pins.
        if [ -n "${REPORTED_VERSION:-}" ]; then
          printf 'connectconformance %s' "${REPORTED_VERSION}"
          exit 0
        fi
        exec "${runner_binary}" --version
        ;;
      --mode)
        mode="${2:-}"
        ;;
      --mode=*)
        mode="${1#--mode=}"
        ;;
      --known-failing)
        known_failing+=("${2}")
        shift 2
        continue
        ;;
      --known-failing=*)
        known_failing+=("${1#--known-failing=}")
        shift
        continue
        ;;
      --run | --report-json)
        shift 2
        continue
        ;;
      --run=* | --report-json=*)
        shift
        continue
        ;;
    esac
    args+=("${1}")
    shift
  done
  report_file="$(mktemp "${REPORTS_DIR}/${side}-XXXXXX")"
  patterns_file="${report_file}.patterns"
  unmatched_file="${report_file}.unmatched"
  cp "${PATTERNS_FILE}" "${patterns_file}"
  : >"${report_file}.not-applicable"
  if [ "${#known_failing[@]}" -gt 0 ]; then
    save_known_failing "${report_file}.known-failing" "${known_failing[@]}"
  fi
  echo "${report_file}" >>"${CALLS_FILE}"
  # The runner rejects run patterns that match none of this call's tests, so
  # suites that don't apply to its config or mode are dropped and it runs again.
  for attempt in 1 2; do
    if [ ! -s "${patterns_file}" ]; then
      echo '{"tests": []}' >"${report_file}"
      break
    fi
    "${runner_binary}" --run "@${patterns_file}" --report-json "${report_file}" "${args[@]}" "${@}" || true
    jq -r 'select(.unmatchedPatterns.kind? == "run") | .unmatchedPatterns.patterns[]' \
      "${report_file}" >"${unmatched_file}" 2>/dev/null || true
    if [ ! -s "${unmatched_file}" ]; then
      break
    fi
    cat "${unmatched_file}" >>"${report_file}.not-applicable"
    grep -vxF -f "${unmatched_file}" "${patterns_file}" >"${patterns_file}.remaining" || true
    mv "${patterns_file}.remaining" "${patterns_file}"
  done
  # A runner that started the client reports its tests, even ones that couldn't
  # run, so a report without tests means the client never started.
  if ! jq -e '(.tests | length) > 0' "${report_file}" >/dev/null 2>&1; then
    start_idle_client "${mode}" "${@}"
  fi
  # Results come from the reports. A failing exit status would stop
  # implementations that call the runner several times after the first failure.
  exit 0
}

# Reads the reports of the last command run and sets SIDE_STATUS to ok,
# not-applicable (no call ran a test), runner-not-called, invalid-list, or
# error. Writes the run patterns that applied to none of the calls to the given
# file.
classify_reports() {
  local not_applicable_file="${1}"
  local report_file
  local has_error="false"
  local has_invalid_list="false"
  local has_tests="false"
  local report_files=()
  : >"${not_applicable_file}"
  if [ ! -s "${CALLS_FILE}" ]; then
    SIDE_STATUS="runner-not-called"
    return
  fi
  mapfile -t report_files <"${CALLS_FILE}"
  for report_file in "${report_files[@]}"; do
    if ! jq -e '.tests | type == "array"' "${report_file}" >/dev/null 2>&1; then
      has_error="true"
      continue
    fi
    case "$(jq -r '.unmatchedPatterns.kind // ""' "${report_file}")" in
      "")
        if jq -e '(.error // "") != "" and (.tests | length) == 0' "${report_file}" >/dev/null; then
          has_error="true"
        fi
        ;;
      # The wrapper already dropped the run patterns the runner rejected.
      run)
        has_error="true"
        ;;
      *)
        has_invalid_list="true"
        ;;
    esac
    if jq -e '(.tests | length) > 0' "${report_file}" >/dev/null; then
      has_tests="true"
    fi
  done
  jq -Rnr --argjson calls "${#report_files[@]}" '[inputs] | group_by(.) | map(select(length == $calls))[][0]' \
    "${report_files[@]/%/.not-applicable}" >"${not_applicable_file}"
  if [ "${has_error}" = "true" ]; then
    SIDE_STATUS="error"
  elif [ "${has_invalid_list}" = "true" ]; then
    SIDE_STATUS="invalid-list"
  elif [ "${has_tests}" = "false" ]; then
    SIDE_STATUS="not-applicable"
  else
    SIDE_STATUS="ok"
  fi
}

# Runs the command with one side's runner. Sets SIDE_STATUS and
# SIDE_HAS_KNOWN_FAILING, and writes the side's test results, marked with
# whether its known-failing list matches them, and the run patterns that
# applied to none of its runner calls to the work directory, in files starting
# with <name>.
run_side() {
  local side="${1}"
  local name="${2}"
  local work_dir="${3}"
  local label="${4}"
  shift 4
  local report_file
  local known_failing_file
  local report_files=()
  echo "[]" >"${work_dir}/${name}-tests.json"
  echo "${side}" >"${SIDE_FILE}"
  SIDE_HAS_KNOWN_FAILING="false"
  rm -rf "${CALLS_FILE}" "${REPORTS_DIR}"
  mkdir -p "${REPORTS_DIR}"
  echo "::group::${label}: ${name} run"
  # Test failures are read from the reports, not the exit status.
  "${@}" </dev/null || true
  echo "::endgroup::"
  classify_reports "${work_dir}/${name}-not-applicable.txt"
  if [ "${SIDE_STATUS}" = "ok" ]; then
    mapfile -t report_files <"${CALLS_FILE}"
    for report_file in "${report_files[@]}"; do
      known_failing_file="${report_file}.known-failing"
      if [ -e "${known_failing_file}" ]; then
        SIDE_HAS_KNOWN_FAILING="true"
      else
        known_failing_file="/dev/null"
      fi
      jq --rawfile patterns "${known_failing_file}" -f "${SCRIPT_DIR}/known-failing.jq" "${report_file}"
    done >"${work_dir}/${name}-marked-tests.json"
    jq -s 'add // []' "${work_dir}/${name}-marked-tests.json" >"${work_dir}/${name}-tests.json"
  fi
}

# Runs the tests that newly failed, newly passed, or couldn't run once more on
# each side, to tell flaky tests from real changes. Sets RETRY_STATUS to none,
# done, too-many, or failed.
retry_changed() {
  local work_dir="${1}"
  local label="${2}"
  shift 2
  local count
  jq -n -L "${SCRIPT_DIR}" \
    --slurpfile head "${work_dir}/head-tests.json" \
    --slurpfile base "${work_dir}/base-tests.json" \
    --slurpfile changes "${CHANGES_FILE}" \
    -f "${SCRIPT_DIR}/retry.jq" >"${work_dir}/retry.json"
  count="$(jq '.head | length' "${work_dir}/retry.json")"
  if [ "${count}" -eq 0 ]; then
    RETRY_STATUS="none"
    return
  fi
  if [ "$(jq '.changed' "${work_dir}/retry.json")" -gt "${MAX_RETRIES}" ]; then
    RETRY_STATUS="too-many"
    return
  fi
  RETRY_STATUS="done"
  jq -r '.head[]' "${work_dir}/retry.json" >"${PATTERNS_FILE}"
  run_side head head-retry "${work_dir}" "${label}" "${@}"
  if [ "${SIDE_STATUS}" = "ok" ]; then
    jq -r '.base[]' "${work_dir}/retry.json" >"${PATTERNS_FILE}"
    if [ -s "${PATTERNS_FILE}" ]; then
      run_side base base-retry "${work_dir}" "${label}" "${@}"
    fi
  fi
  case "${SIDE_STATUS}" in
    ok | not-applicable) ;;
    *)
      RETRY_STATUS="failed"
      echo "[]" >"${work_dir}/head-retry-tests.json"
      echo "[]" >"${work_dir}/base-retry-tests.json"
      ;;
  esac
}

# Succeeds if the base runner has --report-json. A base branch from before the
# flag existed can't report results, so there's nothing to compare with.
base_writes_reports() {
  local help_text
  # Captured, not piped to grep -q, whose early exit would fail the pipe.
  help_text="$("${CONFORMANCE_DIR}/base/connectconformance" --help 2>&1 || true)"
  [[ "${help_text}" == *--report-json* ]]
}

skipped_implementations() {
  local labels="${1}"
  shift
  jq -cn --argjson labels "${labels}" --arg skip "${SKIP_LABEL}" \
    '$ARGS.positional | map(select(. as $name | any($labels[]; . == $skip or . == $skip + " " + $name)))' \
    --args "${@}"
}

run_target() {
  local implementation="${1}"
  local target="${2}"
  shift 2
  local label="${implementation} ${target}"
  local work_dir
  local commit
  local scope
  local status
  local status_side=""
  local has_known_failing="false"
  local notes="[]"
  work_dir="$(mktemp -d)"
  commit="$(git rev-parse --short HEAD 2>/dev/null || echo "unknown")"
  scope="$(jq -r '.scope' "${CHANGES_FILE}")"
  if [ "${PROTOS_PINNED}" = "true" ] && jq -e '.protosChanged' "${CHANGES_FILE}" >/dev/null; then
    notes='["pinned-protos"]'
  fi
  mkdir -p "${RESULTS_DIR}"
  echo "[]" >"${work_dir}/head-tests.json"
  : >"${work_dir}/head-not-applicable.txt"
  echo "[]" >"${work_dir}/base-tests.json"
  echo "[]" >"${work_dir}/head-retry-tests.json"
  echo "[]" >"${work_dir}/base-retry-tests.json"
  RETRY_STATUS="none"

  if [ "${scope}" = "all" ]; then
    echo "**" >"${PATTERNS_FILE}"
  else
    jq -r '.suites[].name + "/**"' "${CHANGES_FILE}" >"${PATTERNS_FILE}"
  fi
  if ! base_writes_reports; then
    status="base-without-report-json"
  else
    run_side head head "${work_dir}" "${label}" "${@}"
    status="${SIDE_STATUS}"
    has_known_failing="${SIDE_HAS_KNOWN_FAILING}"
    if [ "${status}" != "ok" ]; then
      status_side="head"
    fi
  fi
  if [ "${status}" = "ok" ]; then
    # Only suites that apply at head run at base, under their base names.
    if [ "${scope}" = "all" ]; then
      echo "**" >"${PATTERNS_FILE}"
    else
      jq -r --rawfile notApplicable "${work_dir}/head-not-applicable.txt" \
        '($notApplicable | split("\n")) as $skip
        | .suites[] | select(.baseName != "" and ((.name + "/**") | IN($skip[]) | not))
        | .baseName + "/**"' "${CHANGES_FILE}" >"${PATTERNS_FILE}"
    fi
    if [ -s "${PATTERNS_FILE}" ]; then
      run_side base base "${work_dir}" "${label}" "${@}"
      # A base without the suites has no results.
      case "${SIDE_STATUS}" in
        ok | not-applicable) ;;
        *)
          status="${SIDE_STATUS}"
          status_side="base"
          ;;
      esac
    fi
    if [ "${status}" = "ok" ]; then
      retry_changed "${work_dir}" "${label}" "${@}"
    fi
  fi

  jq -n -L "${SCRIPT_DIR}" \
    --arg implementation "${implementation}" \
    --arg target "${target}" \
    --arg commit "${commit}" \
    --arg status "${status}" \
    --arg statusSide "${status_side}" \
    --argjson notes "${notes}" \
    --argjson knownFailingList "${has_known_failing}" \
    --arg retry "${RETRY_STATUS}" \
    --rawfile notApplicable "${work_dir}/head-not-applicable.txt" \
    --slurpfile head "${work_dir}/head-tests.json" \
    --slurpfile base "${work_dir}/base-tests.json" \
    --slurpfile headRetry "${work_dir}/head-retry-tests.json" \
    --slurpfile baseRetry "${work_dir}/base-retry-tests.json" \
    --slurpfile changes "${CHANGES_FILE}" \
    -f "${SCRIPT_DIR}/new-failures.jq" >"${RESULTS_DIR}/result.json"
  "${SCRIPT_DIR}/report.sh" summary "${RESULTS_DIR}/result.json" >>"${GITHUB_STEP_SUMMARY}"
  if [ "${status}" = "runner-not-called" ]; then
    echo "::error title=Conformance runner not called::${label} didn't call the connectconformance wrapper during its ${status_side} run" >&2
    exit 1
  fi
}

case "${1:-}" in
  changes)
    shift
    detect_changes "${@}"
    ;;
  install)
    shift
    install_wrapper "${@}"
    ;;
  install-go)
    shift
    install_go "${@}"
    ;;
  install-program)
    shift
    install_program "${@}"
    ;;
  run)
    shift
    run_target "${@}"
    ;;
  skipped)
    shift
    skipped_implementations "${@}"
    ;;
  exec)
    shift
    exec_runner "${@}"
    ;;
  *)
    echo "usage: ${0} changes <base commit> <output file> | install <path> [<version>] | install-go <path> <wrapper> | install-program <path> <base> <head> | run <implementation> <target> <command...> | skipped <labels> <implementation...> | exec <args...>" >&2
    exit 2
    ;;
esac
