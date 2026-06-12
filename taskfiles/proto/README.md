# Proto Taskfile

A [TaskOtter](https://github.com/mostafakhairy0305-dot/TaskOtter) module for generating Go and gRPC source files from [Protocol Buffer](https://protobuf.dev/) definitions using [protoc](https://github.com/protocolbuffers/protobuf).

## What is this Taskfile?

This module provides tasks to generate, install, and manage protoc and the Go protobuf plugins (`protoc-gen-go`, `protoc-gen-go-grpc`). Tools are installed globally and resolved from PATH — protoc via the platform package manager or `/usr/local`, and Go plugins via the global Go bin (GOBIN or GOPATH/bin).

## Usage

### Standalone

```sh
task -t taskfiles/proto/Taskfile.yml install
task -t taskfiles/proto/Taskfile.yml gen
task -t taskfiles/proto/Taskfile.yml version
```

Lint a specific proto directory:

```sh
task -t taskfiles/proto/Taskfile.yml gen PROTO_PATH=api PROTO_PATTERN="v1/*.proto"
```

Remove generated files before regenerating:

```sh
task -t taskfiles/proto/Taskfile.yml ungen
task -t taskfiles/proto/Taskfile.yml gen
```

### Included in your Taskfile

```yaml
includes:
  proto:
    taskfile: taskfiles/proto/Taskfile.yml
    vars:
      PROTO_PATH_OVERRIDE: "{{.PROTO_PATH}}"
      PROTO_PATTERN_OVERRIDE: "{{.PROTO_PATTERN}}"
```

Then run:

```sh
task proto:gen
task proto:install
task proto:version
```

## Public Tasks

| Task | Description |
|---|---|
| `gen` | Generate Go files from proto definitions |
| `install` | Install protoc and Go proto plugins on the current operating system |
| `install:undo` | Remove protoc and Go proto plugins from the current operating system |
| `upgrade` | Upgrade protoc and Go proto plugins |
| `ungen` | Remove generated protobuf (.pb.go) files from the working tree |
| `version` | Show the installed protoc version |

## Variables

| Variable | Default | Description |
|---|---|---|
| `GO_CMD` | resolved from PATH | Go executable used to install protobuf plugins |
| `GLOBAL_GO_BIN` | from `go env` | Global Go bin directory where plugins are installed |
| `PROTO_PATH` | `"."` | Search root and value passed to protoc `--proto_path` |
| `PROTO_PATTERN` | `"*.proto"` | `find -name` pattern for discovering .proto source files |
| `PROTOC_VERSION` | `"34.0"` | Pinned protoc release for Linux binary download |
| `PROTOC_GEN_GO_VERSION` | `"v1.36.5"` | Pinned version of protoc-gen-go |
| `PROTOC_GEN_GO_GRPC_VERSION` | `"v1.5.1"` | Pinned version of protoc-gen-go-grpc |

## Notes

- **macOS** installs protoc via Homebrew (`brew install protobuf`). Homebrew must be installed.
- **Linux** downloads the pinned `PROTOC_VERSION` release into `/usr/local/bin` and `/usr/local/include`. Requires `curl` and `unzip`. Only `x86_64` and `aarch64` are supported.
- **Windows** installs protoc via Scoop (`scoop install protobuf`). Scoop must be installed.
- Go is installed automatically through the shared Go module before installing or upgrading the protobuf plugins.
- Go plugins are installed with `go install` into `GLOBAL_GO_BIN`. Ensure that directory is on your PATH.
- The `gen` task prepends `GLOBAL_GO_BIN` to PATH so protoc can resolve the plugins automatically.
- On macOS and Windows, the package manager controls the protoc version — `PROTOC_VERSION` applies to Linux only.
