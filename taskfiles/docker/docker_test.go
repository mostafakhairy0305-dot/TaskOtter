package docker_test

import (
	"os/exec"
	"runtime"
	"testing"

	"github.com/mostafakhairy0305-dot/TaskOtter/internal/tasktest"
)

var publicTasks = []string{
	"build",
	"images",
	"install",
	"install:undo",
	"prune",
	"prune:all",
	"ps",
	"ps:all",
	"pull",
	"stop:all",
	"upgrade",
	"verify",
	"version",
}

var publicVars = []string{
	"CONTEXT",
	"EXTRA_ARGS",
	"FILE",
	"IMAGE",
}

func TestTaskfileModuleContract(t *testing.T) {
	tasktest.AssertModule(t, "docker", publicTasks, publicVars)
}

func TestInstallDryRun(t *testing.T) {
	if _, err := exec.LookPath("docker"); err == nil {
		t.Skip("docker already installed; status check short-circuits install body")
	}

	switch runtime.GOOS {
	case "darwin":
		if _, err := exec.LookPath("brew"); err != nil {
			t.Skip("Homebrew is not installed")
		}
		tasktest.AssertDryRunContains(t, "docker", []string{"install"}, "brew install --cask docker")
	case "linux":
		tasktest.AssertDryRunContains(t, "docker", []string{"install"},
			"https://get.docker.com",
			"usermod -aG docker",
		)
	default:
		t.Skip("install dry-run is covered on macOS and Linux")
	}
}

func TestVersionDryRun(t *testing.T) {
	if _, err := exec.LookPath("docker"); err != nil {
		t.Skip("docker is not installed")
	}

	tasktest.AssertDryRunContains(t, "docker", []string{"version"}, "docker version")
}

func TestPsDryRun(t *testing.T) {
	tasktest.AssertDryRunContains(t, "docker", []string{"ps"}, "docker ps")
}

func TestPsAllDryRun(t *testing.T) {
	tasktest.AssertDryRunContains(t, "docker", []string{"ps:all"}, "docker ps -a")
}

func TestImagesDryRun(t *testing.T) {
	tasktest.AssertDryRunContains(t, "docker", []string{"images"}, "docker images")
}

func TestBuildDryRun(t *testing.T) {
	tasktest.AssertDryRunContains(t, "docker", []string{"build", "IMAGE=myapp:latest"},
		"docker build",
		"-t",
		"myapp:latest",
	)
}

func TestPullDryRun(t *testing.T) {
	tasktest.AssertDryRunContains(t, "docker", []string{"pull", "IMAGE=ubuntu:latest"},
		"docker pull",
		"ubuntu:latest",
	)
}

func TestPruneDryRun(t *testing.T) {
	tasktest.AssertDryRunContains(t, "docker", []string{"prune"}, "docker system prune")
}
