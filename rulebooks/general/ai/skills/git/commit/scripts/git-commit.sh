#!/usr/bin/env bash
set -euo pipefail
IFS=$'\n\t'

readonly MAX_MESSAGE_LENGTH=120
readonly ALLOWED_TYPES=(feat fix refactor docs test chore perf build ci)

commit_type=""
commit_message=""
commit_scope=""

help() {
  cat <<'USAGE'
Usage:
  git-commit.sh <type> <message> [--commit-scope=<scope>]

Arguments:
  type            Conventional Commit type (required)
  message         Commit message (required, max 120 characters)

Options:
  --commit-scope  Optional scope (lowercase words separated by hyphens)
  -h, --help      Show this help

Examples:
  git-commit.sh feat "Add checkout flow" --commit-scope=checkout
  git-commit.sh docs "Update API usage" 
USAGE
}

fail() {
  printf 'Error: %s\n' "$1" >&2
  exit 1
}

validate_message() {
  local message="$1"

  if (( ${#message} > MAX_MESSAGE_LENGTH )); then
    fail "Commit message must be at most ${MAX_MESSAGE_LENGTH} characters."
  fi
}

validate_type() {
  local type="$1"
  local allowed=false

  for allowed_type in "${ALLOWED_TYPES[@]}"; do
    if [[ "$type" == "$allowed_type" ]]; then
      allowed=true
      break
    fi
  done

  if [[ "$allowed" != "true" ]]; then
    fail "Invalid commit type: ${type}."
  fi
}

validate_scope() {
  local scope="$1"

  if [[ -z "$scope" ]]; then
    return 0
  fi

  if [[ ! "$scope" =~ ^[a-z]+(-[a-z]+)*$ ]]; then
    fail "Invalid commit scope: ${scope}."
  fi
}

commit() {
  local type="$1"
  local message="$2"
  local scope="$3"
  local subject

  if [[ -n "$scope" ]]; then
    subject="${type}(${scope}): ${message}"
  else
    subject="${type}: ${message}"
  fi

  git commit -m "$subject"
}

main() {
  while [[ $# -gt 0 ]]; do
    case "$1" in
      -h|--help)
        help
        return 0
        ;;
      --commit-scope=*)
        commit_scope="${1#*=}"
        shift
        ;;
      --commit-scope)
        commit_scope="${2-}"
        shift 2
        ;;
      --*)
        fail "Unknown option: $1"
        ;;
      *)
        if [[ -z "$commit_type" ]]; then
          commit_type="$1"
        elif [[ -z "$commit_message" ]]; then
          commit_message="$1"
        else
          fail "Unexpected argument: $1"
        fi
        shift
        ;;
    esac
  done

  if [[ -z "$commit_type" || -z "$commit_message" ]]; then
    help
    fail "Missing required arguments."
  fi

  validate_message "$commit_message"
  validate_type "$commit_type"
  validate_scope "$commit_scope"
  commit "$commit_type" "$commit_message" "$commit_scope"
}

main "$@"
