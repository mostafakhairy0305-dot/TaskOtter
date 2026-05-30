package bun_test

import (
	"os"
	"path/filepath"
	"runtime"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/mostafakhairy0305-dot/TaskOtter/internal/tasktestutil"
	"gopkg.in/yaml.v3"
)

var expectedPublicTasks = []tasktestutil.PublicTaskSpec{
	{
		Name:                "add",
		Args:                map[string]string{"PACKAGES": "prettier"},
		MustDryRunWithArgs:  true,
		RequiresGroupOutput: true,
	},
	{
		Name:                "exec",
		Args:                map[string]string{"BINARY": "prettier"},
		MustDryRunWithArgs:  true,
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
		Name:            "version",
		MustDryRunWithArgs: true,
		RequiresSummary: true,
	},
}

func TestTaskBinaryIsAvailable(t *testing.T) {
	root := tasktestutil.ModuleRoot(t)
	result := tasktestutil.RunTask(t, root, nil, "--version")
	tasktestutil.AssertExitCode(t, result, 0)
	tasktestutil.AssertNotEmpty(t, result.Combined(), "task --version output is empty")
}

func TestTaskfileYamlIsCleanAndValid(t *testing.T) {
	path := tasktestutil.ModuleTaskfilePath(t)
	content := tasktestutil.ReadFile(t, path)
	tasktestutil.AssertTextFileClean(t, path, content)

	var doc yaml.Node
	if err := yaml.Unmarshal([]byte(content), &doc); err != nil {
		t.Fatalf("Taskfile YAML is invalid: %v", err)
	}
	tasktestutil.AssertNoDuplicateMappingKeys(t, &doc, "Taskfile")
	tasktestutil.AssertNoYamlAliases(t, &doc, "Taskfile")

	root := tasktestutil.DocumentRoot(t, &doc)
	version := tasktestutil.ScalarField(root, "version")
	if version != "3" && !strings.HasPrefix(version, "3.") {
		t.Fatalf("Taskfile version must be 3 or 3.x, got %q", version)
	}
	tasks := tasktestutil.MappingField(root, "tasks")
	if tasks == nil || len(tasks.Content) == 0 {
		t.Fatal("Taskfile must contain non-empty tasks map")
	}
}

func TestTaskCliCanLoadTaskfile(t *testing.T) {
	root := tasktestutil.ModuleRoot(t)
	for _, args := range [][]string{
		{"--list"},
		{"--list-all"},
		{"--list-all", "--sort", "alphanumeric"},
		{"--list-all", "--json"},
	} {
		args := args
		t.Run(strings.Join(args, " "), func(t *testing.T) {
			result := tasktestutil.RunTask(t, root, tasktestutil.IsolatedEnv(t), args...)
			tasktestutil.AssertExitCode(t, result, 0)
			tasktestutil.AssertNotContains(t, strings.ToLower(result.Combined()), "taskfile does not exist")
			tasktestutil.AssertNotContains(t, strings.ToLower(result.Combined()), "unknown")
		})
	}
}

func TestTaskListAllJsonIsValid(t *testing.T) {
	root := tasktestutil.ModuleRoot(t)
	result := tasktestutil.RunTask(t, root, tasktestutil.IsolatedEnv(t), "--list-all", "--json")
	tasktestutil.AssertExitCode(t, result, 0)
	if err := tasktestutil.ValidateJSON(result.Stdout); err != nil {
		t.Fatalf("task --list-all --json produced invalid JSON:\n%s\nerror: %v", result.Stdout, err)
	}
}

func TestPublicApiDoesNotDrift(t *testing.T) {
	tf := tasktestutil.LoadTaskfile(t)
	expected := tasktestutil.ExpectedPublicTaskNames(expectedPublicTasks)
	actual := tasktestutil.PublicTaskNamesFromTaskfile(t, tf)
	if !slices.Equal(expected, actual) {
		t.Fatalf(
			"public Taskfile API drift detected\n\nexpected:\n%s\n\nactual:\n%s\n\nFix either the Taskfile public tasks or expectedPublicTasks in the test.",
			tasktestutil.FormatList(expected), tasktestutil.FormatList(actual),
		)
	}
}

