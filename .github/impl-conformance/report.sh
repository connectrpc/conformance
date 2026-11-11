#!/usr/bin/env bash

# Renders implementation conformance results and reports them on the pull
# request that produced them.
#
# Results come from workflows that run a pull request's code, possibly from a
# fork, so every input is untrusted. report.jq sanitizes what it renders, and
# the pull request number is used only if that pull request's head is the
# tested commit.
#
# Usage:
#   report.sh summary <result.json>
#     Prints one target's result as Markdown, for the job summary.
#   report.sh post <artifacts dir>
#     Creates or updates the results comment on the pull request and creates
#     a check run on the tested commit. For a run started by a /rerun comment
#     (EVENT is issue_comment), posts a new comment instead. Requires GH_TOKEN,
#     GITHUB_REPOSITORY, HEAD_SHA (the tested commit, or the default branch's
#     for a comment run), RUN_ID, RUN_URL, and, for a comment run, EVENT and
#     RUN_NAME (the run's display title).

set -euo pipefail

SCRIPT_DIR="$(CDPATH= cd "$(dirname "${0}")" && pwd)"
CHECK_NAME="Implementation conformance results"
COMMENT_MARKER="<!-- impl-conformance-report -->"
RERUN_MARKER="<!-- impl-conformance-rerun -->"
# Keep in sync with IMPLEMENTATIONS in impl-conformance.yaml.
IMPLEMENTATIONS=(connect-go connect-es connect-kotlin connect-swift connect-dart connect-py connect-rust)
# It builds the reference programs from this repository, so its new failures
# predict failures in its own CI.
REFERENCE_IMPLEMENTATION="connect-go"

find_comment() {
  local pr_number="${1}"
  gh api --paginate "repos/${GITHUB_REPOSITORY}/issues/${pr_number}/comments" \
    --jq ".[] | select(.user.login == \"github-actions[bot]\" and (.body | startswith(\"${COMMENT_MARKER}\"))) | .id" |
    head -n 1
}

# Prints the pull request number from the artifact, but only if that pull
# request's head is the tested commit, so a fork can't target other PRs.
# Returns 1 when there's nothing to report and 2 when the number is invalid.
verified_pr_number() {
  local pr_file="${1}"
  local pr_number
  local pr_head_sha
  if [ ! -f "${pr_file}" ]; then
    echo "::notice::No pull request number was uploaded; nothing to report." >&2
    return 1
  fi
  pr_number="$(head -c 20 "${pr_file}" | tr -d '[:space:]')"
  if ! [[ "${pr_number}" =~ ^[0-9]+$ ]]; then
    echo "::error::Invalid pull request number in artifact." >&2
    return 2
  fi
  pr_head_sha="$(gh api "repos/${GITHUB_REPOSITORY}/pulls/${pr_number}" --jq '.head.sha')"
  if [ "${pr_head_sha}" != "${HEAD_SHA}" ]; then
    echo "::notice::Pull request #${pr_number} head is ${pr_head_sha}, not the tested commit ${HEAD_SHA}; skipping." >&2
    return 1
  fi
  echo "${pr_number}"
}

# Prints the /rerun command of a comment run as JSON, with its pull request
# number, comment URL, and the implementation and pull request it names. The
# pull request and comment come from the run's name, which impl-conformance.yaml
# sets from the event, so pull request code can't change them, and the command
# comes from the comment itself. Returns 1 if the name doesn't match.
rerun_command() {
  local pattern='^Implementation conformance for #([0-9]{1,9}) \(/rerun comment ([0-9]{1,20})\)$'
  local pr_number
  local comment_id
  local comment
  if ! [[ "${RUN_NAME:-}" =~ ${pattern} ]]; then
    echo "::error::Run name of comment run doesn't match." >&2
    return 1
  fi
  pr_number="${BASH_REMATCH[1]}"
  comment_id="${BASH_REMATCH[2]}"
  comment="$(gh api "repos/${GITHUB_REPOSITORY}/issues/comments/${comment_id}" --jq '{body, url: .html_url}' ||
    echo '{"body": "", "url": ""}')"
  jq -c -L "${SCRIPT_DIR}" --argjson prNumber "${pr_number}" \
    'include "rerun"; {prNumber: $prNumber, commentUrl: .url} + (.body | parse_command($ARGS.positional) // {})' \
    --args "${IMPLEMENTATIONS[@]}" <<<"${comment}"
}

# Creates the results comment, or updates it if one exists. With "update-only",
# never creates one.
upsert_comment() {
  local pr_number="${1}"
  local body_file="${2}"
  local mode="${3:-}"
  local comment_id
  comment_id="$(find_comment "${pr_number}")"
  if [ -n "${comment_id}" ]; then
    jq -n --rawfile body "${body_file}" '{body: $body}' |
      gh api --method PATCH "repos/${GITHUB_REPOSITORY}/issues/comments/${comment_id}" --input - >/dev/null
  elif [ "${mode}" != "update-only" ]; then
    jq -n --rawfile body "${body_file}" '{body: $body}' |
      gh api --method POST "repos/${GITHUB_REPOSITORY}/issues/${pr_number}/comments" --input - >/dev/null
  fi
}

# Prints the JSON objects in the given files, one per line, skipping files
# that aren't JSON objects. jq reads the output with --slurpfile, since results
# can exceed the size limit for a command-line argument.
json_objects() {
  local json_file
  for json_file in "${@}"; do
    if jq -e 'type == "object"' "${json_file}" >/dev/null 2>&1; then
      jq -c '.' "${json_file}"
    else
      echo "::warning::Skipping invalid results file ${json_file}." >&2
    fi
  done
}

