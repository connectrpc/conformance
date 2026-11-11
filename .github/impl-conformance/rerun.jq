# Parses /rerun comments, which are untrusted, so only bodies that match the
# command exactly are used.
#
# A command is a comment whose first line is
#   /rerun <implementation> [https://github.com/connectrpc/<implementation>/pull/<number>]
# It tests the implementation at the pull request's head, or at its default
# branch without one.

# Parses a comment body. Produces null if it isn't a command, or an object with
# implementation and pullRequest (a number, or null), or with error.
def parse_command($implementations):
  (tostring | split("\n")[0] | gsub("^\\s+|\\s+$"; "")) as $line
  | if ($line | test("^/rerun(\\s|$)") | not) then null
    else ($line | [splits("\\s+")]) as $words
    | ($words[1] // "") as $implementation
    | ($words[2] // "") as $url
    | if ($words | length) < 2 or ($words | length) > 3 then
        {error: "usage"}
      elif ($implementation | IN($implementations[]) | not) then
        {error: "unknown-implementation"}
      elif $url == "" then
        {implementation: $implementation, pullRequest: null}
      else
        ($url | capture("^https://github\\.com/connectrpc/(?<repo>[A-Za-z0-9_.-]+)/pull/(?<number>[0-9]{1,9})/?$")
          // null) as $match
        | if $match == null or $match.repo != $implementation then
            {error: "invalid-pull-request"}
          else
            {implementation: $implementation, pullRequest: ($match.number | tonumber)}
          end
      end
    end;
