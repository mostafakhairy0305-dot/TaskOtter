package bun_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"slices"
	"strings"
	"testing"
	"time"

	"gopkg.in/yaml.v3"
)

type PublicTaskSpec struct {
	Name                string
	Args                map[string]string
	MustDryRunWithArgs  bool
	RequiresGroupOutput bool
	RequiresPrompt      bool
	RequiresSummary     bool
}

var expectedPublicTasks = []PublicTaskSpec{
	{
		Name:               "add",
		Args:               map[string]string{"PACKAGES": "prettier"},
		MustDryRunWithArgs: true,
		RequiresGroupOutput: true,
	},
	{
		Name:               "exec",
		Args:               map[string]string{"BINARY": "prettier"},
		MustDryRunWithArgs: true,
		RequiresGroupOutput: true,
	},
	{
		Name:                "install",
		MustDryRunWithArgs:  true,
		RequiresGroupOutput: true,
		RequiresSummary:     true,
	},
	{
		Name:                "install:undo",
		MustDryRunWithArgs:  true,
		RequiresGroupOutput: true,
		RequiresPrompt:      true,
		RequiresSummary:     true,
	},
	{
		Name:                "upgrade",
		MustDryRunWithArgs:  true,
		RequiresGroupOutput: true,
		RequiresSummary:     true,
	},
	{
		Name:                "upgrade:canary",
		MustDryRunWithArgs:  true,
		RequiresGroupOutput: true,
		RequiresSummary:     true,
	},
	{
		Name:                "upgrade:stable",
		MustDryRunWithArgs:  true,
		RequiresGroupOutput: true,
		RequiresSummary:     true,
	},
	{
		Name:               "version",
		MustDryRunWithArgs: true,
		RequiresSummary:    true,
	},
}

func TestTaskBinaryIsAvailable(t *testing.T) {
	root := repoRoot(t)

	result := runTask(t, root, nil, "--version")

	assertExitCode(t, result, 0)
	assertNotEmpty(t, result.combined(), "task --version output is empty")
}

func TestTaskfileYamlIsCleanAndValid(t *testing.T) {
	taskfilePath := taskfilePath(t)
	content := readFile(t, taskfilePath)

	assertTextFileClean(t, taskfilePath, content)

	var doc yaml.Node
	if err := yaml.Unmarshal([]byte(content), &doc); err != nil {
		t.Fatalf("Taskfile YAML is invalid: %v", err)
	}

	assertNoDuplicateMappingKeys(t, &doc, "Taskfile")
	assertNoYamlAliases(t, &doc, "Taskfile")

	root := documentRoot(t, &doc)

	version := scalarField(root, "version")
	if version != "3" && !strings.HasPrefix(version, "3.") {
		t.Fatalf("Taskfile version must be 3 or 3.x, got %q", version)
	}

	tasks := mappingField(root, "tasks")
	if tasks == nil || len(tasks.Content) == 0 {
		t.Fatal("Taskfile must contain non-empty tasks map")
	}
}

func TestTaskCliCanLoadTaskfile(t *testing.T) {
	root := repoRoot(t)

	for _, args := range [][]string{
		{"--list"},
		{"--list-all"},
		{"--list-all", "--sort", "alphanumeric"},
		{"--list-all", "--json"},
	} {
		args := args

		t.Run(strings.Join(args, " "), func(t *testing.T) {
			result := runTask(t, root, isolatedEnv(t), args...)

			assertExitCode(t, result, 0)
			assertNotContains(t, strings.ToLower(result.combined()), "taskfile does not exist")
			assertNotContains(t, strings.ToLower(result.combined()), "unknown")
		})
	}
}

func TestTaskListAllJsonIsValid(t *testing.T) {
	root := repoRoot(t)

	result := runTask(t, root, isolatedEnv(t), "--list-all", "--json")

	assertExitCode(t, result, 0)

	var payload any
	if err := json.Unmarshal([]byte(result.stdout), &payload); err != nil {
		t.Fatalf("task --list-all --json did not produce valid JSON:\n%s\nerror: %v", result.stdout, err)
	}
}

func TestPublicApiDoesNotDrift(t *testing.T) {
	tf := loadTaskfile(t)

	expected := expectedPublicTaskNames()
	actual := publicTaskNamesFromTaskfile(t, tf)

	if !slices.Equal(expected, actual) {
		t.Fatalf(
			"public Taskfile API drift detected\n\nexpected:\n%s\n\nactual:\n%s\n\nFix either the Taskfile public tasks or expectedPublicTasks in the test.",
			formatList(expected),
			formatList(actual),
		)
	}
}

