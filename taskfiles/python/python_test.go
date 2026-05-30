package python_test

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
	"run",
	"upgrade",
	"verify",
	"version",
	"venv",
}

var publicVars = []string{
	"ARGS",
	"EXTRA_ARGS",
	"FILE",
	"REQUIREMENTS",
	"VENV",
}

func pythonAvailable() bool {
	_, err := exec.LookPath("python3")
	if err == nil {
		return true
	}
	_, err = exec.LookPath("python")
	return err == nil
}

func TestTaskfileModuleContract(t *testing.T) {
	tasktest.AssertModule(t, "python", publicTasks, publicVars)
}

func TestInstallDryRun(t *testing.T) {
	if pythonAvailable() {
		t.Skip("python already installed; status check short-circuits install body")
	}

	switch runtime.GOOS {
	case "darwin":
		if _, err := exec.LookPath("brew"); err != nil {
			t.Skip("Homebrew is not installed")
		}
		tasktest.AssertDryRunContains(t, "python", []string{"install"}, "brew install python3")
	case "linux":
		tasktest.AssertDryRunContains(t, "python", []string{"install"}, "python3-pip")
	default:
		t.Skip("install dry-run is covered on macOS and Linux")
	}
}

func TestVersionDryRun(t *testing.T) {
	if !pythonAvailable() {
		t.Skip("python is not installed")
	}

	switch runtime.GOOS {
	case "darwin", "linux":
		tasktest.AssertDryRunContains(t, "python", []string{"version"}, "python3 --version")
	default:
		tasktest.AssertDryRunContains(t, "python", []string{"version"}, "python --version")
	}
}

func TestVenvDryRun(t *testing.T) {
	switch runtime.GOOS {
	case "darwin", "linux":
		tasktest.AssertDryRunContains(t, "python", []string{"venv"}, "python3 -m venv")
	default:
		tasktest.AssertDryRunContains(t, "python", []string{"venv"}, "python -m venv")
	}
}

func TestPipInstallDryRun(t *testing.T) {
	switch runtime.GOOS {
	case "darwin", "linux":
		tasktest.AssertDryRunContains(t, "python", []string{"pip:install"}, "pip3 install -r")
	default:
		tasktest.AssertDryRunContains(t, "python", []string{"pip:install"}, "pip install -r")
	}
}

func TestRunDryRun(t *testing.T) {
	switch runtime.GOOS {
	case "darwin", "linux":
		tasktest.AssertDryRunContains(t, "python", []string{"run", "FILE=hello.py"},
			"python3",
			"hello.py",
		)
	default:
		tasktest.AssertDryRunContains(t, "python", []string{"run", "FILE=hello.py"},
			"python",
			"hello.py",
		)
	}
}
