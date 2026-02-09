#!/usr/bin/env bash
set -euo pipefail
IFS=$'\n\t'

remote="origin"
branch=""
force_with_lease=false

help() {
  cat <<'USAGE'
Usage:
  git-push.sh [--remote=<name>] [--branch=<name>] [--force-with-lease]

Options:
  --remote           Remote name (default: origin)
  --branch           Branch name (default: current branch)
  --force-with-lease Use --force-with-lease
  -h, --help         Show this help

Examples:
  git-push.sh
  git-push.sh --branch=feat/login
USAGE
}

fail() {
  printf 'Error: %s\n' "$1" >&2
  exit 1
}

ensure_repo() {
  git rev-parse --is-inside-work-tree >/dev/null 2>&1 || fail "Not a git repository."
}

resolve_branch() {
  local current
  current="$(git rev-parse --abbrev-ref HEAD)"
  if [[ "$current" == "HEAD" ]]; then
    fail "Detached HEAD. Specify --branch explicitly."
  fi
  branch="${branch:-$current}"
}

validate_remote() {
  git remote get-url "$remote" >/dev/null 2>&1 || fail "Unknown remote: ${remote}."
}

has_upstream() {
  git rev-parse --abbrev-ref --symbolic-full-name "@{u}" >/dev/null 2>&1
}

push() {
  local args=()
  if [[ "$force_with_lease" == "true" ]]; then
    args+=(--force-with-lease)
  fi

  if has_upstream; then
    git push "${args[@]}" "$remote" "$branch"
    return
  fi

  git push --set-upstream "${args[@]}" "$remote" "$branch"
}

main() {
  while [[ $# -gt 0 ]]; do
    case "$1" in
      -h|--help)
        help
        return 0
        ;;
      --remote=*)
        remote="${1#*=}"
        shift
        ;;
      --remote)
        remote="${2-}"
        shift 2
        ;;
      --branch=*)
        branch="${1#*=}"
        shift
        ;;
      --branch)
        branch="${2-}"
        shift 2
        ;;
      --force-with-lease)
        force_with_lease=true
        shift
        ;;
      --*)
        fail "Unknown option: $1"
        ;;
      *)
        fail "Unexpected argument: $1"
        ;;
    esac
  done

  ensure_repo
  resolve_branch
  validate_remote
  push
}

main "$@"