func TestEveryTaskIsEitherPublicOrInternal(t *testing.T) {
	tf := loadTaskfile(t)

	for name, task := range tf.Tasks {
		name := name
		task := task

		t.Run(name, func(t *testing.T) {
			if strings.HasPrefix(name, "_") {
				return
			}

			if task.boolField("internal") {
				return
			}

			if task.stringField("desc") == "" {
				t.Fatalf("task %q is not internal and has no desc. Either add desc/summary or mark it internal: true", name)
			}
		})
	}
}

func TestPublicTasksHaveMetadata(t *testing.T) {
	tf := loadTaskfile(t)

	for _, spec := range expectedPublicTasks {
		spec := spec

		t.Run(spec.Name, func(t *testing.T) {
			t.Parallel()
			task := mustTask(t, tf, spec.Name)

			if task.node.Kind != yaml.MappingNode {
				t.Fatalf("public task %q must use full mapping syntax, not short syntax", spec.Name)
			}

			desc := task.stringField("desc")
			summary := task.stringField("summary")

			if strings.TrimSpace(desc) == "" {
				t.Fatalf("public task %q is missing desc", spec.Name)
			}

			if len(strings.TrimSpace(desc)) < 12 {
				t.Fatalf("public task %q desc is too short: %q", spec.Name, desc)
			}

			if spec.RequiresSummary && strings.TrimSpace(summary) == "" {
				t.Fatalf("public task %q is missing summary", spec.Name)
			}

			if spec.RequiresSummary && len(strings.TrimSpace(summary)) < 25 {
				t.Fatalf("public task %q summary is too short:\n%s", spec.Name, summary)
			}

			assertNoPlaceholderText(t, spec.Name, desc)
			assertNoPlaceholderText(t, spec.Name, summary)
		})
	}
}

func TestDestructivePublicTasksHavePrompt(t *testing.T) {
	tf := loadTaskfile(t)

	for _, spec := range expectedPublicTasks {
		spec := spec

		t.Run(spec.Name, func(t *testing.T) {
			t.Parallel()
			if !spec.RequiresPrompt {
				return
			}

			task := mustTask(t, tf, spec.Name)

			prompt := task.field("prompt")
			if prompt == nil || nodeText(prompt) == "" {
				t.Fatalf("destructive task %q must have a non-empty prompt", spec.Name)
			}

			text := strings.ToLower(nodeText(prompt))
			if !strings.Contains(text, "sure") &&
				!strings.Contains(text, "confirm") &&
				!strings.Contains(text, "remove") &&
				!strings.Contains(text, "uninstall") &&
				!strings.Contains(text, "delete") &&
				!strings.Contains(text, "continue") {
				t.Fatalf("prompt for task %q does not look explicit enough:\n%s", spec.Name, nodeText(prompt))
			}
		})
	}
}

func TestInstallTasksUseGithubGroupOutput(t *testing.T) {
	tf := loadTaskfile(t)

	for _, spec := range expectedPublicTasks {
		spec := spec

		t.Run(spec.Name, func(t *testing.T) {
			t.Parallel()
			if !spec.RequiresGroupOutput {
				return
			}

			task := mustTask(t, tf, spec.Name)

			outputNode := task.field("output")
			if outputNode == nil {
				outputNode = tf.Root.field("output")
			}

			assertGithubGroupOutput(t, spec.Name, outputNode)
		})
	}
}

func TestPublicTasksHaveCommands(t *testing.T) {
	tf := loadTaskfile(t)

	for _, spec := range expectedPublicTasks {
		spec := spec

		t.Run(spec.Name, func(t *testing.T) {
			t.Parallel()
			task := mustTask(t, tf, spec.Name)

			cmds := task.field("cmds")
			deps := task.field("deps")

			if isEmptyNode(cmds) && isEmptyNode(deps) {
				t.Fatalf("public task %q must have cmds or deps", spec.Name)
			}
		})
	}
}

func TestTaskSummariesWork(t *testing.T) {
	root := repoRoot(t)

	for _, spec := range expectedPublicTasks {
		spec := spec

		t.Run(spec.Name, func(t *testing.T) {
			t.Parallel()
			if !spec.RequiresSummary {
				return
			}

			result := runTask(t, root, isolatedEnv(t), "--summary", spec.Name)

			assertExitCode(t, result, 0)

			out := result.combined()
			assertContains(t, out, spec.Name)
			assertNotContains(t, strings.ToLower(out), "task not found")
			assertNotContains(t, strings.ToLower(out), "unknown task")
			assertNotContains(t, strings.ToLower(out), "no summary")
		})
	}
}

