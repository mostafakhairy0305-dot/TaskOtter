# Docker Taskfile Public Tasks

## What is this Taskfile?

A cross-platform Taskfile for installing Docker, managing upgrades, and running
common container and image operations.

macOS uses Homebrew Cask to install Docker Desktop. Linux uses the official
convenience script from `get.docker.com` (supports apt and dnf), then
automatically adds the current user to the `docker` group so Docker runs
without `sudo` after a re-login. Windows uses winget to install Docker Desktop.

## Usage

### Standalone

```sh
task -t taskfiles/docker/Taskfile.yml install
task -t taskfiles/docker/Taskfile.yml version
task -t taskfiles/docker/Taskfile.yml ps
```

### Included

```yaml
includes:
  docker: ./taskfiles/docker/Taskfile.yml
```

Then run:

```sh
task docker:install
task docker:version
task docker:ps
```

## Public Tasks

| Task | Description | Key variables |
|------|-------------|---------------|
| `install` | Install Docker on the current OS; Linux also adds the user to the docker group | none |
| `install:undo` | Remove Docker from the current OS | none |
| `upgrade` | Upgrade Docker to the latest release | none |
| `version` | Show Docker client and server versions | none |
| `verify` | Verify Docker installation and daemon connectivity | none |
| `ps` | List running containers | none |
| `ps:all` | List all containers including stopped ones | none |
| `stop:all` | Stop all running containers | none |
| `prune` | Remove stopped containers and dangling images | none |
| `prune:all` | Full system prune including volumes | none |
| `images` | List local Docker images | none |
| `build` | Build a Docker image from a Dockerfile | `IMAGE`, `FILE`, `CONTEXT` |
| `pull` | Pull a Docker image from a registry | `IMAGE` |

## Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `IMAGE` | _(empty)_ | Image name and tag used by `build` and `pull` |
| `FILE` | `Dockerfile` | Dockerfile path used by `build` |
| `CONTEXT` | `.` | Build context directory used by `build` |
| `EXTRA_ARGS` | _(empty)_ | Extra flags forwarded to `docker build` or `docker pull` |

## Notes

**Linux:** `install` automatically runs `sudo usermod -aG docker $USER` after
the engine is set up. Log out and back in for the change to take effect. The
step is skipped if the user is already in the `docker` group or is running as
root.

**macOS:** Docker Desktop must be opened at least once after installation to
complete the initial VM and daemon setup. `brew install --cask docker` does not
start Docker automatically.

**`prune:all`** is destructive and irreversible. It removes all stopped
containers, all unused images, unused networks, and all volumes not attached to
a running container. Confirmation is required before it runs.
