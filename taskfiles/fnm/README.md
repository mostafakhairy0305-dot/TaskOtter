# FNM Taskfile Public Tasks

## What is fnm?

fnm (Fast Node Manager) is a cross-platform Node.js version manager written in Rust. It lets you install and switch between multiple Node.js versions on the same machine, and is compatible with `.nvmrc` and `.node-version` files used by nvm.

Unlike nvm, fnm is a single binary that works natively on Linux, macOS, and Windows — there is no separate implementation per platform. It is significantly faster than nvm due to being compiled ahead-of-time.

Key differences from nvm:

- **Single binary** — no shell function to source; fnm is invoked directly.
- **Shell integration via `eval`** — run `eval "$(fnm env)"` to activate the selected Node.js version in your current shell session.
- **Cross-platform** — one tool for Linux, macOS, and Windows.

---

This document describes the public tasks exposed by the FNM Taskfile.

The Taskfile provides one cross-platform interface for managing fnm and Node.js versions.

Linux and macOS use the official fnm install script.

Windows uses winget.

---

## Auto-install behaviour

Every task that requires fnm automatically installs it first if it is not already present — you do not need to run `task install` manually before using any other task.

Every task that requires a specific Node.js version automatically installs it first if it is not already present — you do not need to run `task node:install` manually before using `task node:use`.

Installs are **idempotent**: each internal install task has a `status` check that exits early when the target is already installed, so running any task multiple times is safe and only does work when something is actually missing.

| Task | Auto-installs |
|---|---|
| `version`, `ls`, `node:version`, `node:install`, `node:uninstall` | fnm (if missing) |
| `node:use` | fnm (if missing) → Node.js version (if missing) → activates |

---

## Public Tasks

| Task | Aliases | Variables | Description |
|---|---|---|---|
| `install` | — | — | Install fnm for the current operating system. |
| `install:undo` | `uninstall` | — | Remove fnm from the current operating system. |
| `version` | — | — | Show the installed fnm version. Auto-installs fnm if missing. |
| `node:install` | `node:uninstall:undo` | Optional `VERSION` | Install a Node.js version. If `VERSION` is omitted, install latest LTS. Auto-installs fnm if missing. |
| `node:uninstall` | `node:install:undo` | Required `VERSION` | Uninstall a Node.js version managed by fnm. Auto-installs fnm if missing. |
| `node:use` | — | Optional `VERSION` | Install (if needed) and activate a Node.js version. If `VERSION` is omitted, use latest LTS. Auto-installs fnm and the Node.js version if missing. |
| `ls` | `list` | — | List Node.js versions installed through fnm. Auto-installs fnm if missing. |
| `node:version` | `node:current`, `node:active` | — | Show the active Node.js and npm versions. Auto-installs fnm if missing. |

---

## Install fnm

Install fnm for the current platform.

```bash
task install
```

All other tasks call this automatically, so this is only needed if you want to install fnm without doing anything else yet.

---

## Shell activation

After `task node:use`, Node.js is set as the fnm default for new shells. To activate it in the **current** shell session, run:

```bash
eval "$(fnm env)"
```

Add this to your shell profile (`~/.bashrc`, `~/.zshrc`, etc.) for automatic activation on every new shell.

---

## Security Notes

**Unix install script**: The `install` task fetches and pipes the fnm install script directly into `bash` (`curl | bash`). This is the method recommended by the fnm project. It relies on HTTPS transport for integrity — no additional checksum verification is performed. Review the script at `FNM_INSTALL_URL` before running in security-sensitive environments.

**Windows installer**: The `install` task uses `winget install Schniz.fnm`. winget verifies package signatures from the Microsoft Store source. No manual checksum step is required.