func TestPublicTasksDryRunWithExpectedArgs(t *testing.T) {
	root := repoRoot(t)

	for _, spec := range expectedPublicTasks {
		spec := spec

		t.Run(spec.Name, func(t *testing.T) {
			t.Parallel()
			if !spec.MustDryRunWithArgs {
				return
			}

			args := []string{"--dry", "--yes", spec.Name}
			args = append(args, taskArgs(spec.Args)...)

			result := runTask(t, root, bunDryRunEnv(t), args...)

			assertExitCode(t, result, 0)

			out := result.combined()
			assertNotContains(t, strings.ToLower(out), "task not found")
			assertNotContains(t, strings.ToLower(out), "unknown task")
			assertNotContains(t, strings.ToLower(out), "cannot find")
			assertNotContains(t, strings.ToLower(out), "missing required")
		})
	}
}

func TestUndoPairsExist(t *testing.T) {
	tf := loadTaskfile(t)

	for task, undo := range map[string]string{
		"install": "install:undo",
	} {
		if _, ok := tf.Tasks[task]; !ok {
			t.Fatalf("task %q is missing", task)
		}
		if _, ok := tf.Tasks[undo]; !ok {
			t.Fatalf("undo task %q for %q is missing", undo, task)
		}
	}

	undoTask, ok := tf.Tasks["install:undo"]
	if !ok {
		t.Fatal("task install:undo is missing")
	}
	if !hasAlias(undoTask, "uninstall") {
		t.Fatal("task install:undo is missing alias uninstall")
	}
}

func TestAliasesDryRun(t *testing.T) {
	root := repoRoot(t)

	cases := []struct {
		alias string
		args  []string
	}{
		{"uninstall", nil},
	}

	for _, tc := range cases {
		tc := tc

		t.Run(tc.alias, func(t *testing.T) {
			t.Parallel()
			args := append([]string{"--dry", "--yes", tc.alias}, tc.args...)
			result := runTask(t, root, bunDryRunEnv(t), args...)

			assertExitCode(t, result, 0)

			out := result.combined()
			assertNotContains(t, strings.ToLower(out), "task not found")
			assertNotContains(t, strings.ToLower(out), "unknown task")
		})
	}
}

func TestReferencedScriptsExist(t *testing.T) {
	root := repoRoot(t)
	tf := loadTaskfile(t)

	for taskName, task := range tf.Tasks {
		taskName := taskName
		commands := collectCommandStrings(task.node)

		for _, command := range commands {
			command := command

			t.Run(taskName, func(t *testing.T) {
				t.Parallel()
				for _, scriptPath := range referencedLocalShellScripts(command) {
					abs := filepath.Join(root, scriptPath)

					info, err := os.Stat(abs)
					if err != nil {
						t.Fatalf("task %q references missing script %q", taskName, scriptPath)
					}

					if info.IsDir() {
						t.Fatalf("task %q references script path but it is a directory: %q", taskName, scriptPath)
					}
				}
			})
		}
	}
}

func TestCommandsDoNotContainDangerousPatterns(t *testing.T) {
	tf := loadTaskfile(t)

	dangerousPatterns := []*regexp.Regexp{
		regexp.MustCompile(`(?m)\brm\s+-[a-zA-Z]*r[a-zA-Z]*f[a-zA-Z]*\s+/(?:\s|$)`),
		regexp.MustCompile(`(?m)\bsudo\s+rm\s+-[a-zA-Z]*r[a-zA-Z]*f`),
		regexp.MustCompile(`(?m)\bchmod\s+-R\s+777\s+/`),
		regexp.MustCompile(`(?m)\bcurl\b.*\s-k(?:\s|$)`),
		regexp.MustCompile(`(?m)\bcurl\b.*--insecure`),
	}

	for taskName, task := range tf.Tasks {
		taskName := taskName

		for _, command := range collectCommandStrings(task.node) {
			for _, pattern := range dangerousPatterns {
				if pattern.MatchString(command) {
					t.Fatalf("task %q contains dangerous command pattern %q:\n%s", taskName, pattern.String(), command)
				}
			}
		}
	}
}

func TestNoPlaceholderTextInTaskfile(t *testing.T) {
	content := readFile(t, taskfilePath(t))

	placeholders := []string{
		"TODO",
		"FIXME",
		"CHANGEME",
		"REPLACE_ME",
		"your value here",
		"lorem ipsum",
	}

	upper := strings.ToUpper(content)

	for _, placeholder := range placeholders {
		if strings.Contains(upper, strings.ToUpper(placeholder)) {
			t.Fatalf("Taskfile contains placeholder text: %s", placeholder)
		}
	}
}

// Stub-based behavioral tests — these run real task commands against the
// bunDryRunEnv stub (a fake bun binary) so they exercise actual task logic
// without downloading or installing anything.

func TestVersionTaskExitsSuccessfully(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("stub bun tests target Unix-like systems")
	}
	t.Parallel()

	root := repoRoot(t)
	result := runTask(t, root, bunDryRunEnv(t), "--yes", "version")

	assertExitCode(t, result, 0)
}