func TestEveryTaskIsEitherPublicOrInternal(t *testing.T) {
	tf := tasktestutil.LoadTaskfile(t)
	for name, task := range tf.Tasks {
		name, task := name, task
		t.Run(name, func(t *testing.T) {
			if strings.HasPrefix(name, "_") || task.BoolField("internal") {
				return
			}
			if task.StringField("desc") == "" {
				t.Fatalf("task %q is not internal and has no desc. Either add desc/summary or mark it internal: true", name)
			}
		})
	}
}

func TestPublicTasksHaveMetadata(t *testing.T) {
	tf := tasktestutil.LoadTaskfile(t)
	for _, spec := range expectedPublicTasks {
		spec := spec
		t.Run(spec.Name, func(t *testing.T) {
			t.Parallel()
			task := tasktestutil.MustTask(t, tf, spec.Name)
			if task.Node.Kind != yaml.MappingNode {
				t.Fatalf("public task %q must use full mapping syntax, not short syntax", spec.Name)
			}
			desc := task.StringField("desc")
			summary := task.StringField("summary")
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
			tasktestutil.AssertNoPlaceholderText(t, spec.Name, desc)
			tasktestutil.AssertNoPlaceholderText(t, spec.Name, summary)
		})
	}
}

func TestDestructivePublicTasksHavePrompt(t *testing.T) {
	tf := tasktestutil.LoadTaskfile(t)
	for _, spec := range expectedPublicTasks {
		spec := spec
		t.Run(spec.Name, func(t *testing.T) {
			t.Parallel()
			if !spec.RequiresPrompt {
				return
			}
			task := tasktestutil.MustTask(t, tf, spec.Name)
			prompt := task.Field("prompt")
			if prompt == nil || tasktestutil.NodeText(prompt) == "" {
				t.Fatalf("destructive task %q must have a non-empty prompt", spec.Name)
			}
			text := strings.ToLower(tasktestutil.NodeText(prompt))
			if !strings.Contains(text, "sure") && !strings.Contains(text, "confirm") &&
				!strings.Contains(text, "remove") && !strings.Contains(text, "uninstall") &&
				!strings.Contains(text, "delete") && !strings.Contains(text, "continue") {
				t.Fatalf("prompt for task %q does not look explicit enough:\n%s", spec.Name, tasktestutil.NodeText(prompt))
			}
		})
	}
}

func TestInstallTasksUseGithubGroupOutput(t *testing.T) {
	tf := tasktestutil.LoadTaskfile(t)
	for _, spec := range expectedPublicTasks {
		spec := spec
		t.Run(spec.Name, func(t *testing.T) {
			t.Parallel()
			if !spec.RequiresGroupOutput {
				return
			}
			task := tasktestutil.MustTask(t, tf, spec.Name)
			outputNode := task.Field("output")
			if outputNode == nil {
				outputNode = tf.Root.Field("output")
			}
			tasktestutil.AssertGithubGroupOutput(t, spec.Name, outputNode)
		})
	}
}

func TestPublicTasksHaveCommands(t *testing.T) {
	tf := tasktestutil.LoadTaskfile(t)
	for _, spec := range expectedPublicTasks {
		spec := spec
		t.Run(spec.Name, func(t *testing.T) {
			t.Parallel()
			task := tasktestutil.MustTask(t, tf, spec.Name)
			if tasktestutil.IsEmptyNode(task.Field("cmds")) && tasktestutil.IsEmptyNode(task.Field("deps")) {
				t.Fatalf("public task %q must have cmds or deps", spec.Name)
			}
		})
	}
}

