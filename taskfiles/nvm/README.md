# NVM Taskfile Public Tasks

## What is NVM?

NVM (Node Version Manager) is a command-line tool that lets you install and switch between multiple versions of Node.js on the same machine. It is useful when different projects require different Node.js versions, or when you need to test against a specific version without affecting your system-wide Node.js installation.

There are two separate implementations:

- **[nvm-sh](https://github.com/nvm-sh/nvm)** — the original implementation for Linux, macOS, and WSL. It is sourced into your shell session and manages Node.js versions under `~/.nvm`.
- **[nvm-windows](https://github.com/coreybutler/nvm-windows)** — a separate project for Windows with a similar interface but a different installation mechanism.

---

This document describes the public tasks exposed by the NVM Taskfile.

The Taskfile provides one cross-platform interface for managing NVM and Node.js versions.

Windows uses `nvm-windows`.

Linux, macOS, and WSL use `nvm-sh`.

---

## Auto-install behaviour

Every task that requires NVM automatically installs it first if it is not already present — you do not need to run `task install` manually before using any other task.

Every task that requires a specific Node.js version automatically installs it first if it is not already present — you do not need to run `task node:install` manually before using `task node:use`.

Installs are **idempotent**: each internal install task has a `status` check that exits early when the target is already installed, so running any task multiple times is safe and only does work when something is actually missing.

| Task | Auto-installs |
|---|---|
| `version`, `ls`, `node:version`, `node:install`, `node:uninstall` | NVM (if missing) |
| `node:use` | NVM (if missing) → Node.js version (if missing) → activates |

---

## Public Tasks

| Task | Aliases | Variables | Description |
|---|---|---|---|
| `install` | — | — | Install NVM for the current operating system. |
| `install:undo` | `uninstall` | — | Remove NVM from the current operating system. |
| `version` | — | — | Show the installed NVM version. Auto-installs NVM if missing. |
| `node:install` | `node:uninstall:undo` | Optional `VERSION` | Install a Node.js version. If `VERSION` is omitted, install latest LTS. Auto-installs NVM if missing. |
| `node:uninstall` | `node:install:undo` | Required `VERSION` | Uninstall a Node.js version managed by NVM. Auto-installs NVM if missing. |
| `node:use` | — | Optional `VERSION` | Install (if needed) and activate a Node.js version. If `VERSION` is omitted, use latest LTS. Auto-installs NVM and the Node.js version if missing. |
| `ls` | `list` | — | List Node.js versions installed through NVM. Auto-installs NVM if missing. |
| `node:version` | `node:current`, `node:active` | — | Show the active Node.js and npm versions. Auto-installs NVM if missing. |

---

## Install NVM

Install the correct NVM implementation for the current platform.

```bash
task install
```

All other tasks call this automatically, so this is only needed if you want to install NVM without doing anything else yet.

---

## Security Notes

**Unix install script**: The `install` task fetches and pipes the nvm-sh install script directly into `bash` (`curl | bash`). This is the method recommended by the nvm-sh project. It relies on HTTPS transport for integrity — no additional checksum verification is performed. Review the script at the pinned `NVM_SH_INSTALL_URL` before running in security-sensitive environments.

**Windows installer**: The `install` task downloads a pinned `nvm-setup.zip` from the nvm-windows GitHub release. The version is fixed via `NVM_WINDOWS_VERSION` in `Taskfile.yml`. Update both `NVM_WINDOWS_VERSION` and `NVM_SH_VERSION` when upgrading.