func TestVersionTaskPrintsBunVersion(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("stub bun tests target Unix-like systems")
	}
	t.Parallel()

	root := repoRoot(t)
	result := runTask(t, root, bunDryRunEnv(t), "--yes", "version")

	assertExitCode(t, result, 0)
	assertContains(t, result.combined(), "1.")
}

// TestInstallIsIdempotentWithStubBun verifies that running task install twice
// succeeds both times — the second run should be a no-op because the status
// check sees the stub bun binary already present.
func TestInstallIsIdempotentWithStubBun(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("stub bun tests target Unix-like systems")
	}
	t.Parallel()

	root := repoRoot(t)
	env := bunDryRunEnv(t)

	assertExitCode(t, runTask(t, root, env, "--yes", "install"), 0)
	assertExitCode(t, runTask(t, root, env, "--yes", "install"), 0)
}

// TestInstallSkipsWhenBunIsAlreadyPresent verifies that install does not
// print an "Installing" message when the stub bun binary is already present
// and no VERSION is specified — the status check should short-circuit the task.
func TestInstallSkipsWhenBunIsAlreadyPresent(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("stub bun tests target Unix-like systems")
	}
	t.Parallel()

	root := repoRoot(t)
	result := runTask(t, root, bunDryRunEnv(t), "--yes", "install")

	assertExitCode(t, result, 0)
	assertNotContains(t, result.combined(), "Installing Bun")
}

// TestInstallUndoRemovesBunDir runs install:undo and verifies that the
// $HOME/.bun directory is actually removed from disk.
func TestInstallUndoRemovesBunDir(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("stub bun tests target Unix-like systems")
	}
	t.Parallel()

	root := repoRoot(t)
	env := bunDryRunEnv(t)
	bunDir := bunInstallDir(env)

	assertDirExists(t, bunDir)

	result := runTask(t, root, env, "--yes", "install:undo")
	assertExitCode(t, result, 0)

	assertDirNotExists(t, bunDir)
}

// TestInstallUndoIsIdempotent verifies that running install:undo when Bun is
// already absent exits successfully without error.
func TestInstallUndoIsIdempotent(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("stub bun tests target Unix-like systems")
	}
	t.Parallel()

	root := repoRoot(t)
	env := bunDryRunEnv(t)

	assertExitCode(t, runTask(t, root, env, "--yes", "install:undo"), 0)
	assertExitCode(t, runTask(t, root, env, "--yes", "install:undo"), 0)
}

// TestUpgradeExitsSuccessfully verifies that task upgrade succeeds when the
// stub bun binary is present, exercising the full install dep → upgrade path.
func TestUpgradeExitsSuccessfully(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("stub bun tests target Unix-like systems")
	}
	t.Parallel()

	root := repoRoot(t)
	result := runTask(t, root, bunDryRunEnv(t), "--yes", "upgrade")

	assertExitCode(t, result, 0)
}

// TestUpgradeCanaryExitsSuccessfully verifies that task upgrade:canary succeeds
// when the stub bun binary is present.
func TestUpgradeCanaryExitsSuccessfully(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("stub bun tests target Unix-like systems")
	}
	t.Parallel()

	root := repoRoot(t)
	result := runTask(t, root, bunDryRunEnv(t), "--yes", "upgrade:canary")

	assertExitCode(t, result, 0)
}

// TestUpgradeStableExitsSuccessfully verifies that task upgrade:stable succeeds
// when the stub bun binary is present.
func TestUpgradeStableExitsSuccessfully(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("stub bun tests target Unix-like systems")
	}
	t.Parallel()

	root := repoRoot(t)
	result := runTask(t, root, bunDryRunEnv(t), "--yes", "upgrade:stable")

	assertExitCode(t, result, 0)
}

func TestRealInstallerFlowOnlyWhenExplicitlyEnabled(t *testing.T) {
	if os.Getenv("RUN_INSTALLER_TESTS") != "1" {
		t.Skip("set RUN_INSTALLER_TESTS=1 to run real install/uninstall tests")
	}

	if runtime.GOOS == "windows" {
		t.Skip("real bun installer tests are intended for Unix-like systems")
	}

	root := repoRoot(t)
	env := isolatedEnv(t)
	home := envValue(env, "HOME")
	bunBin := filepath.Join(home, ".bun", "bin", "bun")

	install := runTaskTimeout(t, root, env, 10*time.Minute, "--yes", "install")
	assertExitCode(t, install, 0)

	assertFileExists(t, bunBin)

	version := runTaskTimeout(t, root, env, 10*time.Minute, "version")
	assertExitCode(t, version, 0)

	undo := runTaskTimeout(t, root, env, 10*time.Minute, "--yes", "install:undo")
	assertExitCode(t, undo, 0)

	if _, err := os.Stat(filepath.Join(home, ".bun")); !os.IsNotExist(err) {
		t.Fatalf("expected .bun directory to be removed after install:undo: %s", home)
	}
}

