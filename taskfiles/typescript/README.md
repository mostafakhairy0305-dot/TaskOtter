# TypeScript Taskfile Public Tasks

## What is this Taskfile?

This Taskfile wraps common TypeScript workflows behind consistent, cross-platform
task commands. It covers installing TypeScript tooling, running `.ts` files with
`tsx`, static type-checking with `tsc`, compiling builds, inspecting compiler
configuration, and cleaning generated output.

`tsserver` is included for editor awareness only. It ships with the `typescript`
package and is managed by editors such as VS Code, Neovim, and other TypeScript
integrations.

## Public Tasks

| Task | Variables | Description |
|---|---|---|
| `version` | Optional `PM_EXEC` | Show resolved `tsc`, `tsx`, and `tsserver` information. |
| `tsserver:info` | Optional `PM_EXEC` | Show where `tsserver` resolves from and how editors use it. |
| `install` | — | Auto-detect the package manager and install TypeScript dev dependencies. |
| `install:auto` | — | Install `typescript`, `tsx`, and `@types/node` using lockfile detection. |
| `install:npm` | — | Install TypeScript dev dependencies with npm. |
| `install:yarn` | — | Install TypeScript dev dependencies with Yarn. |
| `install:pnpm` | — | Install TypeScript dev dependencies with pnpm. |
| `install:bun` | — | Install TypeScript dev dependencies with Bun. |
| `run` | Optional `FILE`, `TSX_FLAGS`, `CLI_ARGS`, `PM_EXEC` | Execute one TypeScript file once with `tsx`. |
| `dev` | Optional `FILE`, `TSX_FLAGS`, `PM_EXEC` | Run one TypeScript file in `tsx watch` mode. |
| `typecheck` | Optional `TSCONFIG`, `TSC_FLAGS`, `PM_EXEC` | Run `tsc --noEmit` for the full project. |
| `typecheck:watch` | Optional `TSCONFIG`, `TSC_FLAGS`, `PM_EXEC` | Run `tsc --noEmit --watch`. |
| `typecheck:files` | Required `FILES`; optional `TSC_FLAGS`, `PM_EXEC` | Type-check explicit files without loading `tsconfig.json`. |
| `build` | Optional `TSCONFIG`, `TSC_FLAGS`, `PM_EXEC` | Compile the project with `tsc --noEmitOnError`. |
| `build:watch` | Optional `TSCONFIG`, `TSC_FLAGS`, `PM_EXEC` | Compile in watch mode with `tsc --noEmitOnError --watch`. |
| `build:clean` | Optional `OUT_DIR`, `TSCONFIG`, `TSC_FLAGS`, `PM_EXEC` | Remove the output directory and run a fresh compile. |
| `emit:dts` | Optional `OUT_DIR`, `TSCONFIG`, `TSC_FLAGS`, `PM_EXEC` | Emit declaration files only. |
| `config:show` | Optional `TSCONFIG`, `PM_EXEC` | Print the fully resolved TypeScript config. |
| `config:init` | Optional `PM_EXEC` | Generate a starter `tsconfig.json` with `tsc --init`. |
| `config:files` | Optional `TSCONFIG`, `PM_EXEC` | List every file included in the compilation. |
| `config:diagnostics` | Optional `TSCONFIG`, `PM_EXEC` | Print compiler performance diagnostics. |
| `config:trace` | Optional `TRACE_DIR`, `TSCONFIG`, `TSC_FLAGS`, `PM_EXEC` | Emit a TypeScript performance trace. |
| `start` | Optional `OUTFILE` | Run compiled JavaScript with Node.js. |
| `ci` | Optional `TSCONFIG`, `TSC_FLAGS`, `PM_EXEC` | Run the same strict no-emit type-check used by CI. |
| `clean` | Optional `OUT_DIR` | Remove the compiled output directory. |
| `clean:all` | Optional `OUT_DIR` | Remove output, incremental build cache, and trace output. |

## Examples

```bash
task typescript:install
task typescript:version
task typescript:config:init

task typescript:run FILE=scripts/seed.ts
task typescript:dev FILE=src/server.ts TSX_FLAGS="--env-file .env"

task typescript:typecheck
task typescript:typecheck TSCONFIG=tsconfig.strict.json
task typescript:typecheck:files FILES="src/index.ts src/api.ts" TSC_FLAGS="--strict"

task typescript:build
task typescript:build:clean OUT_DIR=build
task typescript:emit:dts OUT_DIR=types

task typescript:config:show
task typescript:config:files
task typescript:config:diagnostics
task typescript:config:trace TRACE_DIR=.traces/tsc

task typescript:start OUTFILE=dist/server.js
task typescript:clean --yes
task typescript:clean:all --yes
```

## Notes

`tsx` is fast because it strips types and executes through esbuild; it does not
catch type errors. Use `typecheck`, `build`, or `ci` before committing.

`typecheck:files` intentionally bypasses `tsconfig.json`, because TypeScript
ignores project configuration when explicit files are passed on the command
line. Use it only for controlled quick checks.
