# JavaScript Package Manager Helper Taskfile

## What is this Taskfile?

This is an internal helper used by the JavaScript and TypeScript tooling
Taskfiles. It centralizes package-manager detection and command mapping so each
tool module can stay small and consistent.

## Internal Tasks

| Task          | Variables                                                                                        | Description                                                       |
| ------------- | ------------------------------------------------------------------------------------------------ | ----------------------------------------------------------------- |
| `install`     | Required `PACKAGES`; optional `PM`, `EXTRA_ARGS`, `CLI_ARGS`                                     | Install packages as local dev dependencies.                       |
| `exec`        | Required `BINARY`; optional `PM`, `ARGS`, `EXTRA_ARGS`, `CLI_ARGS`                               | Execute a local project binary.                                   |
| `exec:ignore` | Required `BINARY`; optional `PM`, `ARGS`, `IGNORE_PATH`, `IGNORE_FLAG`, `EXTRA_ARGS`, `CLI_ARGS` | Execute a local binary and include an ignore file when it exists. |

## Package Manager Mapping

When `PM` is empty, the helper detects the package manager from lock files:
`bun.lock` or `bun.lockb`, then `pnpm-lock.yaml`, then `yarn.lock`, then
`package-lock.json`, and finally npm.

| PM     | Install                     | Execute                |
| ------ | --------------------------- | ---------------------- |
| `npm`  | `npm install -D <packages>` | `npm exec -- <binary>` |
| `pnpm` | `pnpm add -D <packages>`    | `pnpm exec <binary>`   |
| `yarn` | `yarn add -D <packages>`    | `yarn exec <binary>`   |
| `bun`  | `bun add -d <packages>`     | `bun x <binary>`       |

Each PM's `add` and `exec` commands are implemented in the dedicated PM
taskfiles (`npm`, `yarn`, `pnpm`, `bun`) and called via their `add` and `exec`
tasks. `js-pm` includes all four as optional dependencies.

Yarn Classic usually supports `yarn exec`; if a pinned Yarn version does not,
run the equivalent local binary with `yarn <binary> ...`.