func TestAllPublicTasksIntegration(t *testing.T) {
	if os.Getenv("RUN_INTEGRATION_TESTS") != "1" {
		t.Skip("set RUN_INTEGRATION_TESTS=1 to run integration tests (downloads and installs Bun)")
	}

	if runtime.GOOS == "windows" {
		t.Skip("integration tests target Unix-like systems")
	}

	root := repoRoot(t)
	env := isolatedEnv(t)
	home := envValue(env, "HOME")
	bunBin := filepath.Join(home, ".bun", "bin", "bun")

	step := func(name string, fn func(t *testing.T)) {
		t.Helper()
		t.Run(name, fn)
		if t.Failed() {
			t.FailNow()
		}
	}

	run := func(t *testing.T, args ...string) commandResult {
		t.Helper()
		result := runTaskTimeout(t, root, env, 10*time.Minute, args...)
		assertExitCode(t, result, 0)
		return result
	}

	step("install — bun binary is present on disk", func(t *testing.T) {
		run(t, "--yes", "install")
		assertFileExists(t, bunBin)
	})

	step("version — bun version string is printed", func(t *testing.T) {
		result := run(t, "version")
		assertNotEmpty(t, result.combined(), "version output is empty")
	})

	step("upgrade — bun upgrades without error", func(t *testing.T) {
		run(t, "--yes", "upgrade")
		assertFileExists(t, bunBin)
	})

	step("upgrade:canary — bun switches to canary without error", func(t *testing.T) {
		run(t, "--yes", "upgrade:canary")
		assertFileExists(t, bunBin)
	})

	step("upgrade:stable — bun switches back to stable without error", func(t *testing.T) {
		run(t, "--yes", "upgrade:stable")
		assertFileExists(t, bunBin)
	})

	step("install:undo — .bun directory is removed", func(t *testing.T) {
		run(t, "--yes", "install:undo")
		assertDirNotExists(t, filepath.Join(home, ".bun"))
	})
}

// bunInstallDir returns the path where bun is installed in the stub environment.
func bunInstallDir(env []string) string {
	home := envValue(env, "HOME")
	return filepath.Join(home, ".bun")
}

// bunDryRunEnv returns an isolated environment with a stub bun binary installed
// at $HOME/.bun/bin/bun and added to PATH. The stub satisfies bun precondition
// checks (command -v bun, test -f $HOME/.bun/bin/bun) and responds to the
// subcommands used in task commands without performing real operations.
func bunDryRunEnv(t *testing.T) []string {
	t.Helper()

	env := isolatedEnv(t)
	home := envValue(env, "HOME")

	bunBinDir := filepath.Join(home, ".bun", "bin")
	if err := os.MkdirAll(bunBinDir, 0755); err != nil {
		t.Fatalf("failed to create stub bun dir: %v", err)
	}

	stub := "#!/usr/bin/env bash\n" +
		"case \"$1\" in\n" +
		"  --version) echo \"1.2.3\" ;;\n" +
		"  --revision) echo \"abc1234\" ;;\n" +
		"  upgrade) echo \"Bun is already at the latest version\" ;;\n" +
		"  *) exit 0 ;;\n" +
		"esac\n"

	stubPath := filepath.Join(bunBinDir, "bun")
	if err := os.WriteFile(stubPath, []byte(stub), 0755); err != nil {
		t.Fatalf("failed to create stub bun binary: %v", err)
	}

	path := envValue(env, "PATH")
	env = setEnv(env, "PATH", bunBinDir+":"+path)

	return env
}

type LoadedTaskfile struct {
	Path  string
	Root  taskNode
	Tasks map[string]taskNode
}

type taskNode struct {
	name string
	node *yaml.Node
}

func loadTaskfile(t *testing.T) LoadedTaskfile {
	t.Helper()

	path := taskfilePath(t)
	content := readFile(t, path)

	var doc yaml.Node
	if err := yaml.Unmarshal([]byte(content), &doc); err != nil {
		t.Fatalf("failed to parse Taskfile: %v", err)
	}

	root := documentRoot(t, &doc)
	tasksNode := mappingField(root, "tasks")
	if tasksNode == nil {
		t.Fatal("Taskfile has no tasks map")
	}

	tasks := map[string]taskNode{}

	for i := 0; i < len(tasksNode.Content); i += 2 {
		key := tasksNode.Content[i]
		value := tasksNode.Content[i+1]

		tasks[key.Value] = taskNode{
			name: key.Value,
			node: value,
		}
	}

	return LoadedTaskfile{
		Path:  path,
		Root:  taskNode{name: "root", node: root},
		Tasks: tasks,
	}
}

