package shellcheck_test

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
	"EXTRA_ARGS",
	"TARGETS",
}

func TestTaskfileModuleContract(t *testing.T) {
	tasktest.AssertModule(t, "shellcheck", publicTasks, publicVars)
}

func TestRepresentativeDryRuns(t *testing.T) {
	tasktest.AssertDryRunContains(t, "shellcheck",
		[]string{"lint"},
		"shellcheck",
	)

	tasktest.AssertDryRunContains(t, "shellcheck",
		[]string{"version"},
		"shellcheck",
		"--version",
	)
}

func TestInstallDryRunUsesPlatformPackageManager(t *testing.T) {
	switch runtime.GOOS {
	case "darwin":
		tasktest.AssertDryRunContains(t, "shellcheck",
			[]string{"install"},
			"shellcheck",
		)
	case "linux":
		tasktest.AssertDryRunContains(t, "shellcheck",
			[]string{"install"},
			"shellcheck",
		)
	default:
		t.Skip("install dry-run is covered on macOS and Linux")
	}
}

func TestUpgradeDryRunUsesPlatformPackageManager(t *testing.T) {
	switch runtime.GOOS {
	case "darwin":
		tasktest.AssertDryRunContains(t, "shellcheck",
			[]string{"upgrade"},
			"brew",
			"shellcheck",
		)
	case "linux":
		tasktest.AssertDryRunContains(t, "shellcheck",
			[]string{"upgrade"},
			"shellcheck",
		)
	default:
		t.Skip("upgrade dry-run is covered on macOS and Linux")
	}
}