func TestTaskSummariesWork(t *testing.T) {
	root := tasktestutil.ModuleRoot(t)
	for _, spec := range expectedPublicTasks {
		spec := spec
		t.Run(spec.Name, func(t *testing.T) {
			t.Parallel()
			if !spec.RequiresSummary {
				return
			}
			result := tasktestutil.RunTask(t, root, tasktestutil.IsolatedEnv(t), "--summary", spec.Name)
			tasktestutil.AssertExitCode(t, result, 0)
			out := result.Combined()
			tasktestutil.AssertContains(t, out, spec.Name)
			tasktestutil.AssertNotContains(t, strings.ToLower(out), "task not found")
			tasktestutil.AssertNotContains(t, strings.ToLower(out), "unknown task")
			tasktestutil.AssertNotContains(t, strings.ToLower(out), "no summary")
		})
	}
}

func TestPublicTasksDryRunWithExpectedArgs(t *testing.T) {
	root := tasktestutil.ModuleRoot(t)
	for _, spec := range expectedPublicTasks {
		spec := spec
		t.Run(spec.Name, func(t *testing.T) {
			t.Parallel()
			if !spec.MustDryRunWithArgs {
				return
			}
			args := append([]string{"--dry", "--yes", spec.Name}, tasktestutil.TaskArgs(spec.Args)...)
			result := tasktestutil.RunTask(t, root, bunDryRunEnv(t), args...)
			tasktestutil.AssertExitCode(t, result, 0)
			out := strings.ToLower(result.Combined())
			tasktestutil.AssertNotContains(t, out, "task not found")
			tasktestutil.AssertNotContains(t, out, "unknown task")
			tasktestutil.AssertNotContains(t, out, "cannot find")
			tasktestutil.AssertNotContains(t, out, "missing required")
		})
	}
}