func mustTask(t *testing.T, tf LoadedTaskfile, name string) taskNode {
	t.Helper()

	task, ok := tf.Tasks[name]
	if !ok {
		t.Fatalf("expected public task %q is missing", name)
	}

	return task
}

func hasAlias(task taskNode, alias string) bool {
	aliases := task.field("aliases")
	if aliases == nil || aliases.Kind != yaml.SequenceNode {
		return false
	}

	for _, item := range aliases.Content {
		if item.Value == alias {
			return true
		}
	}

	return false
}

func (n taskNode) field(name string) *yaml.Node {
	if n.node == nil || n.node.Kind != yaml.MappingNode {
		return nil
	}

	for i := 0; i < len(n.node.Content); i += 2 {
		if n.node.Content[i].Value == name {
			return n.node.Content[i+1]
		}
	}

	return nil
}

func (n taskNode) stringField(name string) string {
	return nodeText(n.field(name))
}

func (n taskNode) boolField(name string) bool {
	field := n.field(name)
	if field == nil {
		return false
	}

	return strings.EqualFold(field.Value, "true")
}

func repoRoot(t *testing.T) string {
	t.Helper()

	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get working directory: %v", err)
	}

	current := wd

	for {
		if fileExists(filepath.Join(current, "Taskfile.yml")) ||
			fileExists(filepath.Join(current, "Taskfile.yaml")) {
			return current
		}

		parent := filepath.Dir(current)
		if parent == current {
			t.Fatal("could not find Taskfile.yml or Taskfile.yaml")
		}

		current = parent
	}
}

func taskfilePath(t *testing.T) string {
	t.Helper()

	root := repoRoot(t)

	for _, name := range []string{"Taskfile.yml", "Taskfile.yaml"} {
		path := filepath.Join(root, name)
		if fileExists(path) {
			return path
		}
	}

	t.Fatal("could not find Taskfile.yml or Taskfile.yaml")
	return ""
}

func runTask(t *testing.T, root string, env []string, args ...string) commandResult {
	return runTaskTimeout(t, root, env, 2*time.Minute, args...)
}

func runTaskTimeout(t *testing.T, root string, env []string, timeout time.Duration, args ...string) commandResult {
	t.Helper()

	taskBin := os.Getenv("TASK_BIN")
	if taskBin == "" {
		taskBin = "task"
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, taskBin, args...)
	cmd.Dir = root

	if env != nil {
		cmd.Env = env
	}

	var stdout bytes.Buffer
	var stderr bytes.Buffer

	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	cmd.WaitDelay = 5 * time.Second

	err := cmd.Run()

	return commandResult{
		stdout: stdout.String(),
		stderr: stderr.String(),
		err:    err,
		args:   args,
	}
}

type commandResult struct {
	stdout string
	stderr string
	err    error
	args   []string
}

func (r commandResult) combined() string {
	return r.stdout + "\n" + r.stderr
}

func isolatedEnv(t *testing.T) []string {
	t.Helper()

	home := t.TempDir()
	profile := filepath.Join(home, ".bashrc")

	if err := os.WriteFile(profile, []byte(""), 0644); err != nil {
		t.Fatalf("failed to create fake shell profile: %v", err)
	}

	env := os.Environ()

	env = setEnv(env, "HOME", home)
	env = setEnv(env, "PROFILE", profile)
	env = setEnv(env, "ZDOTDIR", home)
	env = setEnv(env, "CI", "true")
	env = setEnv(env, "TASK_COLOR", "0")
	env = setEnv(env, "NO_COLOR", "1")
	env = setEnv(env, "TASK_ASSUME_YES", "true")

	return env
}

func expectedPublicTaskNames() []string {
	names := make([]string, 0, len(expectedPublicTasks))

	for _, task := range expectedPublicTasks {
		names = append(names, task.Name)
	}

	slices.Sort(names)

	return names
}

func publicTaskNamesFromTaskfile(t *testing.T, tf LoadedTaskfile) []string {
	t.Helper()

	names := []string{}

	for name, task := range tf.Tasks {
		if name == "default" {
			continue
		}

		if strings.HasPrefix(name, "_") {
			continue
		}

		if task.boolField("internal") {
			continue
		}

		if task.stringField("desc") != "" {
			names = append(names, name)
		}
	}

	slices.Sort(names)

	return names
}

func taskArgs(args map[string]string) []string {
	if len(args) == 0 {
		return nil
	}

	keys := make([]string, 0, len(args))
	for key := range args {
		keys = append(keys, key)
	}

	slices.Sort(keys)

	out := make([]string, 0, len(keys))

	for _, key := range keys {
		out = append(out, fmt.Sprintf("%s=%s", key, args[key]))
	}

	return out
}

