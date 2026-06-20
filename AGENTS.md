## Learned User Preferences

- When implementing an attached plan, do not edit the plan file itself; follow the plan as reference only.
- During plan execution, use existing todos (do not recreate them); mark each in_progress when starting and completed when finished; finish all todos without stopping early.
- Create git commits only when explicitly asked.
- CI must pass the `gofmt -l .` formatting check; run `gofmt -w` on offending files before pushing.
- GitHub Actions workflows in this repo should use `actions/checkout@v6` and `actions/setup-go@v6` for Node 24 runner compatibility.

## Learned Workspace Facts

- TaskOtter is a Docker-based Go GitHub Action (`github.com/mostafakhairy0305-dot/taskotter-sync-action`, Go 1.23) with entrypoint `cmd/taskotter-sync` and metadata in `action.yml`.
- The action syncs taskfile modules from `mostafakhairy0305-dot/TaskOtter-store` into the consumer repo; default target folder is `taskfiles`.
- Store modules live under `taskfiles/<module>/`; transitive dependencies are resolved from the store's `.deps.yml`.
- Destination folder names are normalized by stripping package-manager and version-manager suffixes (e.g. `eslint-pnpm-fnm` → `taskfiles/eslint`).
- Managed sync state is tracked in `<target-folder>/.taskotter-lock.yml`; PRs use branch `taskotter/sync-<configuration-hash>`.
- Core packages: `internal/config`, `internal/store`, `internal/resolver`, `internal/syncer` (split across model/load/plan/diff/apply/fs), `internal/app`, `internal/git`, `internal/github`.
- CI in `.github/workflows/test.yml` runs gofmt, `go vet`, `go test -race`, binary/Docker builds, and integration tests under `tests/integration/`; the integration job sets `setup-go` `cache: false` to avoid tar restore failures.
- Test fixtures mirror the store layout under `tests/fixtures/store/`.
