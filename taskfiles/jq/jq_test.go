package jq_test

import (
	"os/exec"
	"runtime"
	"testing"

	"github.com/mostafakhairy0305-dot/TaskOtter/internal/tasktest"
)

var publicTasks = []string{
	"install",
	"install:undo",
	"upgrade",
	"version",
}

func jqAvailable() bool {
	_, err := exec.LookPath("jq")
	return err == nil
}

var publicVars = []string{
	"VERSION",
}

func TestTaskfileModuleContract(t *testing.T) {
	tasktest.AssertModule(t, "jq", publicTasks, publicVars)
}

func TestRepresentativeDryRuns(t *testing.T) {
	tasktest.AssertDryRunContains(t, "jq",
		[]string{"version"},
		"jq",
		"--version",
	)
}

func TestInstallDryRunUsesPlatformPackageManager(t *testing.T) {
	if jqAvailable() {
		t.Skip("jq is already installed; install task would be a no-op")
	}
	switch runtime.GOOS {
	case "darwin":
		tasktest.AssertDryRunContains(t, "jq",
			[]string{"install"},
			"brew",
			"jq",
		)
	case "linux":
		tasktest.AssertDryRunContains(t, "jq",
			[]string{"install"},
			"jq",
		)
	default:
		t.Skip("install dry-run is covered on macOS and Linux")
	}
}

func TestUninstallAliasDryRun(t *testing.T) {
	tasktest.DryRun(t, "jq", "--yes", "uninstall")
}
