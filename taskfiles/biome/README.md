# Biome Taskfile Public Tasks

## What is this Taskfile?

This Taskfile wraps Biome for formatting, linting, combined checks, and CI. It
installs `@biomejs/biome` locally and delegates package-manager behavior to the
shared `js-pm` helper.

## Public Tasks

| Task           | Variables                                                    | Description                                                                |
| -------------- | ------------------------------------------------------------ | -------------------------------------------------------------------------- |
| `install`      | Optional `PM`, `EXTRA_ARGS`, `CLI_ARGS`                      | Install `@biomejs/biome` as a local dev dependency.                        |
| `init`         | Optional `PM`, `EXTRA_ARGS`, `CLI_ARGS`                      | Alias for `config:init`.                                                   |
| `config:init`  | Optional `PM`, `EXTRA_ARGS`, `CLI_ARGS`                      | Run `biome init`. Skipped if `biome.json` or `biome.jsonc` already exists. |
| `check`        | Optional `PM`, `TARGETS`, `CONFIG`, `EXTRA_ARGS`, `CLI_ARGS` | Run `biome check`.                                                         |
| `check:write`  | Optional `PM`, `TARGETS`, `CONFIG`, `EXTRA_ARGS`, `CLI_ARGS` | Run `biome check --write`.                                                 |
| `fix`          | Optional `PM`, `TARGETS`, `CONFIG`, `EXTRA_ARGS`, `CLI_ARGS` | Alias for `check:write`.                                                   |
| `lint`         | Optional `PM`, `TARGETS`, `CONFIG`, `EXTRA_ARGS`, `CLI_ARGS` | Run `biome lint`.                                                          |
| `lint:fix`     | Optional `PM`, `TARGETS`, `CONFIG`, `EXTRA_ARGS`, `CLI_ARGS` | Run `biome lint --write`.                                                  |
| `format`       | Optional `PM`, `TARGETS`, `CONFIG`, `EXTRA_ARGS`, `CLI_ARGS` | Run `biome format`.                                                        |
| `format:write` | Optional `PM`, `TARGETS`, `CONFIG`, `EXTRA_ARGS`, `CLI_ARGS` | Run `biome format --write`.                                                |
| `ci`           | Optional `PM`, `TARGETS`, `CONFIG`, `EXTRA_ARGS`, `CLI_ARGS` | Run `biome ci`.                                                            |
| `cache:clean`  | —                                                            | Remove common Biome cache directories.                                     |
| `version`      | Optional `PM`                                                | Show the resolved Biome version.                                           |
| `help`         | Optional `PM`, `EXTRA_ARGS`, `CLI_ARGS`                      | Show Biome CLI help.                                                       |

## Variables

`PM` defaults to lockfile detection through `js-pm`. `TARGETS` defaults to `.`,
and `CONFIG` adds `--config-path <path>`.

`EXTRA_ARGS` and arguments after `--` are appended to the command.

## Examples

```bash
task biome:install
task biome:config:init
task biome:check
task biome:check:write
task biome:lint
task biome:format:write
task biome:ci PM=pnpm CONFIG=biome.json
```
