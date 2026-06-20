## Learned User Preferences

- When implementing an attached plan, do not edit the plan file itself; follow the plan as reference only.
- During plan execution, use existing todos (do not recreate them); mark each in_progress when starting and completed when finished; finish all todos without stopping early.
- Create git commits only when explicitly asked.
- CI must pass the `gofmt -l .` formatting check; run `gofmt -w` on offending files before pushing.
- GitHub Actions workflows in this repo should use `actions/checkout@v6` and `actions/setup-go@v6` for Node 24 runner compatibility.

## Learned Workspace Facts

- TaskOtter is a Docker-based GitHub Action (Go 1.23) in `mostafakhairy0305-dot/TaskOtter`; the container entrypoint is `/taskotter`.
- TaskOtter syncs taskfile modules from `mostafakhairy0305-dot/TaskOtter-store` into the consumer repo; default target folder is `taskfiles`.
- Store modules live under `taskfiles/<module>/`; transitive dependencies are resolved from the store's `.deps.yml`.
- Destination folder names are normalized by stripping package-manager and version-manager suffixes (e.g. `eslint-pnpm-fnm` → `taskfiles/eslint`).
- Managed sync state is tracked in `<target-folder>/.taskotter-lock.yml`; PRs use branch `taskotter/sync-<configuration-hash>`.
- Sync skips `*_test.*` files; docs (`README.md`, `docs/`) are copied only when `includes-doc: true`; root `Taskfile.yml` includes merge each module's top-level `vars`.
- The action `js` input (YAML: `runtime`, `package-manager`, `version-manager`) replaces separate node-package-manager inputs for Node task resolution.
- Core packages: `internal/config`, `internal/store`, `internal/resolver`, `internal/syncer` (split across model/load/plan/diff/apply/fs), `internal/app`, `internal/git`, `internal/github`.
- `DefaultBranch()` falls back through symbolic-ref → rev-parse → remote set-head → remote show when `origin/HEAD` is missing (common on GHA Git 2.50+); `Stage()` uses `git add -f` for gitignored `.taskotter/metadata.yml`.
- Docker container actions expose hyphenated `INPUT_*` env vars; `internal/config` reads both hyphen and underscore forms and falls back to `GITHUB_TOKEN`.
- CI in `.github/workflows/test.yml` runs gofmt, `go vet`, `go test -race`, binary/Docker builds, and integration tests under `tests/integration/`; integration sets `setup-go` `cache: false`; the `itself` job intentionally fails when TaskOtter `changed: true` (with `::error` annotations) until the sync PR is merged.
- Test fixtures mirror the store layout under `tests/fixtures/store/`.