func assertGithubGroupOutput(t *testing.T, taskName string, outputNode *yaml.Node) {
	t.Helper()

	if outputNode == nil {
		t.Fatalf("task %q requires output.group config but no output config was found", taskName)
	}

	if outputNode.Kind != yaml.MappingNode {
		t.Fatalf("task %q output must use advanced object format, not scalar format", taskName)
	}

	groupNode := nodeMappingValue(outputNode, "group")
	if groupNode == nil || groupNode.Kind != yaml.MappingNode {
		t.Fatalf("task %q output must include group config", taskName)
	}

	begin := nodeText(nodeMappingValue(groupNode, "begin"))
	end := nodeText(nodeMappingValue(groupNode, "end"))
	errorOnly := nodeMappingValue(groupNode, "error_only")

	if begin != "::group::{{.TASK}}" {
		t.Fatalf("task %q output.group.begin must be %q, got %q", taskName, "::group::{{.TASK}}", begin)
	}

	if end != "::endgroup::" {
		t.Fatalf("task %q output.group.end must be %q, got %q", taskName, "::endgroup::", end)
	}

	if errorOnly == nil {
		t.Fatalf("task %q output.group.error_only must be explicitly set to false", taskName)
	}

	if !strings.EqualFold(errorOnly.Value, "false") {
		t.Fatalf("task %q output.group.error_only must be false, got %q", taskName, errorOnly.Value)
	}
}

func assertTextFileClean(t *testing.T, path string, content string) {
	t.Helper()

	if content == "" {
		t.Fatalf("%s is empty", path)
	}

	if strings.Contains(content, "\r\n") {
		t.Fatalf("%s uses CRLF line endings; use LF only", path)
	}

	if strings.Contains(content, "\t") {
		t.Fatalf("%s contains tabs; use spaces in YAML", path)
	}

	if !strings.HasSuffix(content, "\n") {
		t.Fatalf("%s must end with a newline", path)
	}

	lines := strings.Split(content, "\n")
	for i, line := range lines {
		if strings.TrimRight(line, " ") != line {
			t.Fatalf("%s has trailing whitespace at line %d", path, i+1)
		}
	}
}

func assertNoDuplicateMappingKeys(t *testing.T, node *yaml.Node, path string) {
	t.Helper()

	if node == nil {
		return
	}

	if node.Kind == yaml.DocumentNode && len(node.Content) > 0 {
		assertNoDuplicateMappingKeys(t, node.Content[0], path)
		return
	}

	if node.Kind == yaml.MappingNode {
		seen := map[string]bool{}

		for i := 0; i < len(node.Content); i += 2 {
			key := node.Content[i]
			value := node.Content[i+1]

			if seen[key.Value] {
				t.Fatalf("duplicate YAML key at %s.%s", path, key.Value)
			}

			seen[key.Value] = true
			assertNoDuplicateMappingKeys(t, value, path+"."+key.Value)
		}
	}

	if node.Kind == yaml.SequenceNode {
		for i, child := range node.Content {
			assertNoDuplicateMappingKeys(t, child, fmt.Sprintf("%s[%d]", path, i))
		}
	}
}

func assertNoYamlAliases(t *testing.T, node *yaml.Node, path string) {
	t.Helper()

	if node == nil {
		return
	}

	if node.Kind == yaml.AliasNode {
		t.Fatalf("YAML aliases/anchors are not allowed for clean Taskfile config at %s", path)
	}

	for i, child := range node.Content {
		assertNoYamlAliases(t, child, fmt.Sprintf("%s[%d]", path, i))
	}
}

func assertNoPlaceholderText(t *testing.T, taskName string, value string) {
	t.Helper()

	upper := strings.ToUpper(value)

	for _, placeholder := range []string{"TODO", "FIXME", "CHANGEME", "REPLACE_ME", "LOREM IPSUM"} {
		if strings.Contains(upper, placeholder) {
			t.Fatalf("task %q contains placeholder text %q", taskName, placeholder)
		}
	}
}

func collectCommandStrings(node *yaml.Node) []string {
	if node == nil {
		return nil
	}

	var out []string

	switch node.Kind {
	case yaml.ScalarNode:
		if strings.TrimSpace(node.Value) != "" {
			out = append(out, node.Value)
		}

	case yaml.SequenceNode:
		for _, child := range node.Content {
			out = append(out, collectCommandStrings(child)...)
		}

	case yaml.MappingNode:
		for i := 0; i < len(node.Content); i += 2 {
			key := node.Content[i]
			value := node.Content[i+1]

			switch key.Value {
			case "cmd", "sh":
				if value.Kind == yaml.ScalarNode {
					out = append(out, value.Value)
				}
			case "cmds", "status", "preconditions":
				out = append(out, collectCommandStrings(value)...)
			}
		}
	}

	return out
}

