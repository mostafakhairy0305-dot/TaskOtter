package staticcheck_test

import (
	"runtime"
	"testing"

	"github.com/mostafakhairy0305-dot/TaskOtter/internal/tasktest"
)

var publicTasks = []string{
	"install",
	"run",
	"version",
}

var publicVars = []string{
	"EXTRA_ARGS",
	"STATICCHECK_BIN_UNIX",
	"STATICCHECK_BIN_WINDOWS",
	"STATICCHECK_INSTALL_DIR",
	"STATICCHECK_RELEASE_BASE_URL",
	"STATICCHECK_VERSION",
	"TARGETS",
	"TOOLBIN",
}

func TestTaskfileModuleContract(t *testing.T) {
	tasktest.AssertModule(t, "staticcheck", publicTasks, publicVars)
}

func TestRepresentativeDryRuns(t *testing.T) {
	tasktest.AssertDryRunContains(t, "staticcheck",
		[]string{"run", "--", "./cmd/..."},
		"staticcheck",
		"./cmd/...",
	)

	tasktest.AssertDryRunContains(t, "staticcheck",
		[]string{"version"},
		"-version",
	)
}

func TestInstallDryRunUsesPlatformArchive(t *testing.T) {
	switch runtime.GOOS {
	case "darwin":
		tasktest.AssertDryRunContains(t, "staticcheck",
			[]string{"install"},
			"staticcheck_darwin_",
			".tar.gz",
		)
	case "linux":
		tasktest.AssertDryRunContains(t, "staticcheck",
			[]string{"install"},
			"staticcheck_linux_",
			".tar.gz",
		)
	default:
		t.Skip("archive dry-run is covered on macOS and Linux")
	}
}
