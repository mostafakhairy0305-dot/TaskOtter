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
		"INPUT_TASKS":                "go",
		"INPUT_NODE_PACKAGE_MANAGER": "",
		"INPUT_NODE_VERSION_MANAGER": "",
		"INPUT_INCLUDES_DOC":         "",
		"INPUT_STORE_VERSION":        "",
		"INPUT_TARGET_FOLDER":        "",
		"INPUT_GITHUB_TOKEN":         "token",
		"GITHUB_WORKSPACE":           workspace,
		"GITHUB_REPOSITORY":          "owner/repo",
	}
}

func TestLoadFromEnvDefaults(t *testing.T) {
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

func TestParseTasksMultilineAndDedupe(t *testing.T) {
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
	dir := t.TempDir()
	env := baseEnv(dir)
	env["INPUT_TASKS"] = "../evil"
	setEnv(t, env)
	if _, err := config.LoadFromEnv(); err == nil {
		t.Fatal("expected error for unsafe task name")
	}
}

func TestBunWithVersionManagerRejected(t *testing.T) {
	dir := t.TempDir()
	env := baseEnv(dir)
	env["INPUT_TASKS"] = "eslint"
	env["INPUT_NODE_PACKAGE_MANAGER"] = "bun"
	env["INPUT_NODE_VERSION_MANAGER"] = "fnm"
	setEnv(t, env)
	if _, err := config.LoadFromEnv(); err == nil {
		t.Fatal("expected bun+fnm validation error")
	}
}

func TestInvalidPackageManager(t *testing.T) {
	dir := t.TempDir()
	env := baseEnv(dir)
	env["INPUT_NODE_PACKAGE_MANAGER"] = "cargo"
	setEnv(t, env)
	if _, err := config.LoadFromEnv(); err == nil {
		t.Fatal("expected invalid package manager error")
	}
}

func TestInvalidIncludesDoc(t *testing.T) {
	dir := t.TempDir()
	env := baseEnv(dir)
	env["INPUT_INCLUDES_DOC"] = "yes"
	setEnv(t, env)
	if _, err := config.LoadFromEnv(); err == nil {
		t.Fatal("expected invalid includes-doc error")
	}
}

func TestUnsafeStoreVersion(t *testing.T) {
	dir := t.TempDir()
	env := baseEnv(dir)
	env["INPUT_STORE_VERSION"] = "refs/heads/main"
	setEnv(t, env)
	if _, err := config.LoadFromEnv(); err == nil {
		t.Fatal("expected unsafe store-version error")
	}
}

func TestTargetFolderValidation(t *testing.T) {
	dir := t.TempDir()
	cases := []struct {
		name  string
		value string
		ok    bool
	}{
		{"default nested", "automation/taskfiles", true},
		{"absolute unix", "/taskfiles", false},
		{"windows absolute", `C:\taskfiles`, false},
		{"dot dot", "../taskfiles", false},
		{"dot git", ".git", false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			env := baseEnv(dir)
			env["INPUT_TARGET_FOLDER"] = tc.value
			setEnv(t, env)
			_, err := config.LoadFromEnv()
			if tc.ok && err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if !tc.ok && err == nil {
				t.Fatal("expected validation error")
			}
		})
	}
}

func TestTargetFolderSymlinkEscape(t *testing.T) {
	workspace := t.TempDir()
	outside := t.TempDir()
	link := filepath.Join(workspace, "link-out")
	if err := os.Symlink(outside, link); err != nil {
		t.Skip("symlink not permitted")
	}
	env := baseEnv(workspace)
	env["INPUT_TARGET_FOLDER"] = "link-out/taskfiles"
	setEnv(t, env)
	if _, err := config.LoadFromEnv(); err == nil {
		t.Fatal("expected symlink escape rejection")
	}
}