func referencedLocalShellScripts(command string) []string {
	re := regexp.MustCompile(`(?:^|\s)(\./[A-Za-z0-9_./-]+\.sh)(?:\s|$)`)
	matches := re.FindAllStringSubmatch(command, -1)

	var out []string

	for _, match := range matches {
		if len(match) > 1 {
			out = append(out, match[1])
		}
	}

	return out
}

func documentRoot(t *testing.T, doc *yaml.Node) *yaml.Node {
	t.Helper()

	if doc.Kind != yaml.DocumentNode || len(doc.Content) == 0 {
		t.Fatal("invalid YAML document")
	}

	root := doc.Content[0]
	if root.Kind != yaml.MappingNode {
		t.Fatal("Taskfile root must be a YAML mapping")
	}

	return root
}

func mappingField(root *yaml.Node, name string) *yaml.Node {
	node := nodeMappingValue(root, name)
	if node == nil || node.Kind != yaml.MappingNode {
		return nil
	}

	return node
}

func scalarField(root *yaml.Node, name string) string {
	return nodeText(nodeMappingValue(root, name))
}

func nodeMappingValue(mapping *yaml.Node, key string) *yaml.Node {
	if mapping == nil || mapping.Kind != yaml.MappingNode {
		return nil
	}

	for i := 0; i < len(mapping.Content); i += 2 {
		if mapping.Content[i].Value == key {
			return mapping.Content[i+1]
		}
	}

	return nil
}

func nodeText(node *yaml.Node) string {
	if node == nil {
		return ""
	}

	if node.Kind == yaml.ScalarNode {
		return strings.TrimSpace(node.Value)
	}

	var parts []string

	for _, child := range node.Content {
		text := nodeText(child)
		if text != "" {
			parts = append(parts, text)
		}
	}

	return strings.TrimSpace(strings.Join(parts, " "))
}

func isEmptyNode(node *yaml.Node) bool {
	if node == nil {
		return true
	}

	if node.Kind == yaml.ScalarNode {
		return strings.TrimSpace(node.Value) == ""
	}

	return len(node.Content) == 0
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func readFile(t *testing.T, path string) string {
	t.Helper()

	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to read %s: %v", path, err)
	}

	return string(content)
}

func setEnv(env []string, key string, value string) []string {
	prefix := key + "="

	for i, item := range env {
		if strings.HasPrefix(item, prefix) {
			env[i] = prefix + value
			return env
		}
	}

	return append(env, prefix+value)
}

func envValue(env []string, key string) string {
	prefix := key + "="

	for _, item := range env {
		if strings.HasPrefix(item, prefix) {
			return strings.TrimPrefix(item, prefix)
		}
	}

	return ""
}

func assertExitCode(t *testing.T, result commandResult, expected int) {
	t.Helper()

	actual := 0

	if result.err != nil {
		exitErr, ok := result.err.(*exec.ExitError)
		if !ok {
			t.Fatalf(
				"command failed without exit code\nargs: %v\nerror: %v\nstdout:\n%s\nstderr:\n%s",
				result.args,
				result.err,
				result.stdout,
				result.stderr,
			)
		}

		actual = exitErr.ExitCode()
	}

	if actual != expected {
		t.Fatalf(
			"expected exit code %d, got %d\nargs: %v\nerror: %v\nstdout:\n%s\nstderr:\n%s",
			expected,
			actual,
			result.args,
			result.err,
			result.stdout,
			result.stderr,
		)
	}
}

func assertContains(t *testing.T, value string, expected string) {
	t.Helper()

	if !strings.Contains(value, expected) {
		t.Fatalf("expected output to contain %q\n\nOutput:\n%s", expected, value)
	}
}

func assertNotContains(t *testing.T, value string, unexpected string) {
	t.Helper()

	if strings.Contains(value, unexpected) {
		t.Fatalf("expected output not to contain %q\n\nOutput:\n%s", unexpected, value)
	}
}

func assertNotEmpty(t *testing.T, value string, message string) {
	t.Helper()

	if strings.TrimSpace(value) == "" {
		t.Fatal(message)
	}
}

func formatList(values []string) string {
	return "- " + strings.Join(values, "\n- ")
}

func assertFileExists(t *testing.T, path string) {
	t.Helper()

	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("expected file %s to exist: %v", path, err)
	}

	if info.IsDir() {
		t.Fatalf("expected file but found directory at %s", path)
	}
}

func assertDirExists(t *testing.T, path string) {
	t.Helper()

	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("expected directory %s to exist: %v", path, err)
	}

	if !info.IsDir() {
		t.Fatalf("expected directory but found file at %s", path)
	}
}

func assertDirNotExists(t *testing.T, path string) {
	t.Helper()

	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Fatalf("expected %s to not exist, but it does", path)
	}
}
