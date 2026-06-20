package config_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/mostafakhairy0305-dot/TaskOtter/internal/config"
)

func setEnv(t *testing.T, kv map[string]string) {
	t.Helper()

	for k, v := range kv {
		t.Setenv(k, v)
	}
}

func baseEnv(workspace string) map[string]string {
	return map[string]string{
		"INPUT_TASKS":         "go",
		"INPUT_JS":            "",
		"INPUT_INCLUDES_DOC":  "",
		"INPUT_STORE_VERSION": "",
		"INPUT_TARGET_FOLDER": "",
		"INPUT_GITHUB_TOKEN":  "token",
		"GITHUB_WORKSPACE":    workspace,
		"GITHUB_REPOSITORY":   "owner/repo",
	}
}

func TestLoadFromEnvDefaults(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	setEnv(t, baseEnv(dir))

	cfg, err := config.LoadFromEnv()
	if err != nil {
		t.Fatalf("LoadFromEnv() error = %v", err)
	}

	if cfg.TargetFolder != "taskfiles" {
		t.Fatalf("TargetFolder = %q, want taskfiles", cfg.TargetFolder)
	}

	if !cfg.IncludesDoc {
		t.Fatal("IncludesDoc should default to true")
	}

	if len(cfg.Tasks) != 1 || cfg.Tasks[0] != "go" {
		t.Fatalf("Tasks = %#v", cfg.Tasks)
	}
}

func TestLoadFromEnvGitHubTokenFallback(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	env := baseEnv(dir)
	env["INPUT_GITHUB_TOKEN"] = ""
	env["GITHUB_TOKEN"] = "fallback-token"
	setEnv(t, env)

	cfg, err := config.LoadFromEnv()
	if err != nil {
		t.Fatalf("LoadFromEnv() error = %v", err)
	}

	if cfg.GitHubToken != "fallback-token" {
		t.Fatalf("GitHubToken = %q, want fallback-token", cfg.GitHubToken)
	}
}

func TestLoadFromEnvDockerInputEnvNames(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	env := baseEnv(dir)
	env["INPUT_GITHUB_TOKEN"] = ""
	env["INPUT_GITHUB-TOKEN"] = "docker-token"
	env["INPUT_JS"] = "runtime: nodejs\npackage-manager: pnpm\nversion-manager: fnm\n"
	env["INPUT_INCLUDES-DOC"] = "false"
	env["INPUT_TARGET-FOLDER"] = "custom/taskfiles"
	setEnv(t, env)

	cfg, err := config.LoadFromEnv()
	if err != nil {
		t.Fatalf("LoadFromEnv() error = %v", err)
	}

	if cfg.GitHubToken != "docker-token" {
		t.Fatalf("GitHubToken = %q, want docker-token", cfg.GitHubToken)
	}

	if cfg.NodePackageManager != config.PMPnpm {
		t.Fatalf("NodePackageManager = %q, want pnpm", cfg.NodePackageManager)
	}

	if cfg.NodeVersionManager != config.VMFnm {
		t.Fatalf("NodeVersionManager = %q, want fnm", cfg.NodeVersionManager)
	}

	if cfg.IncludesDoc {
		t.Fatal("IncludesDoc = true, want false")
	}

	if cfg.TargetFolder != "custom/taskfiles" {
		t.Fatalf("TargetFolder = %q, want custom/taskfiles", cfg.TargetFolder)
	}
}

func TestParseTasksMultilineAndDedupe(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	env := baseEnv(dir)
	env["INPUT_TASKS"] = "eslint\nprettier,\ngo\ngo\n"
	setEnv(t, env)

	cfg, err := config.LoadFromEnv()
	if err != nil {
		t.Fatalf("LoadFromEnv() error = %v", err)
	}

	want := []string{"eslint", "prettier", "go"}
	if len(cfg.Tasks) != len(want) {
		t.Fatalf("Tasks = %#v, want %#v", cfg.Tasks, want)
	}

	for i := range want {
		if cfg.Tasks[i] != want[i] {
			t.Fatalf("Tasks = %#v, want %#v", cfg.Tasks, want)
		}
	}
}

