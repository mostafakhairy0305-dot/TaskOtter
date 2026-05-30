package ansible_test

import (
	"os/exec"
	"runtime"
	"testing"

	"github.com/mostafakhairy0305-dot/TaskOtter/internal/tasktest"
)

var publicTasks = []string{
	"galaxy:install",
	"install",
	"install:undo",
	"lint",
	"list:hosts",
	"ping",
	"run",
	"syntax:check",
	"upgrade",
	"vault:decrypt",
	"vault:encrypt",
	"version",
}

var publicVars = []string{
	"EXTRA_ARGS",
	"FILE",
	"INVENTORY",
	"PATTERN",
	"PLAYBOOK",
	"REQUIREMENTS",
	"TARGETS",
	"UV_LOAD",
}

func ansibleAvailable() bool {
	_, err := exec.LookPath("ansible")
	return err == nil
}

func TestTaskfileModuleContract(t *testing.T) {
	tasktest.AssertModule(t, "ansible", publicTasks, publicVars)
}

func TestInstallDryRun(t *testing.T) {
	if ansibleAvailable() {
		t.Skip("ansible already installed; status check short-circuits install body")
	}

	switch runtime.GOOS {
	case "darwin", "linux":
		tasktest.AssertDryRunContains(t, "ansible", []string{"install"}, "uv tool install ansible")
	default:
		t.Skip("install dry-run is covered on macOS and Linux")
	}
}

func TestVersionDryRun(t *testing.T) {
	if !ansibleAvailable() {
		t.Skip("ansible is not installed")
	}

	tasktest.AssertDryRunContains(t, "ansible", []string{"version"}, "ansible --version")
}

func TestLintDryRun(t *testing.T) {
	switch runtime.GOOS {
	case "darwin", "linux":
		tasktest.AssertDryRunContains(t, "ansible", []string{"lint"},
			"ansible-lint",
			".",
		)
	default:
		t.Skip("lint dry-run is covered on macOS and Linux")
	}
}

func TestRunDryRun(t *testing.T) {
	switch runtime.GOOS {
	case "darwin", "linux":
		tasktest.AssertDryRunContains(t, "ansible", []string{"run", "PLAYBOOK=site.yml"},
			"ansible-playbook",
			"site.yml",
		)
	default:
		t.Skip("run dry-run is covered on macOS and Linux")
	}
}

func TestPingDryRun(t *testing.T) {
	switch runtime.GOOS {
	case "darwin", "linux":
		tasktest.AssertDryRunContains(t, "ansible", []string{"ping", "INVENTORY=hosts"},
			"ansible",
			"ping",
		)
	default:
		t.Skip("ping dry-run is covered on macOS and Linux")
	}
}
