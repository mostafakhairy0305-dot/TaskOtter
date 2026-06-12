package zizmor_test

import (
	"runtime"
	"testing"

	"github.com/mostafakhairy0305-dot/TaskOtter/internal/tasktest"
)

var publicTasks = []string{
	"audit",
	"install",
	"install:undo",
	"upgrade",
	"version",
}

var publicVars = []string{
	"EXTRA_ARGS",
	"TARGETS",
	"ZIZMOR_VERSION",
}

func TestTaskfileModuleContract(t *testing.T) {
	tasktest.AssertModule(t, "zizmor", publicTasks, publicVars)
}

func TestRepresentativeDryRuns(t *testing.T) {
	tasktest.AssertDryRunContains(t, "zizmor",
		[]string{"audit"},
		"zizmor",
	)

	tasktest.AssertDryRunContains(t, "zizmor",
		[]string{"version"},
		"zizmor",
		"--version",
	)
}

func TestInstallDryRunDownloadsBinary(t *testing.T) {
	switch runtime.GOOS {
	case "darwin":
		tasktest.AssertInstallDryRun(t, "zizmor", "zizmor", "curl", "apple-darwin")
	case "linux":
		tasktest.AssertInstallDryRun(t, "zizmor", "zizmor", "curl", "linux-gnu")
	default:
		t.Skip("install dry-run is covered on macOS and Linux")
	}
}

func TestUpgradeDryRunDownloadsBinary(t *testing.T) {
	switch runtime.GOOS {
	case "darwin", "linux":
		tasktest.AssertDryRunContains(t, "zizmor",
			[]string{"upgrade"},
			"curl",
			"zizmor",
		)
	default:
		t.Skip("upgrade dry-run is covered on macOS and Linux")
	}
}
