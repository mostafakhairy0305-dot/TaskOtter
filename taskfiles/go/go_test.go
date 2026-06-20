package go_test

import (
	"os/exec"
	"runtime"
	"slices"
	"testing"

	"github.com/mostafakhairy0305-dot/TaskOtter/internal/tasktest"
)

var publicTasks = []string{
	"fmt",
	"fmt:check",
	"gofumpt:fmt",
	"gofumpt:lint",
	"goimports:fmt",
	"goimports:lint",
	"golangci-lint:lint",
	"golangci-lint:lint:fix",
	"gosec:lint",
	"govulncheck:lint",
	"install",
	"install:gofumpt",
	"install:goimports",
	"install:golangci-lint",
	"install:gosec",
	"install:govulncheck",
	"install:undo",
	"lint",
	"lint:fix",
	"upgrade",
	"verify",
	"version",
	"which",
}

var publicVars = []string{
	"GO_BIN_UNIX",
	"GO_CMD_UNIX",
	"GO_DOWNLOAD_BASE_URL",
	"GO_VERSION",
	"GO_ROOT_UNIX",
	"GO_VERSION_URL",
	"GOFUMPT_VERSION",
	"GOIMPORTS_VERSION",
	"GOLANGCI_LINT_VERSION",
	"GOSEC_VERSION",
	"GLOBAL_GO_BIN",
	"GOVULNCHECK_VERSION",
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

func TestLintDryRuns(t *testing.T) {
	tests := []struct {
		task  string
		token string
	}{
		{task: "golangci-lint:lint", token: "golangci-lint"},
		{task: "golangci-lint:lint:fix", token: "--fix"},
		{task: "gofumpt:lint", token: "gofumpt"},
		{task: "goimports:lint", token: "goimports"},
		{task: "govulncheck:lint", token: "govulncheck"},
		{task: "gosec:lint", token: "gosec"},
	}

	for _, tt := range tests {
		t.Run(tt.task, func(t *testing.T) {
			tasktest.AssertDryRunContains(t, "go", []string{tt.task}, tt.token)
		})
	}
}

func TestLintFixDryRun(t *testing.T) {
	tasktest.AssertDryRunContains(t, "go", []string{"lint:fix"},
		"golangci-lint",
		"--fix",
		"goimports",
		"gofumpt",
		"-w",
	)
}

func TestFmtDryRuns(t *testing.T) {
	tests := []struct {
		task   string
		tokens []string
	}{
		{task: "gofumpt:fmt", tokens: []string{"gofumpt", "-w"}},
		{task: "goimports:fmt", tokens: []string{"goimports", "-w"}},
		{task: "fmt", tokens: []string{"goimports", "gofumpt", "-w"}},
	}

	for _, tt := range tests {
		t.Run(tt.task, func(t *testing.T) {
			tasktest.AssertDryRunContains(t, "go", []string{tt.task}, tt.tokens...)
		})
	}
}

func TestDevelopmentToolDependencies(t *testing.T) {
	tf := tasktest.LoadTaskfile(t, "go")

	installTasks := map[string][]string{
		"install:golangci-lint": {"install"},
		"install:gofumpt":       {"install"},
		"install:goimports":     {"install"},
		"install:govulncheck":   {"install"},
		"install:gosec":         {"install"},
	}
	lintTasks := map[string][]string{
		"gofumpt:fmt":            {"install:gofumpt"},
		"golangci-lint:lint":     {"install:golangci-lint"},
		"golangci-lint:lint:fix": {"install:golangci-lint"},
		"goimports:fmt":          {"install:goimports"},
		"gofumpt:lint":           {"install:gofumpt"},
		"goimports:lint":         {"install:goimports"},
		"govulncheck:lint":       {"install:govulncheck"},
		"gosec:lint":             {"install:gosec"},
		"lint": {
			"golangci-lint:lint",
			"gofumpt:lint",
			"goimports:lint",
			"govulncheck:lint",
			"gosec:lint",
		},
	}

	for taskName, expected := range installTasks {
		assertTaskDependencies(t, tf, taskName, expected)
	}
	for taskName, expected := range lintTasks {
		assertTaskDependencies(t, tf, taskName, expected)
	}
}

func TestDevelopmentToolInstallVersions(t *testing.T) {
	tests := []struct {
		task       string
		module     string
		versionVar string
	}{
		{
			task:       "install:golangci-lint",
			module:     "github.com/golangci/golangci-lint/v2/cmd/golangci-lint",
			versionVar: "GOLANGCI_LINT_VERSION",
		},
		{
			task:       "install:gofumpt",
			module:     "mvdan.cc/gofumpt",
			versionVar: "GOFUMPT_VERSION",
		},
		{
			task:       "install:goimports",
			module:     "golang.org/x/tools/cmd/goimports",
			versionVar: "GOIMPORTS_VERSION",
		},
		{
			task:       "install:govulncheck",
			module:     "golang.org/x/vuln/cmd/govulncheck",
			versionVar: "GOVULNCHECK_VERSION",
		},
		{
			task:       "install:gosec",
			module:     "github.com/securego/gosec/v2/cmd/gosec",
			versionVar: "GOSEC_VERSION",
		},
	}

	for _, tt := range tests {
		t.Run(tt.task+"/latest", func(t *testing.T) {
			tasktest.AssertDryRunContains(t, "go", []string{tt.task}, tt.module+"@latest")
		})
		t.Run(tt.task+"/explicit", func(t *testing.T) {
			tasktest.AssertDryRunContains(t, "go",
				[]string{tt.task, tt.versionVar + "=v0.0.0-test"},
				tt.module+"@v0.0.0-test",
			)
		})
	}
}

func TestVersionVariablesAreIndependentAndOptional(t *testing.T) {
	tf := tasktest.LoadTaskfile(t, "go")

	if _, exists := tf.Vars["VERSION"]; exists {
		t.Fatal("shared VERSION variable must not be defined")
	}

	for _, name := range []string{
		"GO_VERSION",
		"GOLANGCI_LINT_VERSION",
		"GOFUMPT_VERSION",
		"GOIMPORTS_VERSION",
		"GOVULNCHECK_VERSION",
		"GOSEC_VERSION",
	} {
		value, exists := tf.Vars[name]
		if !exists {
			t.Fatalf("%s must be defined", name)
		}
		if value != "" {
			t.Fatalf("%s default = %#v, want empty", name, value)
		}
	}
}

func TestGoInstallVersionDryRun(t *testing.T) {
	switch runtime.GOOS {
	case "darwin":
		tasktest.AssertDryRunContains(t, "go",
			[]string{"install", "GO_VERSION=go1.99.1"},
			"go1.99.1.darwin-",
			".pkg",
		)
	case "linux":
		tasktest.AssertDryRunContains(t, "go",
			[]string{"install", "GO_VERSION=go1.99.1"},
			"go1.99.1.linux-",
			".tar.gz",
		)
	default:
		t.Skip("explicit Go version dry-run is covered on macOS and Linux")
	}
}

func TestAggregateLintDryRun(t *testing.T) {
	tasktest.AssertDryRunContains(t, "go", []string{"lint"},
		"golangci-lint",
		"gofumpt",
		"goimports",
		"govulncheck",
		"gosec",
	)
}

func assertTaskDependencies(t *testing.T, tf tasktest.Taskfile, taskName string, expected []string) {
	t.Helper()

	rawDeps, ok := tf.Tasks[taskName].Deps.([]any)
	if !ok {
		t.Fatalf("%s deps have type %T, want []any", taskName, tf.Tasks[taskName].Deps)
	}

	actual := make([]string, len(rawDeps))
	for i, rawDep := range rawDeps {
		dep, ok := rawDep.(string)
		if !ok {
			t.Fatalf("%s dependency %d has type %T, want string", taskName, i, rawDep)
		}
		actual[i] = dep
	}

	if !slices.Equal(actual, expected) {
		t.Fatalf("%s deps mismatch\nexpected: %v\nactual:   %v", taskName, expected, actual)
	}
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
