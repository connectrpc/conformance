# Compares one target's test results with and without a pull request's
# changes, including the retries of the tests that differed. Base test names
# are renamed to the suites' head names first, so a renamed suite's tests
# compare with themselves.
#
# Inputs: $head, $base, $headRetry, and $baseRetry (slurped arrays of runner
# report tests, with head tests marked by known-failing.jq), $changes (slurped
# changes.json), $retry (none, done, too-many, or failed), $notApplicable (run
# patterns that applied to none of the head run's runner calls),
# $knownFailingList (whether the target passed a known-failing list), and the
# target's $implementation, $target, $commit, $status, $statusSide, and $notes.

include "outcomes";

# Drops the permutation components that follow the suite name, such as
# "Protocol:PROTOCOL_GRPC" and "(grpc server impl)", so each case is listed once.
def collapse:
  split("/") as $parts
  | [$parts[0]] + ($parts[1:] | until(length == 0 or (.[0] | test("^[A-Za-z]+:|^\\(.*\\)$") | not); .[1:]))
  | join("/");

# $listed maps the names of the permutations the known-failing list matches.
def group_cases($listed):
  group_by(collapse)
  | map({case: (.[0] | collapse), permutations: length, listed: map(select($listed[.])) | length});

def name_set: map({key: ., value: true}) | from_entries;

# A change is flaky if a retry in which every call ran the test doesn't repeat
# it. Tests that couldn't run at head compare by their retry instead.
def classify($headOutcomes; $baseOutcomes; $headRetryOutcomes; $baseRetryOutcomes):
  ($headRetryOutcomes[.] // []) as $headRetry
  | ($baseRetryOutcomes[.] // []) as $baseRetry
  | ($headOutcomes[.] | if includes("not-run") and ($headRetry | length) > 0 then $headRetry else . end) as $h
  | (compare($h; $baseOutcomes[.] // []) // "") as $change
  | if $change != "" then
      if ($headRetry | length) == 0 or ($headRetry + $baseRetry | includes("not-run"))
        or (compare($headRetry; $baseRetry) // "") == $change
      then $change else "flaky-\($change)" end
    elif $h | all(. == "not-run") then "not-run"
    else empty
    end;

base_renames($changes) as $renames
| ($head[0] + $headRetry[0]) as $headTests
| ($head[0] | outcomes) as $headOutcomes
| ($headRetry[0] | outcomes) as $headRetryOutcomes
| ($base[0] | map(.name |= rename($renames)) | outcomes) as $baseOutcomes
| ($baseRetry[0] | map(.name |= rename($renames)) | outcomes) as $baseRetryOutcomes
| ($headOutcomes | keys | map({name: ., kind: classify($headOutcomes; $baseOutcomes; $headRetryOutcomes; $baseRetryOutcomes)})
  | group_by(.kind) | map({key: .[0].kind, value: map(.name)}) | from_entries) as $kinds
# A failure is covered only if every run that failed it has it in its list. The
# runner fails a listed test that passes, so one listing run is enough there.
| ($headTests | map(select(.outcome == "failed")) | group_by(.name)
  | map(select(all(.knownFailing == true)) | .[0].name) | name_set) as $failuresListed
| ($headTests | map(select(.outcome == "passed" and .knownFailing == true) | .name) | name_set) as $passesListed
| ($kinds.failure // [] | group_cases($failuresListed)) as $newFailureCases
| ($kinds.pass // [] | group_cases($passesListed)) as $newPassCases
| ($kinds["flaky-failure"] // [] | group_cases($failuresListed) | map(. + {change: "failure"})) as $flakyFailureCases
| ($kinds["flaky-pass"] // [] | group_cases($passesListed) | map(. + {change: "pass"})) as $flakyPassCases
| ($kinds["not-run"] // [] | group_cases({})) as $notRunCases
| {
    implementation: $implementation,
    target: $target,
    commit: $commit,
    status: $status,
    statusSide: $statusSide,
    notes: $notes,
    notApplicable: ($notApplicable | split("\n") | map(select(. != "") | sub("/\\*\\*$"; ""))),
    knownFailingList: $knownFailingList,
    retry: $retry,
    newFailureCount: ($newFailureCases | length),
    newPassCount: ($newPassCases | length),
    # Test cases whose new failures the list doesn't all match, and test cases
    # with new passes it matches. Either way, the list has to change.
    newFailureUnlistedCount: ($newFailureCases | map(select(.listed < .permutations)) | length),
    newPassListedCount: ($newPassCases | map(select(.listed > 0)) | length),
    flakyCount: ($flakyFailureCases + $flakyPassCases | map(.case) | unique | length),
    notRunCount: ($notRunCases | length),
    newFailures: $newFailureCases,
    newPasses: $newPassCases,
    flaky: ($flakyFailureCases + $flakyPassCases | sort_by(.case)),
    notRun: $notRunCases
  }
