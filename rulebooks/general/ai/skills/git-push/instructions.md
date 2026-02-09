# Git Push

## Purpose
Use this skill to push local commits to a remote repository.

## When to Use
- The user requests to push commits.
- The local branch is ready to be published.

## Inputs
- remote (optional): remote name (default: `origin`).
- branch (optional): branch name (default: current branch).
- force-with-lease (optional): add `--force-with-lease`.

## Command
Run:
`./scripts/git-push.sh [--remote=<name>] [--branch=<name>] [--force-with-lease]`

## Behavior
- Ensures the current directory is a git repository.
- Resolves the current branch when `--branch` is not provided.
- Fails on detached HEAD unless a branch is specified.
- Validates that the remote exists.
- Uses `git push --set-upstream` on first push when no upstream is set.
- Uses regular `git push` when an upstream already exists.

## Notes
- Use `--force-with-lease` instead of `--force`.
- Use `-h` or `--help` to show usage.
