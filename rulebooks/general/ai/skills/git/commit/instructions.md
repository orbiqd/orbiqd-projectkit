# Git Commit

## Purpose
Use this skill to create a Conventional Commit message and execute a git commit.

## When to Use
- The user requests a Conventional Commit.
- Staged changes are ready to be committed.

## Inputs
- type (required): one of `feat`, `fix`, `refactor`, `docs`, `test`, `chore`, `perf`, `build`, `ci`.
- message (required): commit message, max 120 characters.
- scope (optional): lowercase words separated by hyphens (e.g., `api`, `cli-parser`).

## Command
Run:
`./scripts/git-commit.sh <type> "<message>" [--commit-scope=<scope>]`

## Behavior
- Validates message length (<= 120).
- Validates type against the allowed list.
- Validates scope with pattern `^[a-z]+(-[a-z]+)*$` when provided.
- Builds the subject as `type(scope): message` or `type: message`.
- Executes `git commit -m "<subject>"`.

## Notes
- Quote the message if it contains spaces.
- The script does not stage files. Stage changes beforehand when needed.
- Use `-h` or `--help` to show usage.