func TestUndoPairsExist(t *testing.T) {
	tf := tasktestutil.LoadTaskfile(t)
	for task, undo := range map[string]string{"install": "install:undo"} {
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
	if !tasktestutil.HasAlias(undoTask, "uninstall") {
		t.Fatal("task install:undo is missing alias uninstall")
	}
}

func TestAliasesDryRun(t *testing.T) {
	root := tasktestutil.ModuleRoot(t)
	for _, tc := range []struct {
		alias string
		args  []string
	}{
		{"uninstall", nil},
	} {
		tc := tc
		t.Run(tc.alias, func(t *testing.T) {
			t.Parallel()
			args := append([]string{"--dry", "--yes", tc.alias}, tc.args...)
			result := tasktestutil.RunTask(t, root, bunDryRunEnv(t), args...)
			tasktestutil.AssertExitCode(t, result, 0)
			out := strings.ToLower(result.Combined())
			tasktestutil.AssertNotContains(t, out, "task not found")
			tasktestutil.AssertNotContains(t, out, "unknown task")
		})
	}
}

func TestReferencedScriptsExist(t *testing.T) {
	root := tasktestutil.ModuleRoot(t)
	tf := tasktestutil.LoadTaskfile(t)
	for taskName, task := range tf.Tasks {
		taskName, task := taskName, task
		for _, command := range tasktestutil.CollectCommandStrings(task.Node) {
			command := command
			t.Run(taskName, func(t *testing.T) {
				t.Parallel()
				for _, scriptPath := range tasktestutil.ReferencedLocalShellScripts(command) {
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
	dangerousPatterns := tasktestutil.DangerousCommandPatterns()
	tf := tasktestutil.LoadTaskfile(t)
	for taskName, task := range tf.Tasks {
		taskName := taskName
		for _, command := range tasktestutil.CollectCommandStrings(task.Node) {
			for _, pattern := range dangerousPatterns {
				if pattern.MatchString(command) {
					t.Fatalf("task %q contains dangerous command pattern %q:\n%s", taskName, pattern.String(), command)
				}
			}
		}
	}
}

func TestNoPlaceholderTextInTaskfile(t *testing.T) {
	content := tasktestutil.ReadFile(t, tasktestutil.ModuleTaskfilePath(t))
	upper := strings.ToUpper(content)
	for _, p := range []string{"TODO", "FIXME", "CHANGEME", "REPLACE_ME", "YOUR VALUE HERE", "LOREM IPSUM"} {
		if strings.Contains(upper, p) {
			t.Fatalf("Taskfile contains placeholder text: %s", p)
		}
	}
}

func TestVersionTaskExitsSuccessfully(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("stub bun tests target Unix-like systems")
	}
	t.Parallel()
	result := tasktestutil.RunTask(t, tasktestutil.ModuleRoot(t), bunDryRunEnv(t), "--yes", "version")
	tasktestutil.AssertExitCode(t, result, 0)
}

func TestVersionTaskPrintsBunVersion(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("stub bun tests target Unix-like systems")
	}
	t.Parallel()
	result := tasktestutil.RunTask(t, tasktestutil.ModuleRoot(t), bunDryRunEnv(t), "--yes", "version")
	tasktestutil.AssertExitCode(t, result, 0)
	tasktestutil.AssertContains(t, result.Combined(), "1.")
}

func TestInstallIsIdempotentWithStubBun(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("stub bun tests target Unix-like systems")
	}
	t.Parallel()
	root := tasktestutil.ModuleRoot(t)
	env := bunDryRunEnv(t)
	tasktestutil.AssertExitCode(t, tasktestutil.RunTask(t, root, env, "--yes", "install"), 0)
	tasktestutil.AssertExitCode(t, tasktestutil.RunTask(t, root, env, "--yes", "install"), 0)
}

func TestInstallSkipsWhenBunIsAlreadyPresent(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("stub bun tests target Unix-like systems")
	}
	t.Parallel()
	result := tasktestutil.RunTask(t, tasktestutil.ModuleRoot(t), bunDryRunEnv(t), "--yes", "install")
	tasktestutil.AssertExitCode(t, result, 0)
	tasktestutil.AssertNotContains(t, result.Combined(), "Installing Bun")
}

func TestInstallUndoRemovesBunDir(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("stub bun tests target Unix-like systems")
	}
	t.Parallel()
	root := tasktestutil.ModuleRoot(t)
	env := bunDryRunEnv(t)
	bunDir := filepath.Join(tasktestutil.EnvValue(env, "HOME"), ".bun")
	tasktestutil.AssertDirExists(t, bunDir)
	tasktestutil.AssertExitCode(t, tasktestutil.RunTask(t, root, env, "--yes", "install:undo"), 0)
	tasktestutil.AssertDirNotExists(t, bunDir)
}

func TestInstallUndoIsIdempotent(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("stub bun tests target Unix-like systems")
	}
	t.Parallel()
	root := tasktestutil.ModuleRoot(t)
	env := bunDryRunEnv(t)
	tasktestutil.AssertExitCode(t, tasktestutil.RunTask(t, root, env, "--yes", "install:undo"), 0)
	tasktestutil.AssertExitCode(t, tasktestutil.RunTask(t, root, env, "--yes", "install:undo"), 0)
}

func TestUpgradeExitsSuccessfully(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("stub bun tests target Unix-like systems")
	}
	t.Parallel()
	result := tasktestutil.RunTask(t, tasktestutil.ModuleRoot(t), bunDryRunEnv(t), "--yes", "upgrade")
	tasktestutil.AssertExitCode(t, result, 0)
}

func TestUpgradeCanaryExitsSuccessfully(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("stub bun tests target Unix-like systems")
	}
	t.Parallel()
	result := tasktestutil.RunTask(t, tasktestutil.ModuleRoot(t), bunDryRunEnv(t), "--yes", "upgrade:canary")
	tasktestutil.AssertExitCode(t, result, 0)
}

