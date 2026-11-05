#!/usr/bin/env bash

# Tests how new-failures.jq classifies one test's outcomes, given one outcome
# per runner call in each of the head, base, and retry runs.

set -euo pipefail

DIR="$(CDPATH= cd "$(dirname "${0}")" && pwd)"
cd "${DIR}"

# name|head|base|head retry|base retry|expected kind (failure, pass,
# flaky-failure, flaky-pass, not-run, or none)
CASES=(
  "new failure|failed|passed|failed|passed|failure"
  "new failure without a retry|failed|passed|||failure"
  "new failure in an added suite|failed||failed||failure"
  "failure that passes when retried|failed|passed|passed|passed|flaky-failure"
  "failure that also fails at base when retried|failed|passed|failed|failed|flaky-failure"
  "failure whose retry couldn't run|failed|passed|not-run|passed|failure"
  "new failure in one of several calls|passed passed failed|passed passed passed|passed passed failed|passed passed passed|failure"
  "failure in one of several calls that doesn't repeat|passed passed failed|passed passed passed|passed passed passed|passed passed passed|flaky-failure"
  "failure in one call at head and another at base|failed passed|passed failed|||none"
  "new pass|passed|failed|passed|failed|pass"
  "pass that fails at base only once|passed|failed|passed|passed|flaky-pass"
  "new pass in one of several calls|passed passed|passed failed|passed passed|passed failed|pass"
  "couldn't run at head|not-run|passed|||not-run"
  "couldn't run at head and failed when retried|not-run|passed|failed|passed|failure"
  "couldn't run at head and passed when retried|not-run|passed|passed|passed|none"
)

# Prints a runner report's tests for a test with the given outcomes.
tests_json() {
  local outcomes="${1}"
  jq -cn --arg outcomes "${outcomes}" \
    '$outcomes | split(" ") | map(select(. != "") | {name: "Suite/test", outcome: ., knownFailing: false})'
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
    -f new-failures.jq |
    jq -r '[
      (.newFailures[] | "failure"), (.newPasses[] | "pass"),
      (.flaky[] | "flaky-" + .change), (.notRun[] | "not-run")
    ] | if length == 0 then "none" else join(",") end'
}

failures=0
for test_case in "${CASES[@]}"; do
  IFS="|" read -r name head base head_retry base_retry expected <<<"${test_case}"
  actual="$(classify "${head}" "${base}" "${head_retry}" "${base_retry}")"
  if [ "${actual}" != "${expected}" ]; then
    echo "FAIL: ${name}: got ${actual}, want ${expected}" >&2
    failures=$((failures + 1))
  fi
done
if [ "${failures}" -gt 0 ]; then
  exit 1
fi
echo "All ${#CASES[@]} cases passed."
