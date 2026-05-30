package yamllint_test

import (
	"os/exec"
	"runtime"
	"testing"

	"github.com/mostafakhairy0305-dot/TaskOtter/internal/tasktest"
)

var publicTasks = []string{
	"ci",
	"config:init",
	"install",
	"install:undo",
	"lint",
	"lint:fix",
	"upgrade",
	"version",
}

var publicVars = []string{
	"CONFIG",
	"EXTRA_ARGS",
	"TARGETS",
	"UV_LOAD",
}

func yamllintAvailable() bool {
	_, err := exec.LookPath("yamllint")
	return err == nil
}

func TestTaskfileModuleContract(t *testing.T) {
	tasktest.AssertModule(t, "yamllint", publicTasks, publicVars)
}

func TestInstallDryRun(t *testing.T) {
	if yamllintAvailable() {
		t.Skip("yamllint already installed; status check short-circuits install body")
	}

	switch runtime.GOOS {
	case "darwin", "linux":
		tasktest.AssertDryRunContains(t, "yamllint", []string{"install"}, "uv tool install yamllint")
	default:
		t.Skip("install dry-run is covered on macOS and Linux")
	}
}

func TestVersionDryRun(t *testing.T) {
	if !yamllintAvailable() {
		t.Skip("yamllint is not installed")
	}

	tasktest.AssertDryRunContains(t, "yamllint", []string{"version"}, "yamllint --version")
}

func TestLintDryRun(t *testing.T) {
	tasktest.AssertDryRunContains(t, "yamllint", []string{"lint"},
		"yamllint",
		".",
	)
}

func TestCiDryRun(t *testing.T) {
	tasktest.AssertDryRunContains(t, "yamllint", []string{"ci"}, "yamllint --strict")
}

func TestConfigInitDryRun(t *testing.T) {
	tasktest.AssertDryRunContains(t, "yamllint", []string{"config:init"}, "extends: default")
}