func TestUpgradeStableExitsSuccessfully(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("stub bun tests target Unix-like systems")
	}
	t.Parallel()
	result := tasktestutil.RunTask(t, tasktestutil.ModuleRoot(t), bunDryRunEnv(t), "--yes", "upgrade:stable")
	tasktestutil.AssertExitCode(t, result, 0)
}

func TestRealInstallerFlowOnlyWhenExplicitlyEnabled(t *testing.T) {
	if os.Getenv("RUN_INSTALLER_TESTS") != "1" {
		t.Skip("set RUN_INSTALLER_TESTS=1 to run real install/uninstall tests")
	}
	if runtime.GOOS == "windows" {
		t.Skip("real bun installer tests are intended for Unix-like systems")
	}
	root := tasktestutil.ModuleRoot(t)
	env := tasktestutil.IsolatedEnv(t)
	home := tasktestutil.EnvValue(env, "HOME")
	bunBin := filepath.Join(home, ".bun", "bin", "bun")

	tasktestutil.AssertExitCode(t, tasktestutil.RunTaskTimeout(t, root, env, 10*time.Minute, "--yes", "install"), 0)
	tasktestutil.AssertFileExists(t, bunBin)
	tasktestutil.AssertExitCode(t, tasktestutil.RunTaskTimeout(t, root, env, 10*time.Minute, "version"), 0)
	tasktestutil.AssertExitCode(t, tasktestutil.RunTaskTimeout(t, root, env, 10*time.Minute, "--yes", "install:undo"), 0)
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
	root := tasktestutil.ModuleRoot(t)
	env := tasktestutil.IsolatedEnv(t)
	home := tasktestutil.EnvValue(env, "HOME")
	bunBin := filepath.Join(home, ".bun", "bin", "bun")

	step := func(name string, fn func(t *testing.T)) {
		t.Helper()
		t.Run(name, fn)
		if t.Failed() {
			t.FailNow()
		}
	}
	run := func(t *testing.T, args ...string) tasktestutil.CommandResult {
		t.Helper()
		result := tasktestutil.RunTaskTimeout(t, root, env, 10*time.Minute, args...)
		tasktestutil.AssertExitCode(t, result, 0)
		return result
	}

	step("install — bun binary is present on disk", func(t *testing.T) {
		run(t, "--yes", "install")
		tasktestutil.AssertFileExists(t, bunBin)
	})
	step("version — bun version string is printed", func(t *testing.T) {
		result := run(t, "version")
		tasktestutil.AssertNotEmpty(t, result.Combined(), "version output is empty")
	})
	step("upgrade — bun upgrades without error", func(t *testing.T) {
		run(t, "--yes", "upgrade")
		tasktestutil.AssertFileExists(t, bunBin)
	})
	step("upgrade:canary — bun switches to canary without error", func(t *testing.T) {
		run(t, "--yes", "upgrade:canary")
		tasktestutil.AssertFileExists(t, bunBin)
	})
	step("upgrade:stable — bun switches back to stable without error", func(t *testing.T) {
		run(t, "--yes", "upgrade:stable")
		tasktestutil.AssertFileExists(t, bunBin)
	})
	step("install:undo — .bun directory is removed", func(t *testing.T) {
		run(t, "--yes", "install:undo")
		tasktestutil.AssertDirNotExists(t, filepath.Join(home, ".bun"))
	})
}

// bunDryRunEnv returns an isolated environment with a stub bun binary that
// satisfies precondition checks without performing real operations.
func bunDryRunEnv(t *testing.T) []string {
	t.Helper()

	env := tasktestutil.IsolatedEnv(t)
	home := tasktestutil.EnvValue(env, "HOME")

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

	if err := os.WriteFile(filepath.Join(bunBinDir, "bun"), []byte(stub), 0755); err != nil {
		t.Fatalf("failed to create stub bun binary: %v", err)
	}

	path := tasktestutil.EnvValue(env, "PATH")
	return tasktestutil.SetEnv(env, "PATH", bunBinDir+":"+path)
}
