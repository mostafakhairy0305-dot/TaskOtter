# Ansible Taskfile Public Tasks

## What is this Taskfile?

A Taskfile for installing Ansible and ansible-lint, linting Ansible YAML files,
running playbooks, testing connectivity, managing Ansible Galaxy dependencies,
and encrypting/decrypting secrets with Ansible Vault.

Ansible and ansible-lint are installed via uv as isolated tools. yamllint is
installed as a dependency because ansible-lint uses it for YAML syntax checking.

> **Note:** Ansible does not support Windows as a control node. All tasks are
> macOS and Linux only.

## Usage

### Standalone

```sh
task -t taskfiles/ansible/Taskfile.yml install
task -t taskfiles/ansible/Taskfile.yml lint
task -t taskfiles/ansible/Taskfile.yml run PLAYBOOK=site.yml INVENTORY=hosts
```

### Included

```yaml
includes:
  ansible: ./taskfiles/ansible/Taskfile.yml
```

Then run:

```sh
task ansible:install
task ansible:lint
task ansible:run PLAYBOOK=site.yml INVENTORY=hosts
```

## Public Tasks

| Task | Description | Key variables |
|------|-------------|---------------|
| `install` | Install Ansible and ansible-lint via uv | none |
| `install:undo` | Remove Ansible and ansible-lint | none |
| `upgrade` | Upgrade Ansible and ansible-lint to the latest release | none |
| `version` | Show Ansible and ansible-lint versions | none |
| `lint` | Lint Ansible YAML files with ansible-lint | `TARGETS`, `EXTRA_ARGS` |
| `syntax:check` | Check playbook syntax without executing | `PLAYBOOK`, `INVENTORY` |
| `run` | Run an Ansible playbook | `PLAYBOOK`, `INVENTORY`, `EXTRA_ARGS` |
| `ping` | Test connectivity to inventory hosts | `INVENTORY`, `PATTERN`, `EXTRA_ARGS` |
| `list:hosts` | List hosts matching PATTERN from INVENTORY | `INVENTORY`, `PATTERN`, `EXTRA_ARGS` |
| `galaxy:install` | Install roles and collections from a requirements file | `REQUIREMENTS`, `EXTRA_ARGS` |
| `vault:encrypt` | Encrypt a file with Ansible Vault | `FILE`, `EXTRA_ARGS` |
| `vault:decrypt` | Decrypt a file with Ansible Vault | `FILE`, `EXTRA_ARGS` |

## Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `PLAYBOOK` | _(empty)_ | Playbook path; required by `run` and `syntax:check` |
| `INVENTORY` | _(empty)_ | Inventory file or directory; required by `ping` and `list:hosts` |
| `PATTERN` | `all` | Host pattern for `ping` and `list:hosts` |
| `TARGETS` | `.` | Files or directories to lint with `lint` |
| `FILE` | _(empty)_ | File path; required by `vault:encrypt` and `vault:decrypt` |
| `REQUIREMENTS` | `requirements.yml` | Requirements file for `galaxy:install` |
| `EXTRA_ARGS` | _(empty)_ | Extra flags forwarded to the underlying Ansible command |
| `UV_LOAD` | `export PATH="$HOME/.local/bin:$PATH"` | Shell snippet that ensures uv-managed binaries are in PATH |

## Notes

**`lint`** uses ansible-lint, which enforces Ansible best practices on top of
yamllint YAML syntax checking. Configure linting rules with an `.ansible-lint`
file in your project root.

**`vault:decrypt`** prompts for confirmation before decrypting to prevent
accidental plaintext exposure. Both vault tasks prompt interactively for the
vault password.

**`galaxy:install`** installs roles under `~/.ansible/roles` and collections
under `~/.ansible/collections` by default. Override with `EXTRA_ARGS` or a
`ansible.cfg` in your project.
