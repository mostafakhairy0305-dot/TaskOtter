# Knip Taskfile Public Tasks

## What is this Taskfile?

This Taskfile wraps Knip for unused file, export, and dependency analysis. Knip
can report framework-specific false positives, so treat output as review input
instead of an instruction to delete files or packages automatically.

## Public Tasks

| Task | Variables | Description |
|---|---|---|
| `install` | Optional `PM`, `EXTRA_ARGS`, `CLI_ARGS` | Install `knip` as a local dev dependency. |
| `init` | Optional `PM`, `EXTRA_ARGS`, `CLI_ARGS` | Initialize Knip configuration. |
| `config:init` | Optional `PM`, `EXTRA_ARGS`, `CLI_ARGS` | Alias for `init`. |
| `check` | Optional `PM`, `CONFIG`, `EXTRA_ARGS`, `CLI_ARGS` | Run the default Knip analysis. |
| `production` | Optional `PM`, `CONFIG`, `EXTRA_ARGS`, `CLI_ARGS` | Run Knip with `--production`. |
| `dependencies` | Optional `PM`, `CONFIG`, `EXTRA_ARGS`, `CLI_ARGS` | Report unused production dependencies. |
| `dev-dependencies` | Optional `PM`, `CONFIG`, `EXTRA_ARGS`, `CLI_ARGS` | Report unused development dependencies. |
| `files` | Optional `PM`, `CONFIG`, `EXTRA_ARGS`, `CLI_ARGS` | Report unused files. |
| `exports` | Optional `PM`, `CONFIG`, `EXTRA_ARGS`, `CLI_ARGS` | Report unused exports. |
| `fix` | Optional `PM`, `CONFIG`, `EXTRA_ARGS`, `CLI_ARGS` | Run `knip --fix` when supported by the installed version. |
| `ci` | Optional `PM`, `CONFIG`, `EXTRA_ARGS`, `CLI_ARGS` | Run production checks for CI. |
| `version` | Optional `PM` | Show the resolved Knip version. |
| `help` | Optional `PM`, `EXTRA_ARGS`, `CLI_ARGS` | Show Knip CLI help. |

## Variables

`PM` defaults to lockfile detection through `js-pm`. `CONFIG` adds
`--config <path>`. `EXTRA_ARGS` and arguments after `--` are appended to the
command.

Review Knip findings before deleting files or dependencies.

## Examples

```bash
task knip:install
task knip:check
task knip:production
task knip:dependencies
task knip:files
task knip:exports
task knip:check PM=pnpm -- --debug
```
