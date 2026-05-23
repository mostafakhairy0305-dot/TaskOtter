package yarn_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"slices"
	"strings"
	"testing"

	"gopkg.in/yaml.v3"
)

var publicTasks = []string{
	"audit",
	"audit:json",
	"audit:report",
	"build",
	"cache:clean",
	"ci",
	"clean",
	"clean:all",
	"dev",
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

	actual := publicTaskNames(tasks)
	if !slices.Equal(publicTasks, actual) {
		t.Fatalf("public task drift\nexpected: %v\nactual:   %v", publicTasks, actual)
	}

	readme := mustRead(t, filepath.Join(taskDir(t), "README.md"))
	readmeTasks := readmePublicTasks(readme)
	if !slices.Equal(publicTasks, readmeTasks) {
		t.Fatalf("README public task drift\nexpected: %v\nactual:   %v", publicTasks, readmeTasks)
	}
}

func TestTaskCliLoadsAndDryRunsPublicTasks(t *testing.T) {
	for _, args := range [][]string{{"--list"}, {"--list-all"}, {"--list-all", "--json"}} {
		result := task(t, stubEnv(t), args...)
		if result.err != nil {
			t.Fatalf("task %v failed:\n%s", args, result.output)
		}
	}

	for _, name := range publicTasks {
		args := []string{"--dry", "--yes", name}
		if name == "run" {
			args = append(args, "SCRIPT=build")
		}
		if name == "manager:pin" {
			args = append(args, "PACKAGE_MANAGER_VERSION=stable")
		}

		result := task(t, stubEnv(t), args...)
		if result.err != nil {
			t.Fatalf("dry run %s failed:\n%s", name, result.output)
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
		result := task(t, env, args...)
		if result.err != nil {
			t.Fatalf("task %v failed:\n%s", args, result.output)
		}
	}

	result := task(t, env, "--yes", "run", "SCRIPT=dev; exit 1")
	if result.err == nil {
		t.Fatalf("unsafe SCRIPT unexpectedly succeeded:\n%s", result.output)
	}
}

type taskResult struct {
	output string
	err    error
}

func task(t *testing.T, env []string, args ...string) taskResult {
	t.Helper()

	cmd := exec.Command("task", args...)
	cmd.Dir = taskDir(t)
	cmd.Env = env
	out, err := cmd.CombinedOutput()

	return taskResult{output: string(out), err: err}
}

func stubEnv(t *testing.T) []string {
	t.Helper()

	home := t.TempDir()
	binDir := filepath.Join(home, ".local", "bin")
	if err := os.MkdirAll(binDir, 0755); err != nil {
		t.Fatalf("create stub bin dir: %v", err)
	}

	writeStub(t, binDir, "fnm", "#!/usr/bin/env bash\ncase \"$1\" in env) echo '# fnm env stub' ;; use) echo 'Using Node stub' ;; *) exit 0 ;; esac\n")
	writeStub(t, binDir, "node", "#!/usr/bin/env bash\nif [ \"$1\" = '--version' ]; then echo 'v22.0.0 stub'; fi\n")
	writeStub(t, binDir, "corepack", "#!/usr/bin/env bash\necho \"corepack $* stub\"\n")

	env := os.Environ()
	env = setEnv(env, "HOME", home)
	env = setEnv(env, "PATH", binDir+":"+os.Getenv("PATH"))
	env = setEnv(env, "TASK_ASSUME_YES", "true")
	env = setEnv(env, "NO_COLOR", "1")
	return env
}

func writeStub(t *testing.T, dir, name, body string) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(dir, name), []byte(body), 0755); err != nil {
		t.Fatalf("write %s stub: %v", name, err)
	}
}

func loadTaskfile(t *testing.T) yaml.Node {
	t.Helper()

	var doc yaml.Node
	if err := yaml.Unmarshal([]byte(mustRead(t, filepath.Join(taskDir(t), "Taskfile.yml"))), &doc); err != nil {
		t.Fatalf("parse Taskfile YAML: %v", err)
	}
	return doc
}

func publicTaskNames(tasks map[string]any) []string {
	names := []string{}
	for name, raw := range tasks {
		if name == "default" || strings.HasPrefix(name, "_") {
			continue
		}
		task, ok := raw.(map[string]any)
		if ok && task["internal"] == true {
			continue
		}
		names = append(names, name)
	}
	slices.Sort(names)
	return names
}

func readmePublicTasks(content string) []string {
	row := regexp.MustCompile("^\\|\\s*`([^`]+)`\\s*\\|")
	names := []string{}
	inTable := false

	for _, line := range strings.Split(content, "\n") {
		trimmed := strings.TrimSpace(line)
		if trimmed == "## Public Tasks" {
			inTable = true
			continue
		}
		if inTable && strings.HasPrefix(trimmed, "## ") {
			break
		}
		if inTable {
			if match := row.FindStringSubmatch(trimmed); len(match) == 2 {
				names = append(names, match[1])
			}
		}
	}

	slices.Sort(names)
	return names
}

func mustRead(t *testing.T, path string) string {
	t.Helper()
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	return string(content)
}

func taskDir(t *testing.T) string {
	t.Helper()
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("locate test file")
	}
	return filepath.Dir(file)
}

func setEnv(env []string, key, value string) []string {
	prefix := key + "="
	for i, item := range env {
		if strings.HasPrefix(item, prefix) {
			env[i] = prefix + value
			return env
		}
	}
	return append(env, prefix+value)
}
