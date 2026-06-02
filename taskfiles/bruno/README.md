# Bruno Taskfile Public Tasks

## What is this Taskfile?

This Taskfile wraps the [Bruno](https://www.usebruno.com/) CLI (`@usebruno/cli`)
for running API collections from the command line. It installs Bruno as a local
dev dependency and keeps package-manager selection consistent with the shared
`js-pm` helper.

## Public Tasks

| Task      | Variables                                                    | Description                                              |
| --------- | ------------------------------------------------------------ | -------------------------------------------------------- |
| `install` | Optional `PM`, `EXTRA_ARGS`, `CLI_ARGS`                      | Install `@usebruno/cli` as a local dev dependency.       |
| `run`     | Optional `PM`, `COLLECTION`, `ENV`, `EXTRA_ARGS`, `CLI_ARGS` | Run all requests in the Bruno collection.                |
| `ci`      | Optional `PM`, `COLLECTION`, `ENV`, `EXTRA_ARGS`, `CLI_ARGS` | Run collection and stop on the first failure (`--bail`). |
| `version` | Optional `PM`                                                | Show the locally resolved `bru` version.                 |
| `help`    | Optional `PM`, `EXTRA_ARGS`, `CLI_ARGS`                      | Show Bruno CLI help.                                     |

## Variables

`PM` defaults to lockfile detection through `js-pm`. Override it with
`PM=npm`, `PM=pnpm`, `PM=yarn`, or `PM=bun`.

`COLLECTION` is the path to the Bruno collection directory. Defaults to `.`
(the current directory). `ENV` activates a named Bruno environment via
`--env <name>`. `EXTRA_ARGS` and arguments after `--` are appended to the
command.

## Examples

```bash
task bruno:install
task bruno:run
task bruno:run COLLECTION=./api ENV=staging
task bruno:ci PM=pnpm ENV=production
task bruno:run PM=npm -- --reporter-json results.json
```
