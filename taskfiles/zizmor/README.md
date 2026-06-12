# zizmor

A [TaskOtter](https://github.com/mostafakhairy0305-dot/TaskOtter) module for [zizmor](https://github.com/woodruffw/zizmor) — a security auditor for GitHub Actions workflow files.

## What is this Taskfile?

This module provides tasks to audit, install, and manage [zizmor](https://github.com/woodruffw/zizmor). zizmor detects security issues in GitHub Actions workflows including expression injection, excessive permissions, use of mutable actions, and other dangerous patterns.

## Usage

### Standalone

```sh
task -t taskfiles/zizmor/Taskfile.yml audit
task -t taskfiles/zizmor/Taskfile.yml audit TARGETS=.github/workflows/main.yml
task -t taskfiles/zizmor/Taskfile.yml audit EXTRA_ARGS="--min-severity high"
task -t taskfiles/zizmor/Taskfile.yml audit EXTRA_ARGS="--gh-token $GITHUB_TOKEN"
```

### Included in your Taskfile

```yaml
includes:
  zizmor:
    taskfile: taskfiles/zizmor/Taskfile.yml
    vars:
      TARGETS_OVERRIDE: "{{.TARGETS}}"
      EXTRA_ARGS_OVERRIDE: "{{.EXTRA_ARGS}}"
```

Then run:

```sh
task zizmor:audit
task zizmor:install
```

## Public Tasks

| Task | Description |
|---|---|
| `audit` | Audit GitHub Actions workflows for security issues (TARGETS=path) |
| `install` | Install zizmor on the current operating system |
| `install:undo` | Remove zizmor from the current operating system |
| `upgrade` | Upgrade zizmor to the pinned ZIZMOR_VERSION |
| `version` | Show the installed zizmor version |

## Variables

| Variable | Default | Description |
|---|---|---|
| `EXTRA_ARGS` | `"--offline"` | Additional flags passed to `zizmor` (e.g. `--format`, `--min-severity`, `--gh-token`) |
| `TARGETS` | `".github"` | Path to audit; scans workflows and composite actions under `.github` |
| `ZIZMOR_VERSION` | `"1.25.2"` | Pinned release version for binary download |

## Notes

- **All platforms** install via direct binary download from [GitHub Releases](https://github.com/woodruffw/zizmor/releases).
- **macOS** downloads an `apple-darwin` binary to `/usr/local/bin`. Requires `curl`, `tar`, and `install`. Both `arm64` and `x86_64` are supported.
- **Linux** downloads an `unknown-linux-gnu` binary to `/usr/local/bin`. Only `x86_64` and `aarch64` architectures are supported.
- **Windows** downloads an `x86_64-pc-windows-msvc.zip` binary to `%USERPROFILE%\bin`. Add `%USERPROFILE%\bin` to your `PATH` after install.
- The `upgrade` task re-downloads and overwrites the existing binary with the pinned `ZIZMOR_VERSION`. To upgrade to a newer release, update `ZIZMOR_VERSION` in the Taskfile vars.
- The `audit` task auto-installs zizmor if it is not already present in `PATH`.
