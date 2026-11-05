# Definitions shared by retry.jq and new-failures.jq, which include this
# module with -L pointing at this directory.

# Renames a test's suite with $renames, which maps base suite names to head
# names, so a renamed suite's tests compare with themselves.
def rename($renames):
  split("/") as $parts
  | [$renames[$parts[0]] // $parts[0]] + $parts[1:]
  | join("/");

def base_renames($changes): $changes[0].suites | map(select(.baseName != "") | {key: .baseName, value: .name}) | from_entries;

# Whether the runs that called the runner made the same number of calls, so
# call N of one lines up with call N of another. $calls maps runs to counts.
def aligned($calls): [$calls[] | select(. > 0)] | unique | length <= 1;

# Maps each test in an array of runner report tests to its outcomes, keyed by
# its call index and name as JSON. Unless $perCall, every call shares index 0.
def outcomes($perCall):
  reduce .[] as $test ({}; .[[if $perCall then $test.call else 0 end, $test.name] | tojson] += [$test.outcome]);

def key_name: fromjson[1];

def includes($outcome): any(.[]; . == $outcome);

def tally($outcome): map(select(. == $outcome)) | length;

# Prints "failure" or "pass" if a test's outcomes changed from $base to $head.
# Counts matter when calls that don't line up, like call styles, share names.
def compare($head; $base):
  if ($head | tally("failed")) > ($base | tally("failed")) then "failure"
  elif ($base | tally("failed")) > ($head | tally("failed")) and ($head | tally("passed")) > ($base | tally("passed")) then "pass"
  else empty
  end;
