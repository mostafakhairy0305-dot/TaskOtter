# Prettier Taskfile Public Tasks

## What is this Taskfile?

This Taskfile wraps Prettier checks and writes for JavaScript/TypeScript
projects and workspaces. It uses the shared `js-pm` helper for local binary
execution and package-manager detection.

## Public Tasks

| Task | Variables | Description |
|---|---|---|
| `install` | Optional `PM`, `EXTRA_ARGS`, `CLI_ARGS` | Install `prettier` as a local dev dependency. |
| `config:init` | Optional `CONFIG` | Create a starter Prettier config when one does not exist. |
| `check` | Optional `PM`, `TARGETS`, `CONFIG`, `IGNORE_PATH`, `EXTRA_ARGS`, `CLI_ARGS` | Run `prettier --check`. |
| `write` | Optional `PM`, `TARGETS`, `CONFIG`, `IGNORE_PATH`, `EXTRA_ARGS`, `CLI_ARGS` | Run `prettier --write`. |
| `fix` | Optional `PM`, `TARGETS`, `CONFIG`, `IGNORE_PATH`, `EXTRA_ARGS`, `CLI_ARGS` | Alias for `write`. |
| `ci` | Optional `PM`, `TARGETS`, `CONFIG`, `IGNORE_PATH`, `EXTRA_ARGS`, `CLI_ARGS` | Alias for `check`. |
| `version` | Optional `PM` | Show the resolved Prettier version. |
| `help` | Optional `PM`, `EXTRA_ARGS`, `CLI_ARGS` | Show Prettier CLI help. |

## Variables

`PM` defaults to lockfile detection through `js-pm`. `TARGETS` defaults to `.`,
`CONFIG` adds `--config <path>`, and `IGNORE_PATH` defaults to
`.prettierignore`. The ignore path is only passed when the file exists.

`EXTRA_ARGS` and arguments after `--` are appended to the command.

## Examples

```bash
task prettier:install
task prettier:check
task prettier:write TARGETS="src/**/*.ts"
task prettier:fix PM=bun -- --ignore-unknown
```
