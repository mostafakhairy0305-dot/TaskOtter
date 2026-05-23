# pnpm Taskfile Public Tasks

This Taskfile wraps common pnpm project workflows while loading Node.js through
`fnm` by default or `nvm` when `NODE_MANAGER=nvm`. pnpm commands run through
Corepack so the project's `packageManager` field controls the pnpm version. The
runners set `COREPACK_ENABLE_AUTO_PIN=1`, so first use adds the field when it is
missing.

## Public Tasks

| Task | Variables | Description |
|---|---|---|
| `node:setup` | Optional `NODE_MANAGER`, `NODE_VERSION` | Install Node.js via fnm or nvm. |
| `manager:setup` | Optional `NODE_MANAGER`, `NODE_VERSION` | Install Corepack when needed and enable its shims. |
| `manager:pin` | Required `PACKAGE_MANAGER_VERSION` | Pin pnpm in `package.json` with Corepack. |
| `version` | Optional `NODE_MANAGER`, `NODE_VERSION` | Show active Node.js and pnpm versions. |
| `install` | Optional `NODE_MANAGER`, `NODE_VERSION` | Run `pnpm install`. |
| `ci` | Optional `NODE_MANAGER`, `NODE_VERSION` | Run `pnpm install --frozen-lockfile` with `pnpm-lock.yaml`. |
| `run` | Required `SCRIPT`; optional `NODE_MANAGER`, `NODE_VERSION` | Run a script via `pnpm run`. |
| `dev` | Optional `NODE_MANAGER`, `NODE_VERSION` | Run `pnpm run dev`. |
| `test` | Optional `NODE_MANAGER`, `NODE_VERSION` | Run `pnpm test`. |
| `build` | Optional `NODE_MANAGER`, `NODE_VERSION` | Run `pnpm run build`. |
| `lint` | Optional `NODE_MANAGER`, `NODE_VERSION` | Run `pnpm run lint`. |
| `format` | Optional `NODE_MANAGER`, `NODE_VERSION` | Run `pnpm run format`. |
| `typecheck` | Optional `NODE_MANAGER`, `NODE_VERSION` | Run `pnpm run typecheck`. |
| `outdated` | Optional `NODE_MANAGER`, `NODE_VERSION` | List outdated packages without failing. |
| `outdated:strict` | Optional `NODE_MANAGER`, `NODE_VERSION` | List outdated packages with the pnpm exit code. |
| `update` | Optional `NODE_MANAGER`, `NODE_VERSION` | Run `pnpm update`. |
| `audit` | Optional `NODE_MANAGER`, `NODE_VERSION` | Run strict `pnpm audit`. |
| `audit:report` | Optional `NODE_MANAGER`, `NODE_VERSION` | Report audit findings without failing. |
| `audit:fix` | Optional `NODE_MANAGER`, `NODE_VERSION` | Run `pnpm audit --fix`. |
| `audit:json` | Optional `NODE_MANAGER`, `NODE_VERSION` | Emit `pnpm audit --json`. |
| `store:prune` | Optional `NODE_MANAGER`, `NODE_VERSION` | Remove unreferenced pnpm store packages. |
| `clean` | - | Remove `node_modules`. |
| `clean:all` | - | Remove `node_modules` and `pnpm-lock.yaml`. |

## Examples

```bash
task pnpm:node:setup NODE_VERSION=22
task pnpm:manager:setup
task pnpm:manager:pin PACKAGE_MANAGER_VERSION=latest
task pnpm:install
task pnpm:ci
task pnpm:run SCRIPT=test -- --watch
task pnpm:audit:report
task pnpm:store:prune
task pnpm:clean --yes
```
