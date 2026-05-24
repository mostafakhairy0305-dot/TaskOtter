# Go Taskfile Public Tasks

## What is this Taskfile?

A cross-platform Taskfile for installing and checking the Go toolchain.

macOS uses Homebrew. Linux uses the official tarball from go.dev and installs it
under `/usr/local/go` by default. Windows uses the official MSI installer from
go.dev.

## Usage

### Standalone

```sh
task -t taskfiles/go/Taskfile.yml install
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
task go:version
task go:verify
```

## Public Tasks

| Task | Description | Key variables |
|------|-------------|---------------|
| `install` | Install Go on the current operating system if missing | `INSTALL_DIR_UNIX` |
| `upgrade` | Upgrade Go to the latest stable release | `INSTALL_DIR_UNIX` |
| `version` | Show the installed Go version | none |
| `which` | Show the path to the Go binary | none |
| `verify` | Print Go version, GOROOT, and GOPATH | none |

## Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `INSTALL_DIR_UNIX` | `/usr/local` | Parent directory for the Linux tarball install |
| `GO_ROOT_UNIX` | `{{.INSTALL_DIR_UNIX}}/go` | Linux Go root directory |
| `GO_BIN_UNIX` | `{{.GO_ROOT_UNIX}}/bin` | Linux Go binary directory added to shell profiles |
| `GO_CMD_UNIX` | `{{.GO_BIN_UNIX}}/go` | Linux Go binary path used as a fallback before the shell reloads PATH |
| `GO_VERSION_URL` | `https://go.dev/VERSION?m=text` | Endpoint used to resolve the latest stable Go version |
| `GO_DOWNLOAD_BASE_URL` | `https://go.dev/dl` | Base URL for official Go downloads |

## Notes

Linux installs replace `INSTALL_DIR_UNIX/go`. The task uses `sudo` when it is
not already running as root, then adds `GO_BIN_UNIX` to the current user's shell
profile if Go is not already available on PATH.

macOS requires Homebrew to already be installed.
