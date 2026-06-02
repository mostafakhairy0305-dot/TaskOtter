# Corepack Taskfile Public Tasks

Corepack keeps Yarn and pnpm versions tied to a project's `packageManager`
field and can provide package manager shims for the selected Node.js runtime.
Corepack itself may need installation on Node.js releases that do not bundle it.

## Public Tasks

| Task          | Variables                                                   | Description                                                         |
| ------------- | ----------------------------------------------------------- | ------------------------------------------------------------------- |
| `node:setup`  | Optional `NODE_MANAGER`, `NODE_VERSION`                     | Install Node.js via fnm or nvm.                                     |
| `install`     | Optional `NODE_MANAGER`, `NODE_VERSION`, `COREPACK_VERSION` | Install Corepack through npm when the active Node runtime lacks it. |
| `setup`       | Optional `NODE_MANAGER`, `NODE_VERSION`, `COREPACK_VERSION` | Install Corepack when needed and enable its shims.                  |
| `version`     | Optional `NODE_MANAGER`, `NODE_VERSION`                     | Show the active Corepack version.                                   |
| `enable`      | Optional `NODE_MANAGER`, `NODE_VERSION`                     | Enable Corepack shims.                                              |
| `disable`     | Optional `NODE_MANAGER`, `NODE_VERSION`                     | Disable Corepack shims.                                             |
| `use`         | Required `PACKAGE_MANAGER`, `VERSION`                       | Pin `npm`, `pnpm`, or `yarn` in the current `package.json`.         |
| `cache:clean` | Optional `NODE_MANAGER`, `NODE_VERSION`                     | Clear cached package manager archives.                              |

## Examples

```bash
task corepack:setup
task corepack:install COREPACK_VERSION=latest
task corepack:use PACKAGE_MANAGER=yarn VERSION=stable
task corepack:use PACKAGE_MANAGER=pnpm VERSION=10.0.0
task corepack:cache:clean
```

`COREPACK_VERSION` defaults to `0.34.0` for reproducible CI/tooling. Override it
when you intentionally want another release.
