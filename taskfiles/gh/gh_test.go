package gh_test

import (
	"os/exec"
	"runtime"
	"testing"

	"github.com/mostafakhairy0305-dot/TaskOtter/taskfiles/internal/tasktest"
)

// expectedPublicTasks is the canonical list of public gh Taskfile tasks.
// It must stay in sync with the tasks: block in Taskfile.yml.
var expectedPublicTasks = []string{
	"alias:delete",
	"alias:list",
	"alias:set",
	"api:delete:danger",
	"api:get",
	"api:patch",
	"api:post",
	"auth:login",
	"auth:login:ssh",
	"auth:login:web",
	"auth:logout",
	"auth:refresh",
	"auth:setup-git",
	"auth:status",
	"browse",
	"config:get",
	"config:list",
	"config:set",
	"doctor",
	"extension:install",
	"extension:list",
	"extension:remove",
	"extension:upgrade",
	"gist:create",
	"gist:delete:danger",
	"gist:list",
	"gist:view",
	"help",
	"install",
	"issue:assign",
	"issue:close",
	"issue:comment",
	"issue:create",
	"issue:label",
	"issue:list",
	"issue:reopen",
	"issue:view",
	"open",
	"pr:checkout",
	"pr:close",
	"pr:comment",
	"pr:create",
	"pr:diff",
	"pr:list",
	"pr:merge",
	"pr:ready",
	"pr:review",
	"pr:status",
	"pr:view",
	"project:create",
	"project:list",
	"project:view",
	"release:create",
	"release:delete:danger",
	"release:download",
	"release:download:all",
	"release:list",
	"release:upload",
	"release:view",
	"repo:archive",
	"repo:clone",
	"repo:create",
	"repo:delete:danger",
	"repo:fork",
	"repo:list",
	"repo:sync",
	"repo:view",
	"run:cancel",
	"run:list",
	"run:logs",
	"run:rerun",
	"run:view",
	"search:issues",
	"search:prs",
	"search:repos",
	"secret:delete:danger",
	"secret:list",
	"secret:set",
	"ssh-key:add",
	"ssh-key:delete:danger",
	"ssh-key:list",
	"upgrade",
	"variable:delete:danger",
	"variable:list",
	"variable:set",
	"verify",
	"version",
	"which",
	"workflow:list",
	"workflow:run",
	"workflow:view",
	"workflow:watch",
}

// expectedVars is the list of required top-level vars with non-empty defaults.
var expectedVars = []string{
	"BASE",
	"CLONE_DIR",
	"DATA",
	"DOWNLOAD_DIR",
	"MERGE_METHOD",
	"PAT_TOKEN",
	"VISIBILITY",
}

// ghAvailable reports whether the gh binary is present on PATH.
func ghAvailable() bool {
	_, err := exec.LookPath("gh")
	return err == nil
}

func TestModule(t *testing.T) {
	tasktest.AssertModule(t, "gh", expectedPublicTasks, expectedVars)
}

// Install dry-run tests: these run the install task for the current platform.
// They succeed because the status check (command -v gh) fails when gh is absent,
// so the install cmds are shown in the dry-run output even without gh present.

func TestInstallMacosBrewDryRun(t *testing.T) {
	if runtime.GOOS != "darwin" {
		t.Skip("macOS-only test")
	}
	tasktest.AssertDryRunContains(t, "gh", []string{"install"},
		"brew install gh",
	)
}

func TestInstallLinuxAptDryRun(t *testing.T) {
	if runtime.GOOS != "linux" {
		t.Skip("Linux-only test")
	}
	if ghAvailable() {
		t.Skip("gh is already installed; install task would be a no-op")
	}
	tasktest.AssertDryRunContains(t, "gh", []string{"install"},
		"apt-get",
		"github-cli.list",
	)
}

func TestInstallLinuxDnfDryRun(t *testing.T) {
	if runtime.GOOS != "linux" {
		t.Skip("Linux-only test")
	}
	if ghAvailable() {
		t.Skip("gh is already installed; install task would be a no-op")
	}
	tasktest.AssertDryRunContains(t, "gh", []string{"install"},
		"gh-cli.repo",
	)
}

// Operation dry-run tests: these require gh to be installed so the deps pass.

func TestAuthLoginDryRun(t *testing.T) {
	if !ghAvailable() {
		t.Skip("gh is not installed")
	}
	tasktest.AssertDryRunContains(t, "gh", []string{"auth:login"},
		"gh auth login",
	)
}

func TestAuthLoginWebDryRun(t *testing.T) {
	if !ghAvailable() {
		t.Skip("gh is not installed")
	}
	tasktest.AssertDryRunContains(t, "gh", []string{"auth:login:web"},
		"gh auth login --web",
	)
}

func TestAuthLoginSshDryRun(t *testing.T) {
	if !ghAvailable() {
		t.Skip("gh is not installed")
	}
	tasktest.AssertDryRunContains(t, "gh", []string{"auth:login:ssh"},
		"gh auth login --git-protocol ssh",
	)
}

func TestPrMergeDryRun(t *testing.T) {
	if !ghAvailable() {
		t.Skip("gh is not installed")
	}
	tasktest.AssertDryRunContains(t, "gh",
		[]string{"pr:merge", "PR_NUMBER=42", "MERGE_METHOD=squash"},
		"gh pr merge 42 --squash",
	)
}

func TestApiGetDryRun(t *testing.T) {
	if !ghAvailable() {
		t.Skip("gh is not installed")
	}
	tasktest.AssertDryRunContains(t, "gh",
		[]string{"api:get", "ENDPOINT=/repos/github/cli"},
		"gh api /repos/github/cli",
	)
}

func TestSearchReposDryRun(t *testing.T) {
	if !ghAvailable() {
		t.Skip("gh is not installed")
	}
	tasktest.AssertDryRunContains(t, "gh",
		[]string{"search:repos", "QUERY=cli"},
		"gh search repos",
	)
}
