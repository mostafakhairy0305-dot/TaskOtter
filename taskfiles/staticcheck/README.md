# Staticcheck Taskfile Public Tasks

## What is this Taskfile?

A cross-platform Taskfile for installing and running
[Staticcheck](https://staticcheck.dev), the Go static analysis tool.

Staticcheck is downloaded from the pinned GitHub release and installed into the
repo-local `.toolbin` directory by version. The `run` task also ensures the Go
toolchain is installed through the local Go Taskfile before analysis starts.

## Usage

### Standalone

```sh
task -t taskfiles/staticcheck/Taskfile.yml install
task -t taskfiles/staticcheck/Taskfile.yml run
task -t taskfiles/staticcheck/Taskfile.yml version
```

Pass Staticcheck arguments after `--`:

```sh
task -t taskfiles/staticcheck/Taskfile.yml run -- ./cmd/... ./internal/...
```

### Included

```yaml
includes:
  staticcheck: ./taskfiles/staticcheck/Taskfile.yml
```

Then run:

```sh
task staticcheck:run
task staticcheck:version
```

## Public Tasks

| Task | Description | Key variables |
|------|-------------|---------------|
| `install` | Install the pinned Staticcheck binary into `.toolbin` | `STATICCHECK_VERSION`, `TOOLBIN` |
| `run` | Run Staticcheck against Go packages | `TARGETS`, `EXTRA_ARGS` |
| `version` | Print the installed Staticcheck version | none |

## Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `STATICCHECK_VERSION` | `2026.1` | Staticcheck release tag to download |
| `STATICCHECK_RELEASE_BASE_URL` | `https://github.com/dominikh/go-tools/releases/download` | Base URL for Staticcheck release assets |
| `TOOLBIN` | `{{.TASKFILE_DIR}}/../../.toolbin` | Repo-local tool install directory |
| `STATICCHECK_INSTALL_DIR` | `{{.TOOLBIN}}/staticcheck/{{.STATICCHECK_VERSION}}` | Versioned Staticcheck install directory |
| `STATICCHECK_BIN_UNIX` | `{{.STATICCHECK_INSTALL_DIR}}/staticcheck` | Staticcheck binary path on macOS and Linux |
| `STATICCHECK_BIN_WINDOWS` | `{{.STATICCHECK_INSTALL_DIR}}/staticcheck.exe` | Staticcheck binary path on Windows |
| `TARGETS` | `./...` | Default package pattern for `run` |
| `EXTRA_ARGS` | empty | Extra arguments appended when CLI_ARGS is not provided |
