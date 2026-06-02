# sqlfluff Taskfile

## What is this Taskfile?

A cross-platform Taskfile for installing sqlfluff, managing upgrades, linting
and auto-fixing SQL files, and generating a project configuration.

sqlfluff is installed via [uv](../uv/) into an isolated tool environment so it
never conflicts with project dependencies.

## Usage

### Standalone

```sh
task -t taskfiles/sqlfluff/Taskfile.yml install
task -t taskfiles/sqlfluff/Taskfile.yml config:init
task -t taskfiles/sqlfluff/Taskfile.yml lint
```

### Included

```yaml
includes:
  sqlfluff: ./taskfiles/sqlfluff/Taskfile.yml
```

Then run:

```sh
task sqlfluff:install
task sqlfluff:lint
task sqlfluff:fix DIALECT=postgres
```

## Public Tasks

| Task           | Description                                     | Key variables                                |
| -------------- | ----------------------------------------------- | -------------------------------------------- |
| `install`      | Install sqlfluff on the current OS if missing   | none                                         |
| `install:undo` | Remove sqlfluff from the current OS             | none                                         |
| `upgrade`      | Upgrade sqlfluff to the latest release          | none                                         |
| `version`      | Show the installed sqlfluff version             | none                                         |
| `lint`         | Lint SQL files with sqlfluff                    | `TARGETS`, `CONFIG`, `DIALECT`, `EXTRA_ARGS` |
| `fix`          | Auto-fix SQL lint violations                    | `TARGETS`, `CONFIG`, `DIALECT`, `EXTRA_ARGS` |
| `parse`        | Print the sqlfluff parse tree for SQL files     | `TARGETS`, `CONFIG`, `DIALECT`, `EXTRA_ARGS` |
| `config:init`  | Create a default `.sqlfluff` configuration file | none                                         |

## Variables

| Variable     | Default   | Description                                                  |
| ------------ | --------- | ------------------------------------------------------------ |
| `TARGETS`    | `.`       | Files or directories to lint/fix/parse                       |
| `CONFIG`     | _(empty)_ | Path to a sqlfluff config file passed via `--config`         |
| `DIALECT`    | _(empty)_ | SQL dialect passed via `--dialect` (e.g. `ansi`, `postgres`) |
| `EXTRA_ARGS` | _(empty)_ | Extra flags forwarded to sqlfluff                            |

## Notes

**`config:init`** writes a `.sqlfluff` file in the current directory and is
skipped if the file already exists. To regenerate, delete `.sqlfluff` first.

**Dialect:** sqlfluff requires a dialect to lint most SQL. Either set `DIALECT`
or declare it in `.sqlfluff` under `[sqlfluff] dialect = <name>`.
