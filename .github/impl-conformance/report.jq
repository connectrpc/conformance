# Renders implementation conformance results as Markdown. Results come from a
# pull request's own workflow run, so they are untrusted: free text is shown
# only in sanitized code spans, numbers are clamped, and codes are mapped from
# fixed lists.
#
# With $mode "summary", . is one target's result and the output is its job
# summary. With $mode "comment", the inputs are $results (all targets),
# $changes (changes.json, or empty if unavailable), $failedJobs, $sha, $runUrl,
# and $reference (the reference implementation), and the output is an object
# with the comment body, check title, and check conclusion.

def code:
  tostring | gsub("[\\p{Cc}\\p{Cf}]"; "") | .[0:120]
  | gsub("`"; "'") | gsub("\\|"; "\\|") | "`" + . + "`";

def count: if type == "number" and . > 0 then floor else 0 end;

def plural($word): "\(.) \($word)\(if . == 1 then "" else "s" end)";

def list: if type == "array" then . else [] end;

def has_problem: (.status | tostring) as $status | $status != "ok" and $status != "not-applicable";

def status_text:
  {
    "runner-not-called": "didn't call the runner",
    "invalid-list": "known-flaky or skip list matches no test",
    "base-without-report-json": "the base branch's runner predates `--report-json`, so there's nothing to compare with",
    "error": "runner error"
  }[tostring] // "unknown status";

def side_text: {"base": " in the base run"}[tostring] // "";

def note_text:
  {
    "pinned-protos": "its generated code is from its pinned conformance release, but this pull request changes `proto/`"
  }[tostring] // empty;

def retry_text:
  {
    "too-many": "too many test cases changed to retry them",
    "failed": "retrying the changed test cases failed"
  }[tostring] // empty;

def listed_text:
  if .listed == 0 then "no"
  elif .listed >= .permutations then "yes"
  else "\(.listed) of \(.permutations)"
  end;

def change_text: {"failure": "new failure", "pass": "new pass"}[.change | tostring] // "";

# Lists each result's test cases under $key, at most $maxRows of them, with
# the extra column that $column names: "listed", "change", or none.
def case_table($key; $column; $maxRows):
  [.[] | . as $result | (.[$key] | list)[]
    | {target: $result.target, case: .case, permutations: (.permutations | count), listed: (.listed | count), change: .change}] as $rows
  | {"listed": "In known-failing list", "change": "Didn't repeat"}[$column] as $title
  | if ($rows | length) == 0 then empty else
      "| Target | Test case | Permutations |\(if $title then " \($title) |" else "" end)",
      "|---|---|---|\(if $title then "---|" else "" end)",
      ($rows[:$maxRows][] | "| \(.target | code) | \(.case | code) | \(.permutations) |\(
        if $column == "listed" then " \(listed_text) |" elif $column == "change" then " \(change_text) |" else "" end)"),
      (if ($rows | length) > $maxRows
        then "", "Only \($maxRows) of \($rows | length) are shown. The job summaries list all of them."
        else empty end)
    end;

def summary_table($title; $key; $column):
  [[.] | case_table($key; $column; 500)] as $lines
  | if ($lines | length) == 0 then empty else "", $title, "", $lines[] end;

def list_changes($add; $remove):
  if $add == 0 and $remove == 0 then "no changes"
  else [(if $add > 0 then "add \($add)" else empty end), (if $remove > 0 then "remove \($remove)" else empty end)] | join(", ")
  end;

def summary:
  "#### \(.implementation | code) \(.target | code) at \(.commit | code)",
  "",
  (if has_problem then "**Not tested: \(.status | status_text)\(.statusSide | side_text).**"
    else "Test cases with this pull request's changes: \(.newFailureCount | count) newly failing, \(.newPassCount | count) newly passing." end),
  (if has_problem then empty
    elif .knownFailingList == true then "", "Known-failing list: \(list_changes(.newFailureUnlistedCount | count; .newPassListedCount | count))."
    else "", "This target passes no known-failing list to the runner." end),
  (if (.notApplicable | list | length) > 0
    then "", "Not applicable: \(.notApplicable | list | map(code) | join(", "))."
    else empty end),
  ((.retry | retry_text) as $text | "", "Flaky test cases couldn't be found: \($text)."),
  (if (.flakyCount | count) > 0
    then "", "Possibly flaky: \(.flakyCount | count | plural("test case")) had a result that didn't repeat when retried."
    else empty end),
  (if (.notRunCount | count) > 0
    then "", "**\(.notRunCount | count | plural("test case")) couldn't run, because the program under test timed out or exited.**"
    else empty end),
  summary_table("Newly failing test cases:"; "newFailures"; "listed"),
  summary_table("Newly passing test cases:"; "newPasses"; "listed"),
  summary_table("Possibly flaky test cases:"; "flaky"; "change"),
  summary_table("Test cases that couldn't run:"; "notRun"; ""),
  "";

def implementations($results):
  $results | group_by(.implementation | tostring) | map({
    name: (.[0].implementation | tostring),
    commits: (map(.commit | tostring) | unique),
    targets: length,
    newFailures: (map(.newFailureCount | count) | add),
    failingTargets: (map(select((.newFailureCount | count) > 0)) | length),
    newPasses: (map(.newPassCount | count) | add),
    flaky: (map(.flakyCount | count) | add),
    notRun: (map(.notRunCount | count) | add),
    hasList: any(.[]; .knownFailingList == true),
    toAdd: (map(.newFailureUnlistedCount | count) | add),
    toRemove: (map(.newPassListedCount | count) | add),
    problems: [.[] | (select(has_problem) | "\(.target | code): \(.status | status_text)\(.statusSide | side_text)"),
      (select((.notRunCount | count) > 0) | "\(.target | code): \(.notRunCount | count | plural("test case")) couldn't run")],
    notes: ([.[] | .notes | list | .[] | note_text] | unique)
      + [.[] | (.retry | retry_text) as $text | "\(.target | code): \($text)"]
      + (map(.flakyCount | count) | add | if . > 0 then [plural("possibly flaky test case")] else [] end),
    results: .
  });

