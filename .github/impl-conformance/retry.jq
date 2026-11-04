# Lists the tests to run again: those that newly failed or newly passed, and
# those that couldn't run at head. Prints {head: [names], base: [names],
# changed: <count>}, with base names for the base run; tests in suites added at
# head aren't run there. changed counts the tests that newly failed or passed.
#
# Inputs: $head and $base (slurped arrays of runner report tests) and $changes
# (slurped changes.json).

include "outcomes";

base_renames($changes) as $renames
| ($renames | to_entries | map({key: .value, value: .key}) | from_entries) as $baseNames
| ($changes[0].suites | map(select(.baseName == "") | .name)) as $added
| ($head[0] | outcomes) as $headOutcomes
| ($base[0] | map(.name |= rename($renames)) | outcomes) as $baseOutcomes
| [$headOutcomes | to_entries[]
    | ($baseOutcomes[.key] // []) as $baseResults
    | select(
        (.value | includes("failed")) and ($baseResults | includes("failed") | not)
        or (.value | includes("passed")) and ($baseResults | includes("failed"))
      )
    | .key] as $changed
| [$headOutcomes | to_entries[] | select(.value | includes("not-run")) | .key] as $notRun
| ($changed + $notRun | unique) as $names
| {
    changed: ($changed | length),
    head: $names,
    base: [$names[] | split("/") as $parts | select(($parts[0] | IN($added[])) | not)
      | [$baseNames[$parts[0]] // $parts[0]] + $parts[1:] | join("/")]
  }
