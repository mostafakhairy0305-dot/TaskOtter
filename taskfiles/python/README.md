# Python Taskfile Public Tasks

## What is this Taskfile?

A cross-platform Taskfile for installing Python, managing upgrades, and running
common project operations such as creating virtual environments, installing
dependencies, and executing scripts.

macOS uses Homebrew to install `python3`. Linux installs `python3`, `python3-pip`,
and `python3-venv` via apt or dnf. Windows uses winget to install Python 3.

## Usage

### Standalone

```sh
task -t taskfiles/python/Taskfile.yml install
task -t taskfiles/python/Taskfile.yml version
task -t taskfiles/python/Taskfile.yml venv
```

### Included

```yaml
includes:
  python: ./taskfiles/python/Taskfile.yml
```

Then run:

```sh
task python:install
task python:venv
task python:pip:install
```

## Public Tasks

| Task           | Description                                 | Key variables                |
| -------------- | ------------------------------------------- | ---------------------------- |
| `install`      | Install Python on the current OS if missing | none                         |
| `install:undo` | Remove Python from the current OS           | none                         |
| `upgrade`      | Upgrade Python to the latest release        | none                         |
| `version`      | Show the installed Python version           | none                         |
| `verify`       | Show Python and pip versions                | none                         |
| `venv`         | Create a virtual environment                | `VENV`                       |
| `pip:install`  | Install packages from a requirements file   | `REQUIREMENTS`, `EXTRA_ARGS` |
| `run`          | Run a Python script                         | `FILE`, `ARGS`, `EXTRA_ARGS` |

## Variables

| Variable       | Default            | Description                                                      |
| -------------- | ------------------ | ---------------------------------------------------------------- |
| `VENV`         | `.venv`            | Virtual environment directory used by `venv`                     |
| `REQUIREMENTS` | `requirements.txt` | Requirements file used by `pip:install`                          |
| `FILE`         | _(empty)_          | Script path; required by `run`                                   |
| `ARGS`         | _(empty)_          | Positional arguments forwarded to the script in `run`            |
| `EXTRA_ARGS`   | _(empty)_          | Extra flags forwarded to `pip install` or the Python interpreter |

## Notes

**macOS:** Requires Homebrew. `brew install python3` installs the latest stable
Python 3 formula alongside the system Python; both coexist safely.

**Linux:** apt-based systems receive `python3`, `python3-pip`, and `python3-venv`.
dnf-based systems receive `python3` and `python3-pip` (`python3-venv` is bundled
with the interpreter on most dnf distributions).

**Windows:** winget installs the official Python 3 release from python.org.
Restart your terminal after installation for `python` and `pip` to be available
in PATH.

**`install:undo`** is supported on macOS (Homebrew) and Windows (winget) only.
On Linux, removing system Python via a task is unsafe and may break OS tooling,
so this task exits non-zero with an explanatory message instead. Use your
distribution package manager directly to remove Python on Linux.
