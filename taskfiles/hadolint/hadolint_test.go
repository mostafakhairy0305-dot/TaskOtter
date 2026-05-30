package hadolint_test

import (
	"runtime"
	"testing"

	"github.com/mostafakhairy0305-dot/TaskOtter/internal/tasktest"
)

var publicTasks = []string{
	"install",
	"install:undo",
	"lint",
	"upgrade",
	"version",
}

var publicVars = []string{
	"CONFIG",
	"DOCKERFILE",
	"EXTRA_ARGS",
}

func TestTaskfileModuleContract(t *testing.T) {
	tasktest.AssertModule(t, "hadolint", publicTasks, publicVars)
}

func TestRepresentativeDryRuns(t *testing.T) {
	tasktest.AssertDryRunContains(t, "hadolint",
		[]string{"lint"},
		"hadolint",
		"Dockerfile",
	)

	tasktest.AssertDryRunContains(t, "hadolint",
		[]string{"version"},
		"hadolint",
		"--version",
	)
}

func TestInstallDryRunUsesPlatformPackageManager(t *testing.T) {
	switch runtime.GOOS {
	case "darwin":
		tasktest.AssertDryRunContains(t, "hadolint",
			[]string{"install"},
			"brew",
			"hadolint",
		)
	case "linux":
		tasktest.AssertDryRunContains(t, "hadolint",
			[]string{"install"},
			"hadolint",
		)
	default:
		t.Skip("install dry-run is covered on macOS and Linux")
	}
}
