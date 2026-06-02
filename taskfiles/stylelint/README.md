# Stylelint Taskfile Public Tasks

## What is this Taskfile?

This Taskfile wraps Stylelint for stylesheet checks and fixes. It installs
`stylelint` and `stylelint-config-standard` locally, enables cache by default,
and uses `js-pm` for package-manager detection and binary execution.

## Public Tasks

| Task          | Variables                                                                                  | Description                                                          |
| ------------- | ------------------------------------------------------------------------------------------ | -------------------------------------------------------------------- |
| `install`     | Optional `PM`, `EXTRA_ARGS`, `CLI_ARGS`                                                    | Install Stylelint and the standard config as local dev dependencies. |
| `config:init` | Optional `CONFIG`                                                                          | Create a starter Stylelint config when one does not exist.           |
| `check`       | Optional `PM`, `TARGETS`, `CONFIG`, `CACHE`, `ALLOW_EMPTY_INPUT`, `EXTRA_ARGS`, `CLI_ARGS` | Lint stylesheet targets.                                             |
| `fix`         | Optional `PM`, `TARGETS`, `CONFIG`, `CACHE`, `ALLOW_EMPTY_INPUT`, `EXTRA_ARGS`, `CLI_ARGS` | Run Stylelint with `--fix`.                                          |
| `ci`          | Optional `PM`, `TARGETS`, `CONFIG`, `CACHE`, `ALLOW_EMPTY_INPUT`, `EXTRA_ARGS`, `CLI_ARGS` | Run Stylelint with `--max-warnings=0`.                               |
| `cache:clean` | —                                                                                          | Remove `.cache/stylelint`.                                           |
| `version`     | Optional `PM`                                                                              | Show the resolved Stylelint version.                                 |
| `help`        | Optional `PM`, `EXTRA_ARGS`, `CLI_ARGS`                                                    | Show Stylelint CLI help.                                             |

## Variables

`PM` defaults to lockfile detection through `js-pm`. `TARGETS` defaults to
`**/*.{css,scss,sass,less,vue,svelte,astro}`. `CONFIG` adds `--config <path>`.

`CACHE` and `ALLOW_EMPTY_INPUT` both default to `true`; set either to `false`
to omit that flag. `EXTRA_ARGS` and arguments after `--` are appended to the
command.

## Examples

```bash
task stylelint:install
task stylelint:check
task stylelint:fix TARGETS="src/**/*.scss"
task stylelint:ci PM=yarn
task stylelint:fix PM=yarn -- --formatter verbose
```
