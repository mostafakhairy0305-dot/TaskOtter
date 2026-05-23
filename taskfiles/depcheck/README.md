# Depcheck Taskfile Public Tasks

## What is this Taskfile?

This Taskfile wraps Depcheck for unused and missing dependency reports. It runs
against the project root by default and uses the shared `js-pm` helper for local
binary execution.

## Public Tasks

| Task | Variables | Description |
|---|---|---|
| `install` | Optional `PM`, `EXTRA_ARGS`, `CLI_ARGS` | Install `depcheck` as a local dev dependency. |
| `check` | Optional `PM`, `PROJECT_PATH`, `TARGETS`, `EXTRA_ARGS`, `CLI_ARGS` | Run Depcheck. |
| `json` | Optional `PM`, `PROJECT_PATH`, `TARGETS`, `EXTRA_ARGS`, `CLI_ARGS` | Run Depcheck with `--json`. |
| `ignores` | Optional `PM`, `PROJECT_PATH`, `TARGETS`, `IGNORE_PACKAGES`, `EXTRA_ARGS`, `CLI_ARGS` | Run Depcheck with ignored packages. |
| `skip-missing` | Optional `PM`, `PROJECT_PATH`, `TARGETS`, `EXTRA_ARGS`, `CLI_ARGS` | Run Depcheck with `--skip-missing=true`. |
| `ci` | Optional `PM`, `PROJECT_PATH`, `TARGETS`, `EXTRA_ARGS`, `CLI_ARGS` | Run Depcheck and fail on findings. |
| `version` | Optional `PM` | Show the resolved Depcheck version. |
| `help` | Optional `PM`, `EXTRA_ARGS`, `CLI_ARGS` | Show Depcheck CLI help. |

## Variables

`PM` defaults to lockfile detection through `js-pm`. `PROJECT_PATH` defaults to
`.` and can be overridden for monorepo packages. `TARGETS` is accepted as an
alias for the project path when used from aggregate tasks.

`IGNORE_PACKAGES` is a comma-separated list for the `ignores` task. `EXTRA_ARGS`
and arguments after `--` are appended to the command.

## Examples

```bash
task depcheck:install
task depcheck:check
task depcheck:json
task depcheck:check PROJECT_PATH=packages/app
task depcheck:check -- --ignores="@types/*,eslint-*"
```
