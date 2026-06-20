# Go Taskfile

## What is this Taskfile?

A cross-platform Taskfile for installing the Go toolchain and running Go
formatting, lint, vulnerability, and security checks.

macOS uses Homebrew. Linux uses the official tarball from go.dev and installs it
under `/usr/local/go` by default. Windows uses the official MSI installer from
go.dev. Development tools are installed into `GOBIN`, falling back to
`GOPATH/bin`.

## Usage

### Standalone

```sh
task -t taskfiles/go/Taskfile.yml install
task -t taskfiles/go/Taskfile.yml fmt
task -t taskfiles/go/Taskfile.yml lint
task -t taskfiles/go/Taskfile.yml lint:fix
task -t taskfiles/go/Taskfile.yml version
task -t taskfiles/go/Taskfile.yml verify
```

### Included

```yaml
includes:
  go: ./taskfiles/go/Taskfile.yml
```

Then run:

```sh
task go:install
task go:fmt
task go:lint
task go:lint:fix
task go:version
task go:verify
```

## Linting

Run every configured check:

```sh
task -t taskfiles/go/Taskfile.yml lint
```

When this Taskfile is included under the `go` namespace:

```sh
task go:lint
```

The aggregate task runs:

- `golangci-lint:lint`: runs `golangci-lint run ./...`
- `gofumpt:lint`: reports Go files under `.` that require formatting
- `goimports:lint`: reports Go files under `.` with incorrect imports
- `govulncheck:lint`: runs `govulncheck ./...`
- `gosec:lint`: runs `gosec ./...`

Each check depends on its matching `install:<tool>` task, so missing tools are
installed automatically. The formatting checks exit with a nonzero status when
files require changes.

Run an individual check or override its default arguments with `--`:

```sh
task -t taskfiles/go/Taskfile.yml gosec:lint
task -t taskfiles/go/Taskfile.yml golangci-lint:lint -- ./internal/...
task go:govulncheck:lint -- -test ./...
```

Auto-fix lint issues that supported tools can rewrite:

```sh
task -t taskfiles/go/Taskfile.yml lint:fix
task go:lint:fix -- ./internal/...
```

`lint:fix` runs `golangci-lint:lint:fix`, then `fmt` so any generated edits are
normalized with `goimports` and `gofumpt`. `golangci-lint:lint:fix` is also
available directly when you only want `golangci-lint run --fix`.

## Formatting

Format Go files in place:

```sh
task -t taskfiles/go/Taskfile.yml fmt
task go:fmt
```

The aggregate formatter runs `goimports:fmt` first, then `gofumpt:fmt`, so
imports are organized before stricter formatting is applied. Both tasks default
to `.` and accept CLI arguments after `--`:

```sh
task go:fmt -- ./internal/...
task go:goimports:fmt -- ./taskfiles/go
task go:gofumpt:fmt -- ./taskfiles/go
```

## Versions

Use `GO_VERSION` to install a specific Go toolchain release. It must use the
official release name, including the `go` prefix:

```sh
task -t taskfiles/go/Taskfile.yml install GO_VERSION=go1.26.2
task go:install GO_VERSION=go1.26.2
```

When `GO_VERSION` is empty, `install` uses the latest stable Go release. On
macOS, latest uses Homebrew while an explicit version uses the official Go
package. Linux and Windows use official Go downloads for both modes.

Each development tool has its own optional version variable:

```sh
task go:install:golangci-lint GOLANGCI_LINT_VERSION=v2.1.6
task go:install:gofumpt GOFUMPT_VERSION=v0.8.0
task go:install:goimports GOIMPORTS_VERSION=v0.36.0
task go:install:govulncheck GOVULNCHECK_VERSION=v1.1.4
task go:install:gosec GOSEC_VERSION=v2.22.7
```

An empty tool version defaults to `latest`. Supplying a tool version forces its
installer to run even when the executable already exists.

## Public Tasks

