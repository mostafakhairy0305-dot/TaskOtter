# Buf Taskfile

## What is this Taskfile?

A cross-platform Taskfile for linting, formatting, breaking-change detection,
and code generation from [Protocol Buffer](https://protobuf.dev/) definitions
using [Buf](https://buf.build/), the modern proto toolchain.

Buf is installed globally via the platform package manager — Homebrew on macOS,
a direct binary download to `/usr/local/bin` on Linux, and Scoop on Windows. The
install task is skipped automatically when Buf is already present in PATH.

## Usage

### Standalone

```sh
task -t taskfiles/buf/Taskfile.yml install
task -t taskfiles/buf/Taskfile.yml lint
task -t taskfiles/buf/Taskfile.yml format
task -t taskfiles/buf/Taskfile.yml version
```

Lint a specific proto directory:

```sh
task -t taskfiles/buf/Taskfile.yml lint INPUT=api/v1
```

Check for breaking changes against a branch:

```sh
task -t taskfiles/buf/Taskfile.yml breaking AGAINST=.git#branch=main
```

Pass buf flags directly after `--`:

```sh
task -t taskfiles/buf/Taskfile.yml lint -- api/v1 --error-format json
```

### Included

```yaml
includes:
  buf:
    taskfile: taskfiles/buf/Taskfile.yml
    vars:
      CONFIG_OVERRIDE: "{{.CONFIG}}"
      INPUT_OVERRIDE: "{{.INPUT}}"
      AGAINST_OVERRIDE: "{{.AGAINST}}"
      EXTRA_ARGS_OVERRIDE: "{{.EXTRA_ARGS}}"
```

Then run:

```sh
task buf:lint
task buf:format
task buf:breaking AGAINST=.git#branch=main
task buf:generate INPUT=api/v1
```

## Public Tasks

| Task            | Description                                              | Key variables                  |
| --------------- | -------------------------------------------------------- | ------------------------------ |
| `breaking`      | Check proto files for breaking changes against AGAINST   | `INPUT`, `AGAINST`, `EXTRA_ARGS` |
| `format`        | Check proto file formatting with Buf                     | `INPUT`, `EXTRA_ARGS`          |
| `format:write`  | Format proto files in place with Buf                     | `INPUT`, `EXTRA_ARGS`          |
| `generate`      | Generate code from proto files with Buf                  | `INPUT`, `EXTRA_ARGS`          |
| `install`       | Install Buf on the current operating system              | none                           |
| `install:undo`  | Remove Buf from the current operating system             | none                           |
| `lint`          | Lint proto files with Buf                                | `INPUT`, `CONFIG`, `EXTRA_ARGS` |
| `upgrade`       | Upgrade Buf to the latest release                        | `BUF_VERSION` (Linux only)     |
| `version`       | Show the installed Buf version                           | none                           |

## Variables

| Variable      | Default              | Description                                              |
| ------------- | -------------------- | -------------------------------------------------------- |
| `AGAINST`     | `.git#branch=main`   | Baseline for `breaking`: a git ref, Buf module, or path |
| `BUF_VERSION` | `1.47.2`             | Buf release to download on Linux                        |
| `CONFIG`      | empty                | Path to a `buf.yaml` config file passed via `--config`  |
| `EXTRA_ARGS`  | empty                | Extra arguments appended when `CLI_ARGS` is not provided |
| `INPUT`       | `.`                  | Proto source directory or Buf module passed to buf       |

## Notes

On macOS, Buf is installed from the Homebrew tap `bufbuild/buf` (`brew install
bufbuild/buf/buf`). The `BUF_VERSION` variable is ignored on macOS and Windows
— the package manager controls the installed version.

On Linux, only `x86_64` and `aarch64` architectures are supported via the
direct binary download. Other architectures require a manual installation; see
the [Buf installation docs](https://buf.build/docs/installation).

The `generate` task requires a `buf.gen.yaml` file in the working tree. See the
[buf generate docs](https://buf.build/docs/generate/tutorial) for configuration
details.
