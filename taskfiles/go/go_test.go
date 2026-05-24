package go_test

import (
	"os/exec"
	"runtime"
	"testing"

	"github.com/mostafakhairy0305-dot/TaskOtter/taskfiles/internal/tasktest"
)

var publicTasks = []string{
	"install",
	"upgrade",
	"verify",
	"version",
	"which",
}

var publicVars = []string{
	"GO_BIN_UNIX",
	"GO_CMD_UNIX",
	"GO_DOWNLOAD_BASE_URL",
	"GO_ROOT_UNIX",
	"GO_VERSION_URL",
	"INSTALL_DIR_UNIX",
}

func goAvailable() bool {
	_, err := exec.LookPath("go")
	return err == nil
}

func TestTaskfileModuleContract(t *testing.T) {
	tasktest.AssertModule(t, "go", publicTasks, publicVars)
}

func TestVersionDryRun(t *testing.T) {
	if !goAvailable() {
		t.Skip("go is not installed")
	}

	tasktest.AssertDryRunContains(t, "go", []string{"version"}, "go version")
}

func TestUpgradeDryRun(t *testing.T) {
	switch runtime.GOOS {
	case "darwin":
		if _, err := exec.LookPath("brew"); err != nil {
			t.Skip("Homebrew is not installed")
		}
		tasktest.AssertDryRunContains(t, "go", []string{"upgrade"}, "brew upgrade go")
	case "linux":
		tasktest.AssertDryRunContains(t, "go", []string{"upgrade"},
			"https://go.dev/VERSION?m=text",
			"sudo tar",
		)
	default:
		t.Skip("upgrade dry-run is covered on macOS and Linux")
	}
}
