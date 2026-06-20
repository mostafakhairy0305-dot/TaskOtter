package uv_test

import (
	"os/exec"
	"runtime"
	"testing"

	"github.com/mostafakhairy0305-dot/TaskOtter/internal/tasktest"
)

var publicTasks = []string{
	"install",
	"install:undo",
	"pip:install",
	"python:install",
	"run",
	"tool:install",
	"tool:upgrade",
	"upgrade",
	"venv",
	"version",
}

var publicVars = []string{
	"ARGS",
	"EXTRA_ARGS",
	"FILE",
	"PYTHON_VERSION",
	"REQUIREMENTS",
	"TOOL",
	"UV_INSTALL_URL",
	"UV_INSTALL_URL_WINDOWS",
	"UV_LOAD",
	"UV_VERSION",
	"VENV",
}

func uvAvailable() bool {
	_, err := exec.LookPath("uv")
	return err == nil
}

func TestTaskfileModuleContract(t *testing.T) {
	tasktest.AssertModule(t, "uv", publicTasks, publicVars)
}

func TestInstallDryRun(t *testing.T) {
	if uvAvailable() {
		t.Skip("uv already installed; status check short-circuits install body")
	}

	switch runtime.GOOS {
	case "darwin", "linux":
		tasktest.AssertDryRunContains(t, "uv", []string{"install"}, "astral.sh/uv/install.sh")
	default:
		t.Skip("install dry-run is covered on macOS and Linux")
	}
}

func TestVersionDryRun(t *testing.T) {
	if !uvAvailable() {
		t.Skip("uv is not installed")
	}

	tasktest.AssertDryRunContains(t, "uv", []string{"version"}, "uv --version")
}

func TestToolInstallDryRun(t *testing.T) {
	tasktest.AssertDryRunContains(t, "uv", []string{"tool:install", "TOOL=yamllint"},
		"uv tool install",
		"yamllint",
	)
}

func TestVenvDryRun(t *testing.T) {
	tasktest.AssertDryRunContains(t, "uv", []string{"venv"}, "uv venv")
}

func TestPipInstallDryRun(t *testing.T) {
	tasktest.AssertDryRunContains(t, "uv", []string{"pip:install"}, "uv pip install -r")
}

func TestRunDryRun(t *testing.T) {
	tasktest.AssertDryRunContains(t, "uv", []string{"run", "FILE=main.py"},
		"uv run",
		"main.py",
	)
}