func TestInvalidTaskName(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	env := baseEnv(dir)
	env["INPUT_TASKS"] = "../evil"
	setEnv(t, env)

	_, err := config.LoadFromEnv()
	if err == nil {
		t.Fatal("expected error for unsafe task name")
	}
}

func TestBunWithVersionManagerRejected(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	env := baseEnv(dir)
	env["INPUT_TASKS"] = "eslint"
	env["INPUT_JS"] = "runtime: bun\nversion-manager: fnm\n"
	setEnv(t, env)

	_, err := config.LoadFromEnv()
	if err == nil {
		t.Fatal("expected bun+version-manager validation error")
	}
}

func TestInvalidPackageManager(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	env := baseEnv(dir)
	env["INPUT_JS"] = "runtime: nodejs\npackage-manager: cargo\n"
	setEnv(t, env)

	_, err := config.LoadFromEnv()
	if err == nil {
		t.Fatal("expected invalid package manager error")
	}
}

func TestInvalidIncludesDoc(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	env := baseEnv(dir)
	env["INPUT_INCLUDES_DOC"] = "yes"
	setEnv(t, env)

	_, err := config.LoadFromEnv()
	if err == nil {
		t.Fatal("expected invalid includes-doc error")
	}
}

func TestFailOnChangesDefaultsFalse(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	setEnv(t, baseEnv(dir))

	cfg, err := config.LoadFromEnv()
	if err != nil {
		t.Fatalf("LoadFromEnv() error = %v", err)
	}

	if cfg.FailOnChanges {
		t.Fatal("FailOnChanges should default to false")
	}
}

func TestFailOnChangesTrue(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	env := baseEnv(dir)
	env["INPUT_FAIL-ON-CHANGES"] = "true"
	setEnv(t, env)

	cfg, err := config.LoadFromEnv()
	if err != nil {
		t.Fatalf("LoadFromEnv() error = %v", err)
	}

	if !cfg.FailOnChanges {
		t.Fatal("FailOnChanges = false, want true")
	}
}

func TestInvalidFailOnChanges(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	env := baseEnv(dir)
	env["INPUT_FAIL-ON-CHANGES"] = "yes"
	setEnv(t, env)

	_, err := config.LoadFromEnv()
	if err == nil {
		t.Fatal("expected invalid fail-on-changes error")
	}
}

func TestUnsafeStoreVersion(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	env := baseEnv(dir)
	env["INPUT_STORE_VERSION"] = "refs/heads/main"
	setEnv(t, env)

	_, err := config.LoadFromEnv()
	if err == nil {
		t.Fatal("expected unsafe store-version error")
	}
}

func TestTargetFolderValidation(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()

	cases := []struct {
		name  string
		value string
		ok    bool
	}{
		{"default nested", "automation/taskfiles", true},
		{"absolute unix", "/taskfiles", false},
		{"windows absolute", `C:\taskfiles`, false},
		{"dot", "../taskfiles", false},
		{"dot git", ".git", false},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			env := baseEnv(dir)
			env["INPUT_TARGET_FOLDER"] = testCase.value
			setEnv(t, env)

			_, err := config.LoadFromEnv()
			if testCase.ok && err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if !testCase.ok && err == nil {
				t.Fatal("expected validation error")
			}
		})
	}
}

func TestTargetFolderSymlinkEscape(t *testing.T) {
	t.Parallel()
	workspace := t.TempDir()
	outside := t.TempDir()

	link := filepath.Join(workspace, "link-out")

	err := os.Symlink(outside, link)
	if err != nil {
		t.Skip("symlink not permitted")
	}

	env := baseEnv(workspace)
	env["INPUT_TARGET_FOLDER"] = "link-out/taskfiles"
	setEnv(t, env)

	_, err = config.LoadFromEnv()
	if err == nil {
		t.Fatal("expected symlink escape rejection")
	}
}
