# Staticcheck Taskfile Public Tasks

## What is this Taskfile?

A cross-platform Taskfile for installing and running
[Staticcheck](https://staticcheck.dev), the Go static analysis tool.

Staticcheck is downloaded from the pinned GitHub release and installed into the
global Go bin directory (`GOBIN`, or `GOPATH/bin` when `GOBIN` is unset). The
`run` task also ensures the Go toolchain is installed through the local Go
Taskfile before analysis starts.

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

| Task      | Description                                           | Key variables                          |
| --------- | ----------------------------------------------------- | -------------------------------------- |
| `install` | Install the pinned Staticcheck binary into the Go bin | `STATICCHECK_VERSION`, `GLOBAL_GO_BIN` |
| `run`     | Run Staticcheck against Go packages                   | none (pass args via `--`)              |
| `version` | Print the installed Staticcheck version               | none                                   |

## Variables

| Variable                       | Default                                                  | Description                                                               |
| ------------------------------ | -------------------------------------------------------- | ------------------------------------------------------------------------- |
| `STATICCHECK_VERSION`          | `2026.1`                                                 | Staticcheck release tag to download and enforce                           |
| `STATICCHECK_RELEASE_BASE_URL` | `https://github.com/dominikh/go-tools/releases/download` | Base URL for Staticcheck release assets                                   |
| `GLOBAL_GO_BIN`                | `$(go env GOBIN)` or `$(go env GOPATH)/bin`              | Go bin directory where the Staticcheck binary is installed                |
