# TaskOtter Sync GitHub Action

Docker-based GitHub Action that synchronizes [TaskOtter](https://github.com/mostafakhairy0305-dot/TaskOtter-store) task modules from the official store into your repository, resolves transitive dependencies, normalizes destination folder names, updates your root `Taskfile.yml`, and opens or updates a deterministic pull request when changes exist.

## Features

- Validates all inputs before modifying files
- Resolves logical task names to store variants (`eslint` + `pnpm` + `fnm` → `eslint-pnpm-fnm`)
- Recursively resolves dependencies from `.deps.yml`
- Supports latest default branch or a pinned store Git tag
- Normalizes destination directories (`eslint-pnpm-fnm` → `taskfiles/eslint`)
- Rewrites internal Taskfile dependency includes to normalized paths
- Tracks managed files in `<target-folder>/.taskotter-lock.yml`
- Creates or updates a deterministic PR branch `taskotter/sync-<configuration-hash>`
- Never executes downloaded Taskfiles or scripts

## Permissions

```yaml
permissions:
  contents: write
  pull-requests: write
```

## Inputs

| Input | Required | Default | Description |
|-------|----------|---------|-------------|
| `tasks` | yes | — | Comma-separated or multiline list of logical task names (`eslint`, `go`, …) |
| `github-token` | yes | — | Token for pushing the sync branch and managing pull requests |
| `node-package-manager` | no | empty | `npm`, `yarn`, `pnpm`, or `bun` for Node-specific tasks |
| `node-version-manager` | no | empty | `fnm` or `nvm` (required for npm/yarn/pnpm; must be empty for bun) |
| `includes-doc` | no | `true` | Copy `README.md` and `docs/` when `true` |
| `store-version` | no | empty | Store Git tag (for example `v1.4.0`); empty uses default branch HEAD |
| `target-folder` | no | `taskfiles` | Repository-relative destination directory |

## Outputs

| Output | Description |
|--------|-------------|
| `changed` | `true` when changes were generated |
| `store-version` | Requested store tag, or empty for latest default branch |
| `source-ref` | Resolved source reference (`refs/tags/...` or `refs/heads/...`) |
| `source-sha` | Immutable source commit SHA |
| `target-folder` | Normalized target folder |
| `resolved-tasks` | JSON map of requested tasks with source/destination modules |
| `resolved-dependencies` | JSON array of dependency modules |
| `pull-request-number` | Created or updated PR number |
| `pull-request-url` | Created or updated PR URL |

## Usage

### Latest store version

```yaml
name: Sync TaskOtter Tasks

on:
  workflow_dispatch:
  schedule:
    - cron: "0 6 * * 1"

permissions:
  contents: write
  pull-requests: write

jobs:
  sync:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
        with:
          fetch-depth: 0

      - name: Synchronize TaskOtter modules
        id: taskotter
        uses: mostafakhairy0305-dot/taskotter-sync-action@v1
        with:
          github-token: ${{ secrets.GITHUB_TOKEN }}
          tasks: |
            eslint
            prettier
            typescript
            go
          node-package-manager: pnpm
          node-version-manager: fnm
          includes-doc: true
          target-folder: taskfiles
```

### Pinned store tag

```yaml
- uses: mostafakhairy0305-dot/taskotter-sync-action@v1
  with:
    github-token: ${{ secrets.GITHUB_TOKEN }}
    store-version: v1.4.0
    target-folder: automation/taskfiles
    tasks: |
      eslint
      prettier
      go
    node-package-manager: pnpm
    node-version-manager: fnm
```

### Bun (no version manager)

```yaml
- uses: mostafakhairy0305-dot/taskotter-sync-action@v1
  with:
    github-token: ${{ secrets.GITHUB_TOKEN }}
    tasks: |
      eslint
      prettier
      typescript
    node-package-manager: bun
    includes-doc: false
```

### Non-Node tasks only

```yaml
- uses: mostafakhairy0305-dot/taskotter-sync-action@v1
  with:
    github-token: ${{ secrets.GITHUB_TOKEN }}
    target-folder: build/taskfiles
    tasks: |
      actionlint
      shellcheck
      docker
      go
```

## Prerequisites

Your repository must already contain a root `Taskfile.yml`. The action updates managed `includes` entries but does not create a new root Taskfile.

## Behavior

### Module resolution

- Non-Node tasks (`go`, `docker`, `shellcheck`, …) resolve directly by name
- Node tasks require `node-package-manager`
- `npm`, `yarn`, and `pnpm` also require `node-version-manager`
- `bun` rejects `node-version-manager`

### Destination layout

Modules copy to `<target-folder>/<normalized-name>/`:

```text
taskfiles/eslint/Taskfile.yml      ← eslint-pnpm-fnm
taskfiles/pnpm/Taskfile.yml        ← pnpm-fnm
taskfiles/corepack/Taskfile.yml    ← corepack-fnm
taskfiles/fnm/Taskfile.yml         ← fnm
```

### Lock file and metadata

- Lock file: `<target-folder>/.taskotter-lock.yml`
- Root metadata: `.taskotter/metadata.yml` (target folder migration)

### Pull requests

- Branch: `taskotter/sync-<configuration-hash>`
- Title: `chore(taskotter): synchronize TaskOtter taskfiles`
- Updates an existing open PR for the same branch instead of creating duplicates
- No commit or PR when synchronized content is unchanged

## Validation failures

| Condition | Result |
|-----------|--------|
| Empty or invalid task names | Fail before download |
| Missing store task or variant | Fail with close matches |
| Bun + version manager | Fail |
| npm/yarn/pnpm without version manager | Fail |
| Node task without package manager | Fail |
| Invalid or unsafe `store-version` | Fail |
| Unsafe `target-folder` | Fail |
| Missing root `Taskfile.yml` | Fail |
| Unmanaged existing destination directory | Fail |
| Destination name collision | Fail |

## Development

```bash
go vet ./...
go test -race ./...
go build ./cmd/taskotter-sync
docker build -t taskotter-sync:local .
```

## Security

- Input validation before any filesystem writes
- Secure tar.gz extraction with size limits and traversal protection
- Git invoked with argument arrays (no shell interpolation)
- GitHub token is never logged
- Downloaded Taskfiles and scripts are never executed

## License

MIT