# $key names both an implementation's count and its results' test case lists.
def details($implementations; $key; $column; $title; $note; $maxRows):
  [$implementations[] | select(.[$key] > 0)] as $selected
  | if ($selected | length) == 0 then empty else
      "<details><summary>\($title)</summary>",
      "",
      $note,
      "",
      ($selected[] | "##### \(.name | code)", "", (.results | case_table($key; $column; $maxRows)), ""),
      "</details>",
      ""
    end;

def comment:
  $ARGS.named as $args
  | $args.results as $results
  | ($args.changes[0] // null) as $changes
  | $args.failedJobs as $failedJobs
  | $args.sha as $sha
  | $args.runUrl as $runUrl
  | $args.reference as $reference
  | implementations($results) as $implementations
  | ($changes.skipped? | list | map(tostring)) as $skipped
  | any($skipped[]; . == $reference) as $referenceSkipped
  | ([$implementations[] | select(.newFailures > 0)] | length) as $failingImplementations
  | ([$implementations[] | select(.name == $reference and .newFailures > 0)] | length > 0) as $referenceFails
  | (([$implementations[].problems[]] | length) + ($failedJobs | length)) as $problems
  | (if $changes == null then "Couldn't determine which suites changed."
      elif $referenceFails then "\($reference), the reference implementation, has newly failing tests. They block merging."
      elif $failingImplementations > 0 then "\($failingImplementations | plural("implementation")) \(if $failingImplementations == 1 then "has" else "have" end) newly failing tests."
      elif ($results | length) == 0 and ($skipped | length) > 0 then "Every implementation was skipped by label."
      elif ($results | length) == 0 then "No added or modified suites to run."
      else "No tests newly fail."
      end
    + (if $problems > 0 then " \($problems | plural("problem")) running tests." else "" end)) as $headline
  | [
      "### Implementation conformance",
      "",
      "**\($headline)**",
      "",
      (if $referenceSkipped
        then "**\($reference), the reference implementation, was skipped by label, so this pull request's effect on it is unknown.**", ""
        else empty end),
      (if ($implementations | length) == 0 then empty else
        "| Implementation | Newly failing test cases | Newly passing test cases | Known-failing list | Notes |",
        "|---|---|---|---|---|",
        ($implementations[]
          | "| \(.name | code) \(.commits | map(code) | join(" "))\(if .name == $reference then " (reference)" else "" end)"
            + " | \(if .newFailures > 0 then "**\(.newFailures)** in \(.failingTargets) of \(.targets) targets" else "0" end)"
            + " | \(.newPasses)"
            + " | \(if .hasList then list_changes(.toAdd; .toRemove) else "no list" end)"
            + " | \(.problems + .notes | join("; ")) |"),
        ""
      end),
      (if ($failedJobs | length) > 0 then "Failed jobs: \($failedJobs | map(code) | join(", ")).", "" else empty end),
      (if ($skipped | length) > 0 then "Skipped by label: \($skipped | .[:20] | map(code) | join(", ")).", "" else empty end),
      (if $changes == null then empty
        elif $changes.scope == "all" then
          "This pull request changes code or protos\(if ($results | length) > 0 then ", so every suite ran" else "" end).", ""
        elif ($changes.suites | list | length) > 0 then
          "Changed suites: \($changes.suites | list | .[:20] | map(.name | code) | join(", "))"
          + (if ($changes.suites | list | length) > 20 then ", and \(($changes.suites | length) - 20) more." else "." end),
          ""
        else empty end),
      (if ($changes.deleted? | list | length) > 0 then
        "Deleted suites: \($changes.deleted | .[:20] | map(code) | join(", ")). Implementations must remove known-failing and known-flaky entries for their tests when they update.",
        ""
        else empty end),
      (if ($results | length) > 0
        then "Each implementation ran the suites with and without this pull request's changes, without known-failing lists, at the commit shown. Test cases whose results differed ran once more on each side, and those whose results didn't repeat are listed as possibly flaky instead of newly failing or passing. The new failures and passes were then checked against the known-failing list each target passes to the runner. "
        else "" end)
      + "Results for \($sha[0:12] | code) from [this workflow run](\($runUrl)).",
      ""
    ] as $header
  | [
      details($implementations; "newFailures"; "listed"; "Newly failing test cases";
        "Test cases that aren't in the known-failing list need an entry there, or a fix."; 50),
      details($implementations; "newPasses"; "listed"; "Newly passing test cases";
        "The runner fails known-failing test cases that pass, so the list has to stop matching the ones it lists."; 20),
      details($implementations; "flaky"; "change"; "Possibly flaky test cases";
        "These test cases newly failed or passed once, but not when they ran again."; 20),
      details($implementations; "notRun"; ""; "Test cases that couldn't run";
        "The program under test timed out or exited before these test cases ran."; 20)
    ] as $details
  | ($header + $details | join("\n")) as $body
  | {
      title: $headline,
      conclusion: (if $failingImplementations == 0 and $problems == 0 and $changes != null and ($referenceSkipped | not)
        then "success" else "neutral" end),
      body: (if ($body | length) < 60000 then $body
        else $header + ["The details are too long for a comment. The job summaries in the workflow run list them."] | join("\n") end)
    };

if $ARGS.named.mode == "summary" then summary else comment end
