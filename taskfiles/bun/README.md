# Bun Taskfile Public Tasks

## What is Bun?

Bun is an all-in-one JavaScript runtime and toolkit — a single binary that replaces Node.js, npm, a bundler, and a test runner. It is written in Zig and designed to be significantly faster than Node.js for startup, module resolution, and package installation.

Key characteristics:

- **Single binary** — one executable provides the runtime, package manager (`bun install`), bundler, and test runner.
- **Node.js compatible** — runs most Node.js programs and npm packages without modification.
- **Fast install** — `bun install` is typically 10–100× faster than npm due to a global binary cache and parallel fetching.
- **Cross-platform** — one tool for Linux, macOS, and Windows (Windows 10 build 17763 or later).

---

This document describes the public tasks exposed by the Bun Taskfile.

The Taskfile provides one cross-platform interface for installing, removing, and upgrading Bun.

Linux and macOS use the official Bun install script. Windows uses the official PowerShell script.

---

## Auto-install behaviour

Every task that requires Bun automatically installs it first if it is not already present — you do not need to run `task install` manually before using `version` or any `upgrade` task.

Installs are **idempotent**: the internal install task has a `status` check that exits early when Bun is already present and no specific version was requested, so running any task multiple times is safe.

| Task | Auto-installs |
|---|---|
| `version` | Bun (if missing) |
| `upgrade`, `upgrade:canary`, `upgrade:stable` | Bun (if missing) |
| `install:undo` | — (removal; Bun being absent is already the goal) |

---

## Public Tasks

| Task | Aliases | Variables | Description |
|---|---|---|---|
| `install` | — | Optional `VERSION` | Install Bun for the current operating system. Pass `VERSION=1.x.y` for a specific release. |
| `install:undo` | `uninstall` | — | Remove Bun from the current operating system. |
| `version` | — | — | Show the installed Bun version and revision. Auto-installs Bun if missing. |
| `upgrade` | — | — | Upgrade Bun to the latest stable release. Auto-installs Bun if missing. |
| `upgrade:canary` | — | — | Switch Bun to the latest canary (pre-release) build. Auto-installs Bun if missing. |
| `upgrade:stable` | — | — | Switch Bun back from canary to the latest stable release. Auto-installs Bun if missing. |

---

## Install Bun

Install the latest Bun release for the current platform:

```bash
task install
```

Install a specific release:

```bash
task install VERSION=1.1.38
```

All other tasks call `install` automatically, so this is only needed if you want to install Bun without doing anything else yet, or if you need a specific version pinned.

---

## Remove Bun

```bash
task install:undo
# or
task uninstall
```

On Linux and macOS this removes `~/.bun` entirely and reports any shell profile files that may still reference Bun's PATH so you can clean them manually.

On Windows, the task runs the official Bun uninstall script if present at `%USERPROFILE%\.bun\uninstall.ps1`. If the script is missing but the `.bun` directory still exists, the directory is removed directly. The task detects Bun by directory presence rather than PATH, so it works correctly even if a fresh install has not yet been added to PATH.

---

## Check the installed version

```bash
task version
```

Prints the version number and exact commit revision. Installs Bun first if it is not already present.

---

## Upgrade Bun

Upgrade to the latest stable release:

```bash
task upgrade
```

Switch to the latest canary (pre-release) build:

```bash
task upgrade:canary
```

Canary builds are untested and automatically upload crash reports to the Bun team. Use only in non-production environments.

Switch back from canary to the latest stable release:

```bash
task upgrade:stable
```

All upgrade tasks install Bun first if it is not already present.

---

## Security Notes

**Unix install script**: The `install` task downloads the official Bun install script from `BUN_INSTALL_URL` to a temporary file using `curl -fsSL`, then executes it with `bash`. The temporary file is removed after execution via a shell `trap`. This relies on HTTPS transport for integrity — no additional checksum verification is performed. Review the script at `BUN_INSTALL_URL` before running in security-sensitive environments.

**Windows install script**: The `install` task downloads the official Bun PowerShell script from `BUN_INSTALL_PS1_URL` using `Invoke-WebRequest`, writes it to a temporary `.ps1` file, executes it, and removes the file in a `finally` block. This relies on HTTPS transport for integrity.
