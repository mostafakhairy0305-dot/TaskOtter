package yarn_test

import (
	"os"
	"path/filepath"
	"runtime"
	"slices"
	"testing"

	"github.com/mostafakhairy0305-dot/TaskOtter/internal/tasktestutil"
	"gopkg.in/yaml.v3"
)

var publicTasks = []string{
	"add",
	"audit",
	"audit:json",
	"audit:report",
	"build",
	"cache:clean",
	"ci",
	"clean",
	"clean:all",
	"dev",
	"exec",
	"format",
	"install",
	"lint",
	"manager:pin",
	"manager:setup",
	"node:setup",
	"run",
	"test",
	"typecheck",
	"update",
	"version",
}

func TestTaskfileAndReadmePublicApi(t *testing.T) {
	doc := loadTaskfile(t)

	var root map[string]any
	if err := doc.Decode(&root); err != nil {
		t.Fatalf("decode Taskfile: %v", err)
	}

	tasks, ok := root["tasks"].(map[string]any)
	if !ok || len(tasks) == 0 {
		t.Fatal("Taskfile tasks map is missing")
	}

	actual := tasktestutil.SimplePublicTaskNames(tasks)
	if !slices.Equal(publicTasks, actual) {
		t.Fatalf("public task drift\nexpected: %v\nactual:   %v", publicTasks, actual)
	}

	readmeTasks := tasktestutil.ReadmePublicTaskNames(tasktestutil.MustRead(t, filepath.Join(".", "README.md")))
	if !slices.Equal(publicTasks, readmeTasks) {
		t.Fatalf("README public task drift\nexpected: %v\nactual:   %v", publicTasks, readmeTasks)
	}
}

func TestTaskCliLoadsAndDryRunsPublicTasks(t *testing.T) {
	for _, args := range [][]string{{"--list"}, {"--list-all"}, {"--list-all", "--json"}} {
		result := tasktestutil.RunSimpleTask(t, ".", stubEnv(t), args...)
		if result.Err != nil {
			t.Fatalf("task %v failed:\n%s", args, result.Output)
		}
	}

	for _, name := range publicTasks {
		args := []string{"--dry", "--yes", name}
		switch name {
		case "run":
			args = append(args, "SCRIPT=build")
		case "add":
			args = append(args, "PACKAGES=prettier")
		case "exec":
			args = append(args, "BINARY=prettier")
		case "manager:pin":
			args = append(args, "PACKAGE_MANAGER_VERSION=stable")
		}

		result := tasktestutil.RunSimpleTask(t, ".", stubEnv(t), args...)
		if result.Err != nil {
			t.Fatalf("dry run %s failed:\n%s", name, result.Output)
		}
	}
}

func TestStubbedYarnFlows(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Unix shell stubs cover these flows")
	}

	env := stubEnv(t)
	for _, args := range [][]string{
		{"--yes", "version"},
		{"--yes", "install"},
		{"--yes", "ci"},
		{"--yes", "run", "SCRIPT=test", "--", "--watch"},
	} {
		result := tasktestutil.RunSimpleTask(t, ".", env, args...)
		if result.Err != nil {
			t.Fatalf("task %v failed:\n%s", args, result.Output)
		}
	}

	result := tasktestutil.RunSimpleTask(t, ".", env, "--yes", "run", "SCRIPT=dev; exit 1")
	if result.Err == nil {
		t.Fatalf("unsafe SCRIPT unexpectedly succeeded:\n%s", result.Output)
	}
}

func stubEnv(t *testing.T) []string {
	t.Helper()

	home := t.TempDir()
	binDir := filepath.Join(home, ".local", "bin")
	if err := os.MkdirAll(binDir, 0755); err != nil {
		t.Fatalf("create stub bin dir: %v", err)
	}

	tasktestutil.WriteStub(t, binDir, "fnm", "#!/usr/bin/env bash\ncase \"$1\" in env) echo '# fnm env stub' ;; use) echo 'Using Node stub' ;; *) exit 0 ;; esac\n")
	tasktestutil.WriteStub(t, binDir, "node", "#!/usr/bin/env bash\nif [ \"$1\" = '--version' ]; then echo 'v22.0.0 stub'; fi\n")
	tasktestutil.WriteStub(t, binDir, "corepack", "#!/usr/bin/env bash\necho \"corepack $* stub\"\n")

	env := os.Environ()
	env = tasktestutil.SetEnv(env, "HOME", home)
	env = tasktestutil.SetEnv(env, "PATH", binDir+":"+os.Getenv("PATH"))
	env = tasktestutil.SetEnv(env, "TASK_ASSUME_YES", "true")
	env = tasktestutil.SetEnv(env, "NO_COLOR", "1")
	return env
}

func loadTaskfile(t *testing.T) yaml.Node {
	t.Helper()
	var doc yaml.Node
	if err := yaml.Unmarshal([]byte(tasktestutil.MustRead(t, filepath.Join(".", "Taskfile.yml"))), &doc); err != nil {
		t.Fatalf("parse Taskfile YAML: %v", err)
	}
	return doc
}
