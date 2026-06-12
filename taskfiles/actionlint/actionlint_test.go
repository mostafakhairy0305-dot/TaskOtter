package actionlint_test

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
	"ACTIONLINT_VERSION",
	"EXTRA_ARGS",
	"TARGETS",
}

func TestTaskfileModuleContract(t *testing.T) {
	tasktest.AssertModule(t, "actionlint", publicTasks, publicVars)
}

func TestRepresentativeDryRuns(t *testing.T) {
	tasktest.AssertDryRunContains(t, "actionlint",
		[]string{"lint"},
		"actionlint",
	)

	tasktest.AssertDryRunContains(t, "actionlint",
		[]string{"version"},
		"actionlint",
		"--version",
	)
}

func TestInstallDryRunUsesPlatformPackageManager(t *testing.T) {
	switch runtime.GOOS {
	case "darwin":
		tasktest.AssertDryRunContains(t, "actionlint",
			[]string{"install"},
			"brew",
			"actionlint",
		)
	case "linux":
		tasktest.AssertDryRunContains(t, "actionlint",
			[]string{"install"},
			"curl",
			"actionlint_",
		)
	default:
		t.Skip("install dry-run is covered on macOS and Linux")
	}
}

func TestUpgradeDryRunUsesPlatformPackageManager(t *testing.T) {
	switch runtime.GOOS {
	case "darwin":
		tasktest.AssertDryRunContains(t, "actionlint",
			[]string{"upgrade"},
			"brew",
			"actionlint",
		)
	case "linux":
		tasktest.AssertDryRunContains(t, "actionlint",
			[]string{"upgrade"},
			"curl",
			"actionlint_",
		)
	default:
		t.Skip("upgrade dry-run is covered on macOS and Linux")
	}
}
