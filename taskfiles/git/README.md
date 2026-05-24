# git — Git Taskfile

## What is this Taskfile?

A production-ready, cross-platform Taskfile for everyday git operations and GitHub-integrated
workflows. It wraps the `git` CLI with consistent defaults and integrates with the
[GitHub CLI (`gh`)](https://cli.github.com) for authentication, pull requests, and releases.

Run `task auth:setup` once to configure git credential delegation to gh — after that,
`clone`, `push`, `pull`, `pr:create`, and `release:create` all authenticate transparently.

## Usage

### Standalone

```sh
task -t taskfiles/git/Taskfile.yml auth:setup
task -t taskfiles/git/Taskfile.yml clone OWNER=github REPO=cli
task -t taskfiles/git/Taskfile.yml commit COMMIT_MSG="feat: add login page"
task -t taskfiles/git/Taskfile.yml pr:create TITLE="feat: login page" BASE=main
```

### Included (recommended)

```yaml
# Taskfile.yml
includes:
  git: ./taskfiles/git/Taskfile.yml
```

Then run:

```sh
task git:auth:setup
task git:commit COMMIT_MSG="feat: add feature"
task git:pr:create TITLE="feat: add feature" BASE=main
```

## Public Tasks

| Task | Description | Key variables |
|------|-------------|---------------|
| `auth:setup` | Configure git to use gh as credential helper | — |
| `init` | Initialize a new git repository | `BRANCH` |
| `clone` | Clone a GitHub repository using the GitHub CLI | `REPO`, `OWNER`, `CLONE_DIR` |
| `status` | Show the current working tree status | — |
| `add` | Stage files for the next commit | `FILES` |
| `add:all` | Stage all changes including untracked files | — |
| `commit` | Create a commit from staged changes | `COMMIT_MSG` |
| `commit:amend` | Amend the most recent commit | `COMMIT_MSG` |
| `push` | Push commits to the remote repository | `REMOTE`, `BRANCH` |
| `push:force` | Force-push using --force-with-lease | `REMOTE`, `BRANCH` |
| `pull` | Pull changes from the remote repository | `REMOTE`, `BRANCH` |
| `fetch` | Fetch branches from the remote without merging | `REMOTE` |
| `sync` | Sync the current branch with its GitHub upstream | `BRANCH` |
| `diff` | Show unstaged changes in the working tree | `FILES` |
| `diff:staged` | Show changes staged for the next commit | — |
| `log` | Show the commit history with author and date | `EXTRA_ARGS` |
| `log:graph` | Show commit history as an ASCII branch graph | `EXTRA_ARGS` |
| `branch:list` | List all local and remote branches | — |
| `branch:create` | Create and switch to a new branch from current HEAD | `BRANCH` |
| `branch:switch` | Switch to an existing branch | `BRANCH` |
| `branch:delete` | Delete a local branch | `BRANCH` |
| `branch:rename` | Rename the current branch to a new name | `BRANCH` |
| `tag:list` | List all tags sorted by version descending | — |
| `tag:create` | Create an annotated tag at HEAD | `TAG`, `MESSAGE` |
| `tag:push` | Push a tag or all tags to the remote | `TAG`, `REMOTE` |
| `tag:delete` | Delete a tag locally and from the remote | `TAG`, `REMOTE` |
| `stash` | Stash uncommitted changes in the working tree | `MESSAGE` |
| `stash:pop` | Apply and remove the latest stash entry | `STASH_INDEX` |
| `stash:list` | List all stash entries | — |
| `stash:drop` | Discard a stash entry without applying it | `STASH_INDEX` |
| `reset:soft` | Soft-reset HEAD to a commit, preserving staged changes | `COMMIT` |
| `reset:hard` | Hard-reset HEAD to a commit, discarding all local changes | `COMMIT` |
| `clean` | Remove untracked files and directories from the working tree | — |
| `config:user` | Set the global git user name and email address | `NAME`, `EMAIL` |
| `config:list` | List all git configuration values | — |
| `remote:list` | List all configured remotes and their URLs | — |
| `remote:add` | Add a new remote to the repository | `NAME`, `URL` |
| `remote:remove` | Remove a configured remote from the repository | `NAME` |
| `remote:set-url` | Update the URL of a configured remote | `NAME`, `URL` |
| `pr:create` | Push the current branch and open a pull request on GitHub | `TITLE`, `BASE`, `BODY`, `REMOTE` |
| `pr:open` | Open the current pull request in the browser via the GitHub CLI | — |
| `release:create` | Create a git tag and a GitHub release via the GitHub CLI | `TAG`, `TITLE`, `NOTES`, `REMOTE` |
| `version` | Show the installed git version | — |
| `help` | Show the git built-in help and command list | — |

## Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `REMOTE` | `origin` | Remote name for push, pull, fetch, and tag operations |
| `BASE` | `main` | Base branch for pull requests |
| `MERGE_METHOD` | `merge` | PR merge strategy: `merge`, `squash`, `rebase` |
| `FILES` | `.` | Files or globs for `add` and `diff` |
| `STASH_INDEX` | `0` | Stash entry index for `stash:pop` and `stash:drop` |
| `BRANCH` | _(empty)_ | Branch name |
| `CLONE_DIR` | _(empty)_ | Destination directory for `clone` |
| `COMMIT` | _(empty)_ | Commit ref for `reset:soft` and `reset:hard` |
| `COMMIT_MSG` | _(empty)_ | Commit message |
| `EMAIL` | _(empty)_ | Git user email for `config:user` |
| `NAME` | _(empty)_ | Git user name or remote name |
| `NOTES` | _(empty)_ | Release notes body |
| `OWNER` | _(empty)_ | GitHub user or organisation for `clone` |
| `REPO` | _(empty)_ | Repository name for `clone` |
| `TAG` | _(empty)_ | Tag name |
| `TITLE` | _(empty)_ | PR or release title |
| `BODY` | _(empty)_ | PR description |
| `URL` | _(empty)_ | Remote URL for `remote:add` and `remote:set-url` |
| `MESSAGE` | _(empty)_ | Tag annotation or stash description |
| `EXTRA_ARGS` | _(empty)_ | Extra arguments appended to the underlying command |

## Examples

```sh
# Set up gh as git credential helper (run once)
task -t taskfiles/git/Taskfile.yml auth:setup

# Clone a repository
task -t taskfiles/git/Taskfile.yml clone OWNER=github REPO=cli
task -t taskfiles/git/Taskfile.yml clone OWNER=myorg REPO=private-repo CLONE_DIR=~/src/project

# Stage and commit
task -t taskfiles/git/Taskfile.yml add FILES=src/
task -t taskfiles/git/Taskfile.yml commit COMMIT_MSG="feat: add login page"
task -t taskfiles/git/Taskfile.yml commit:amend COMMIT_MSG="feat: add login page with tests"

# Push and sync
task -t taskfiles/git/Taskfile.yml push
task -t taskfiles/git/Taskfile.yml push:force REMOTE=origin
task -t taskfiles/git/Taskfile.yml pull REMOTE=origin BRANCH=main
task -t taskfiles/git/Taskfile.yml sync BRANCH=main

# Branches
task -t taskfiles/git/Taskfile.yml branch:create BRANCH=feature/my-feature
task -t taskfiles/git/Taskfile.yml branch:switch BRANCH=main
task -t taskfiles/git/Taskfile.yml branch:delete BRANCH=feature/old-feature
task -t taskfiles/git/Taskfile.yml branch:rename BRANCH=new-name

# Pull requests (push + gh pr create in one step)
task -t taskfiles/git/Taskfile.yml pr:create TITLE="feat: login page" BASE=main
task -t taskfiles/git/Taskfile.yml pr:create TITLE="fix: auth bug" BASE=develop BODY="Closes #42"
task -t taskfiles/git/Taskfile.yml pr:open

# Tags and releases
task -t taskfiles/git/Taskfile.yml tag:create TAG=v1.0.0 MESSAGE="Release v1.0.0"
task -t taskfiles/git/Taskfile.yml tag:push TAG=v1.0.0
task -t taskfiles/git/Taskfile.yml release:create TAG=v1.0.0 TITLE="v1.0.0" NOTES="Initial release"
task -t taskfiles/git/Taskfile.yml tag:delete TAG=v0.1.0

# Stash
task -t taskfiles/git/Taskfile.yml stash MESSAGE="WIP: refactor auth"
task -t taskfiles/git/Taskfile.yml stash:pop
task -t taskfiles/git/Taskfile.yml stash:list

# Reset and clean
task -t taskfiles/git/Taskfile.yml reset:soft COMMIT=HEAD~1
task -t taskfiles/git/Taskfile.yml reset:hard COMMIT=HEAD~1
task -t taskfiles/git/Taskfile.yml clean

# Config
task -t taskfiles/git/Taskfile.yml config:user NAME="Ada Lovelace" EMAIL=ada@example.com
task -t taskfiles/git/Taskfile.yml config:list

# Remotes
task -t taskfiles/git/Taskfile.yml remote:list
task -t taskfiles/git/Taskfile.yml remote:add NAME=upstream URL=https://github.com/org/repo.git
task -t taskfiles/git/Taskfile.yml remote:set-url NAME=origin URL=git@github.com:org/repo.git
```