| Task                        | Description                                           | Key variables      |
| --------------------------- | ----------------------------------------------------- | ------------------ |
| `fmt`                       | Format Go files with goimports and gofumpt            | none               |
| `fmt:check`                 | Check Go file formatting with goimports and gofumpt   | none               |
| `gofumpt:fmt`               | Format Go files with gofumpt                          | none               |
| `install`                   | Install Go on the current operating system if missing | `INSTALL_DIR_UNIX`, `GO_VERSION` |
| `install:undo`              | Remove Go from the current operating system            | `INSTALL_DIR_UNIX` |
| `install:golangci-lint`     | Install golangci-lint into the global Go bin          | `GLOBAL_GO_BIN`, `GOLANGCI_LINT_VERSION` |
| `install:gofumpt`           | Install gofumpt into the global Go bin                 | `GLOBAL_GO_BIN`, `GOFUMPT_VERSION` |
| `install:goimports`         | Install goimports into the global Go bin               | `GLOBAL_GO_BIN`, `GOIMPORTS_VERSION` |
| `install:govulncheck`       | Install govulncheck into the global Go bin             | `GLOBAL_GO_BIN`, `GOVULNCHECK_VERSION` |
| `install:gosec`             | Install gosec into the global Go bin                   | `GLOBAL_GO_BIN`, `GOSEC_VERSION` |
| `lint`                      | Run all Go lint and security checks                    | none               |
| `lint:fix`                  | Auto-fix Go lint and formatting issues                 | none               |
| `golangci-lint:lint`        | Lint all Go packages with golangci-lint                | none               |
| `golangci-lint:lint:fix`    | Auto-fix Go lint issues with golangci-lint             | none               |
| `goimports:fmt`             | Format Go imports with goimports                      | none               |
| `gofumpt:lint`              | Check Go formatting with gofumpt                       | none               |
| `goimports:lint`            | Check Go imports with goimports                        | none               |
| `govulncheck:lint`          | Scan Go packages for known vulnerabilities             | none               |
| `gosec:lint`                | Scan Go packages for security issues                   | none               |
| `upgrade`                   | Upgrade Go to the selected or latest stable release    | `INSTALL_DIR_UNIX`, `GO_VERSION` |
| `version`                   | Show the installed Go version                          | none               |
| `which`                     | Show the path to the Go binary                         | none               |
| `verify`                    | Print Go version, GOROOT, and GOPATH                   | none               |

## Variables

| Variable               | Default                         | Description                                                           |
| ---------------------- | ------------------------------- | --------------------------------------------------------------------- |
| `INSTALL_DIR_UNIX`     | `/usr/local`                    | Parent directory for the Linux tarball install                        |
| `GO_ROOT_UNIX`         | `{{.INSTALL_DIR_UNIX}}/go`      | Linux Go root directory                                               |
| `GO_BIN_UNIX`          | `{{.GO_ROOT_UNIX}}/bin`         | Linux Go binary directory added to shell profiles                     |
| `GO_CMD_UNIX`          | `{{.GO_BIN_UNIX}}/go`           | Linux Go binary path used as a fallback before the shell reloads PATH |
| `GO_VERSION_URL`       | `https://go.dev/VERSION?m=text` | Endpoint used to resolve the latest stable Go version                 |
| `GO_DOWNLOAD_BASE_URL` | `https://go.dev/dl`             | Base URL for official Go downloads                                    |
| `GO_VERSION`           | empty (latest stable)           | Optional official Go release name, such as `go1.26.2`                 |
| `GOLANGCI_LINT_VERSION` | empty (`latest`)               | Optional golangci-lint module version                                 |
| `GOFUMPT_VERSION`      | empty (`latest`)                | Optional gofumpt module version                                       |
| `GOIMPORTS_VERSION`    | empty (`latest`)                | Optional goimports module version                                     |
| `GOVULNCHECK_VERSION`  | empty (`latest`)                | Optional govulncheck module version                                   |
| `GOSEC_VERSION`        | empty (`latest`)                | Optional gosec module version                                         |
| `GLOBAL_GO_BIN`        | `GOBIN` or `GOPATH/bin`         | Destination and lookup directory for installed Go development tools   |

## Notes

Linux installs replace `INSTALL_DIR_UNIX/go`. The task uses `sudo` when it is
not already running as root, then adds `GO_BIN_UNIX` to the current user's shell
profile if Go is not already available on PATH.

macOS requires Homebrew to already be installed.