post_report() {
  local artifacts_dir="${1}"
  : "${GH_TOKEN:?}" "${GITHUB_REPOSITORY:?}" "${HEAD_SHA:?}" "${RUN_ID:?}" "${RUN_URL:?}"
  local pr_number
  local pr_status=0
  local result_files=()
  local data_dir
  local body_file
  local rerun="null"
  local sha="${HEAD_SHA}"

  if [ "${EVENT:-}" = "issue_comment" ]; then
    # Only runs for maintainers' /rerun comments get past the prepare job's
    # condition, and it uploads this first.
    if [ ! -f "${artifacts_dir}/impl-conformance-pr/pr-number" ]; then
      echo "::notice::The comment wasn't a /rerun command; nothing to report." >&2
      return
    fi
    rerun="$(rerun_command)"
    pr_number="$(jq -r '.prNumber' <<<"${rerun}")"
    # The tested commit is the pull request's head when the run started. It
    # comes from the run's artifacts, so it's only used if it looks like one.
    sha="$(head -c 100 "${artifacts_dir}/impl-conformance-pr/head-sha" 2>/dev/null | tr -d '[:space:]' || true)"
    if ! [[ "${sha}" =~ ^[0-9a-f]{40}$ ]]; then
      sha="unknown"
    fi
  else
    pr_number="$(verified_pr_number "${artifacts_dir}/impl-conformance-pr/pr-number")" || pr_status="${?}"
    case "${pr_status}" in
      0) ;;
      1) return ;;
      *) exit 1 ;;
    esac
  fi

  mapfile -t result_files < <(find "${artifacts_dir}" -type f -path '*/impl-conformance-results-*/*.json' | sort)
  data_dir="$(mktemp -d)"
  json_objects "${result_files[@]}" >"${data_dir}/results.json"
  json_objects "${artifacts_dir}/impl-conformance-changes/changes.json" >"${data_dir}/changes.json"
  gh api --paginate "repos/${GITHUB_REPOSITORY}/actions/runs/${RUN_ID}/jobs" \
    --jq '.jobs[] | select(.conclusion == "failure" or .conclusion == "cancelled" or .conclusion == "timed_out") | .name' |
    jq -R 'select(. != "")' >"${data_dir}/failed-jobs.json"

  body_file="$(mktemp)"
  if [ "${rerun}" != "null" ]; then
    echo "${RERUN_MARKER}" >"${body_file}"
    jq -n \
      --arg mode comment \
      --slurpfile results "${data_dir}/results.json" \
      --slurpfile changes "${data_dir}/changes.json" \
      --slurpfile failedJobs "${data_dir}/failed-jobs.json" \
      --arg sha "${sha}" \
      --arg runUrl "${RUN_URL}" \
      --arg reference "${REFERENCE_IMPLEMENTATION}" \
      --argjson rerun "${rerun}" \
      -f "${SCRIPT_DIR}/report.jq" | jq -r '.body' >>"${body_file}"
    jq -n --rawfile body "${body_file}" '{body: $body}' |
      gh api --method POST "repos/${GITHUB_REPOSITORY}/issues/${pr_number}/comments" --input - >/dev/null
    return
  fi
  echo "${COMMENT_MARKER}" >"${body_file}"
  if jq -n -e --slurpfile results "${data_dir}/results.json" --slurpfile failedJobs "${data_dir}/failed-jobs.json" \
    --slurpfile changes "${data_dir}/changes.json" \
    '$results == [] and $failedJobs == [] and ($changes | length) == 1
    and $changes[0].scope == "changed" and $changes[0].suites == [] and $changes[0].deleted == []' >/dev/null; then
    {
      echo "### Implementation conformance"
      echo
      echo "No test suites were added, modified, or deleted."
    } >>"${body_file}"
    upsert_comment "${pr_number}" "${body_file}" update-only
    return
  fi

  jq -n \
    --arg mode comment \
    --slurpfile results "${data_dir}/results.json" \
    --slurpfile changes "${data_dir}/changes.json" \
    --slurpfile failedJobs "${data_dir}/failed-jobs.json" \
    --arg sha "${HEAD_SHA}" \
    --arg runUrl "${RUN_URL}" \
    --arg reference "${REFERENCE_IMPLEMENTATION}" \
    -f "${SCRIPT_DIR}/report.jq" >"${data_dir}/report.json"
  jq -r '.body' "${data_dir}/report.json" >>"${body_file}"
  upsert_comment "${pr_number}" "${body_file}"

  jq -n \
    --arg name "${CHECK_NAME}" \
    --arg head_sha "${HEAD_SHA}" \
    --arg details_url "${RUN_URL}" \
    --slurpfile report "${data_dir}/report.json" \
    --rawfile summary "${body_file}" \
    '{name: $name, head_sha: $head_sha, status: "completed", conclusion: $report[0].conclusion,
      details_url: $details_url, output: {title: $report[0].title, summary: $summary}}' |
    gh api --method POST "repos/${GITHUB_REPOSITORY}/check-runs" --input - >/dev/null
}

case "${1:-}" in
  summary)
    shift
    jq -r --arg mode summary -f "${SCRIPT_DIR}/report.jq" "${@}"
    ;;
  post)
    shift
    post_report "${@}"
    ;;
  *)
    echo "usage: ${0} summary <result.json> | post <artifacts dir>" >&2
    exit 2
    ;;
esac
