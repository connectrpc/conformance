# Definitions shared by retry.jq and new-failures.jq, which include this
# module with -L pointing at this directory.

# Renames a test's suite with $renames, which maps base suite names to head
# names, so a renamed suite's tests compare with themselves.
def rename($renames):
  split("/") as $parts
  | [$renames[$parts[0]] // $parts[0]] + $parts[1:]
  | join("/");

def base_renames($changes): $changes[0].suites | map(select(.baseName != "") | {key: .baseName, value: .name}) | from_entries;

# Maps each test name in an array of runner report tests to its outcomes. A
# target can call the runner more than once, so a test can have several.
def outcomes: reduce .[] as $test ({}; .[$test.name] += [$test.outcome]);

def includes($outcome): any(.[]; . == $outcome);

def tally($outcome): map(select(. == $outcome)) | length;

# Prints "failure" or "pass" if a test's outcomes changed from $base to $head.
# Counts matter because calls like connect-kotlin's call styles share names.
def compare($head; $base):
  if ($head | tally("failed")) > ($base | tally("failed")) then "failure"
  elif ($base | tally("failed")) > ($head | tally("failed")) and ($head | tally("passed")) > ($base | tally("passed")) then "pass"
  else empty
  end;
