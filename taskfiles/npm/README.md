# NPM Taskfile Public Tasks

## What is this Taskfile?

This Taskfile wraps common `npm` operations — installing dependencies, running scripts, auditing, and cleaning — behind consistent, cross-platform task commands. It delegates Node.js installation to either **fnm** or **nvm** (your choice), keeping each concern in its own taskfile.

---

## Node manager selection

Every task that invokes `npm` loads the Node.js runtime through the node manager you choose. Set `NODE_MANAGER` to `fnm` (the default) or `nvm`:

```bash
task npm:install                        # uses fnm (default)
task npm:install NODE_MANAGER=nvm       # uses nvm instead
```

The selected manager must be included in your root `Taskfile.yml`. Both are available in this collection:

```yaml
includes:
  fnm:
    taskfile: taskfiles/fnm/Taskfile.yml
  nvm:
    taskfile: taskfiles/nvm/Taskfile.yml
  corepack:
    taskfile: taskfiles/corepack/Taskfile.yml
  npm:
    taskfile: taskfiles/npm/Taskfile.yml
```

---

## Installing Node.js

Before using npm tasks on a fresh machine, install Node.js via the selected manager:

```bash
task npm:node:setup                             # install LTS via fnm
task npm:node:setup NODE_MANAGER=nvm            # install LTS via nvm
task npm:node:setup NODE_VERSION=20             # install Node.js 20.x
task npm:node:setup NODE_VERSION=22.0.0         # install exact version
```

This delegates to `task fnm:node:install` or `task nvm:node:install` depending on `NODE_MANAGER`.

---

## Public Tasks

| Task | Variables | Description |
|---|---|---|
| `add` | Required `PACKAGES`; optional `EXTRA_ARGS` | Add packages as devDependencies with `npm install -D`. |
| `node:setup` | Optional `NODE_MANAGER`, `NODE_VERSION` | Install Node.js and npm via the selected node manager. |
| `version` | Optional `NODE_MANAGER`, `NODE_VERSION` | Show the active Node.js and npm versions. |
| `install` | Optional `NODE_MANAGER`, `NODE_VERSION` | Run `npm install` to install all dependencies from `package.json`. |
| `ci` | Optional `NODE_MANAGER`, `NODE_VERSION` | Run `npm ci` for a clean, reproducible install from `package-lock.json`. |
| `run` | Required `SCRIPT`; optional `NODE_MANAGER`, `NODE_VERSION` | Run a `package.json` script by name. Example: `SCRIPT=dev`. |
| `dev` | Optional `NODE_MANAGER`, `NODE_VERSION` | Run `npm run dev`. |
| `exec` | Required `BINARY`; optional `ARGS`, `EXTRA_ARGS` | Execute a local project binary via `npm exec --`. |
| `test` | Optional `NODE_MANAGER`, `NODE_VERSION` | Run `npm test`. |
| `build` | Optional `NODE_MANAGER`, `NODE_VERSION` | Run `npm run build`. |
| `lint` | Optional `NODE_MANAGER`, `NODE_VERSION` | Run `npm run lint`. |
| `format` | Optional `NODE_MANAGER`, `NODE_VERSION` | Run `npm run format`. |
| `typecheck` | Optional `NODE_MANAGER`, `NODE_VERSION` | Run `npm run typecheck`. |
| `manager:setup` | Optional `NODE_MANAGER`, `NODE_VERSION` | Install Corepack when needed and enable its shims. |
| `manager:pin` | Required `PACKAGE_MANAGER_VERSION` | Pin npm in `package.json` with Corepack. |
| `outdated` | Optional `NODE_MANAGER`, `NODE_VERSION` | List newer package versions without failing the task. |
| `outdated:strict` | Optional `NODE_MANAGER`, `NODE_VERSION` | List newer package versions and propagate `npm outdated` failures for CI. |
| `update` | Optional `NODE_MANAGER`, `NODE_VERSION` | Update all packages within declared version ranges. |
| `audit` | Optional `NODE_MANAGER`, `NODE_VERSION` | Audit vulnerabilities and propagate `npm audit` failures for CI. |
| `audit:report` | Optional `NODE_MANAGER`, `NODE_VERSION` | Report vulnerabilities without failing the task. |
| `audit:fix` | Optional `NODE_MANAGER`, `NODE_VERSION` | Auto-fix vulnerabilities where a non-breaking fix exists. |
| `audit:json` | Optional `NODE_MANAGER`, `NODE_VERSION` | Emit `npm audit --json` output for tooling. |
| `doctor` | Optional `NODE_MANAGER`, `NODE_VERSION` | Check npm environment health with `npm doctor`. |
| `cache:clean` | Optional `NODE_MANAGER`, `NODE_VERSION` | Clear the npm cache with `npm cache clean --force`. |
| `clean` | — | Remove `node_modules`. |
| `clean:all` | — | Remove `node_modules` and `package-lock.json`. |

---

## Dependency workflow

```bash
# Install or restore dependencies
task npm:install

# Clean install (CI-style, from lock file)
task npm:ci

# Remove and restore from scratch
task npm:clean --yes
task npm:install

# Nuclear clean
task npm:clean:all --yes
task npm:install
```

## Running scripts

```bash
task npm:run SCRIPT=dev          # npm run dev
task npm:run SCRIPT=start        # npm run start
task npm:dev                     # npm run dev
task npm:test                    # npm test
task npm:build                   # npm run build
task npm:lint                    # npm run lint
task npm:format                  # npm run format
task npm:typecheck               # npm run typecheck
task npm:manager:setup           # ensure Corepack is available
task npm:manager:pin PACKAGE_MANAGER_VERSION=latest
```

## Keeping dependencies healthy

```bash
task npm:outdated                # see what has newer versions
task npm:outdated:strict         # fail when npm outdated exits non-zero
task npm:update                  # update within declared ranges
task npm:audit                   # fail when npm audit finds vulnerabilities
task npm:audit:report            # report vulnerabilities without failing
task npm:audit:json              # emit npm audit output as JSON
task npm:audit:fix               # auto-fix where possible
task npm:doctor                  # inspect npm environment health
task npm:cache:clean             # clear npm cache data
```

---

## Security notes

All npm project commands run through `corepack npm` inside `bash -c` on Unix or PowerShell on Windows with the selected Node.js runtime loaded first. The runners set `COREPACK_ENABLE_AUTO_PIN=1`, so Corepack adds a missing `packageManager` field the first time the manager is used in a project. `manager:pin` remains available when you want to pick the package manager version explicitly. No credentials, tokens, or registry configuration are set by this Taskfile; those are managed by your npm config (`~/.npmrc`) and the project's `.npmrc` as usual.
