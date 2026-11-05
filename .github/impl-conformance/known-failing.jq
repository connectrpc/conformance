# Marks each test in a runner report with whether the implementation's
# known-failing list matches it, using the runner's rules: patterns and names
# are split at "/", "*" matches one component, and "**" matches any number.
#
# Inputs: . (a runner report) and $patterns (the list's text: one pattern per
# line, ignoring blank lines and lines starting with "#").

# Like the runner, only a single trailing "**" matches once the name runs out.
def matches($pattern; $name):
  if ($name | length) == 0 then $pattern == [] or $pattern == ["**"]
  elif ($pattern | length) == 0 then false
  elif $pattern[0] == "**" then any(range(0; ($name | length) + 1); matches($pattern[1:]; $name[.:]))
  elif $pattern[0] == "*" or $pattern[0] == $name[0] then matches($pattern[1:]; $name[1:])
  else false
  end;

($patterns | split("\n") | map(gsub("^\\s+|\\s+$"; "") | select(length > 0 and (startswith("#") | not)) | split("/"))) as $parsed
| (.tests // []) | map(. + {knownFailing: ((.name | split("/")) as $name | any($parsed[]; matches(.; $name)))})
