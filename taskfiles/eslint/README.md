# ESLint Taskfile Public Tasks

## What is this Taskfile?

This Taskfile wraps ESLint for JavaScript and TypeScript projects. It installs
ESLint as a local dev dependency, runs cached checks by default, supports strict
CI mode, and keeps package-manager selection consistent with the shared
`js-pm` helper.

## Public Tasks

| Task | Variables | Description |
|---|---|---|
| `install` | Optional `PM`, `EXTRA_ARGS`, `CLI_ARGS` | Install `eslint` as a local dev dependency. |
| `init` | Optional `PM`, `EXTRA_ARGS`, `CLI_ARGS` | Run the ESLint configuration wizard. |
| `config:init` | Optional `PM`, `EXTRA_ARGS`, `CLI_ARGS` | Alias for `init`. |
| `check` | Optional `PM`, `TARGETS`, `CONFIG`, `CACHE`, `EXTRA_ARGS`, `CLI_ARGS` | Lint targets with cache enabled by default. |
| `fix` | Optional `PM`, `TARGETS`, `CONFIG`, `CACHE`, `EXTRA_ARGS`, `CLI_ARGS` | Run ESLint with `--fix`. |
| `ci` | Optional `PM`, `TARGETS`, `CONFIG`, `CACHE`, `EXTRA_ARGS`, `CLI_ARGS` | Run ESLint with `--max-warnings=0`. |
| `cache:clean` | — | Remove `.cache/eslint`. |
| `version` | Optional `PM` | Show the resolved ESLint version. |
| `help` | Optional `PM`, `EXTRA_ARGS`, `CLI_ARGS` | Show ESLint CLI help. |

## Variables

`PM` defaults to lockfile detection through `js-pm`. Override it with
`PM=npm`, `PM=pnpm`, `PM=yarn`, or `PM=bun`.

`TARGETS` defaults to `src/**/*.{js,jsx,ts,tsx}`. `CONFIG` adds
`--config <path>`. `CACHE` defaults to `true`; set `CACHE=false` to omit cache
flags. `EXTRA_ARGS` and arguments after `--` are appended to the command.

## Examples

```bash
task eslint:install
task eslint:check
task eslint:fix TARGETS="src test"
task eslint:ci PM=pnpm
task eslint:check PM=pnpm -- --quiet
```
