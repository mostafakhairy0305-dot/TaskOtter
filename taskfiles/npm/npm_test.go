package npm_test

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
		Name:                "audit",
		MustDryRunWithArgs:  true,
		RequiresGroupOutput: true,
		RequiresSummary:     true,
	},
	{
		Name:                "audit:fix",
		MustDryRunWithArgs:  true,
		RequiresGroupOutput: true,
		RequiresSummary:     true,
	},
	{
		Name:                "audit:json",
		MustDryRunWithArgs:  true,
		RequiresGroupOutput: true,
		RequiresSummary:     true,
	},
	{
		Name:                "audit:report",
		MustDryRunWithArgs:  true,
		RequiresGroupOutput: true,
		RequiresSummary:     true,
	},
	{
		Name:                "build",
		MustDryRunWithArgs:  true,
		RequiresGroupOutput: true,
		RequiresSummary:     true,
	},
	{
		Name:                "cache:clean",
		MustDryRunWithArgs:  true,
		RequiresGroupOutput: true,
		RequiresSummary:     true,
	},
	{
		Name:                "ci",
		MustDryRunWithArgs:  true,
		RequiresGroupOutput: true,
		RequiresSummary:     true,
	},
	{
		Name:                "clean",
		MustDryRunWithArgs:  true,
		RequiresGroupOutput: true,
		RequiresPrompt:      true,
		RequiresSummary:     true,
	},
	{
		Name:                "clean:all",
		MustDryRunWithArgs:  true,
		RequiresGroupOutput: true,
		RequiresPrompt:      true,
		RequiresSummary:     true,
	},
	{
		Name:                "dev",
		MustDryRunWithArgs:  true,
		RequiresGroupOutput: true,
		RequiresSummary:     true,
	},
	{
		Name:                "exec",
		Args:                map[string]string{"BINARY": "prettier"},
		MustDryRunWithArgs:  true,
		RequiresGroupOutput: true,
	},
	{
		Name:                "doctor",
		MustDryRunWithArgs:  true,
		RequiresGroupOutput: true,
		RequiresSummary:     true,
	},
	{
		Name:                "format",
		MustDryRunWithArgs:  true,
		RequiresGroupOutput: true,
		RequiresSummary:     true,
	},
	{
		Name:                "install",
		MustDryRunWithArgs:  true,
		RequiresGroupOutput: true,
		RequiresSummary:     true,
	},
	{
		Name:                "lint",
		MustDryRunWithArgs:  true,
		RequiresGroupOutput: true,
		RequiresSummary:     true,
	},
	{
		Name:                "manager:pin",
		Args:                map[string]string{"PACKAGE_MANAGER_VERSION": "latest"},
		MustDryRunWithArgs:  true,
		RequiresGroupOutput: true,
		RequiresSummary:     true,
	},
	{
		Name:                "manager:setup",
		MustDryRunWithArgs:  true,
		RequiresGroupOutput: true,
		RequiresSummary:     true,
	},
	{
		Name:                "node:setup",
		MustDryRunWithArgs:  true,
		RequiresGroupOutput: true,
		RequiresSummary:     true,
	},
	{
		Name:                "outdated",
		MustDryRunWithArgs:  true,
		RequiresGroupOutput: true,
		RequiresSummary:     true,
	},
	{
		Name:                "outdated:strict",
		MustDryRunWithArgs:  true,
		RequiresGroupOutput: true,
		RequiresSummary:     true,
	},
	{
		Name:                "run",
		Args:                map[string]string{"SCRIPT": "build"},
		MustDryRunWithArgs:  true,
		RequiresGroupOutput: true,
		RequiresSummary:     true,
	},
	{
		Name:                "test",
		MustDryRunWithArgs:  true,
		RequiresGroupOutput: true,
		RequiresSummary:     true,
	},
	{
		Name:                "typecheck",
		MustDryRunWithArgs:  true,
		RequiresGroupOutput: true,
		RequiresSummary:     true,
	},
	{
		Name:                "update",
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

func TestReadmePublicTaskTableDoesNotDrift(t *testing.T) {
	content := tasktestutil.ReadFile(t, filepath.Join(filepath.Dir(tasktestutil.ModuleTaskfilePath(t)), "README.md"))
	expected := tasktestutil.ExpectedPublicTaskNames(expectedPublicTasks)
	actual := readmePublicTaskNames(content)
	if !slices.Equal(expected, actual) {
		t.Fatalf(
			"README public task table drift detected\n\nexpected:\n%s\n\nactual:\n%s\n\nKeep README.md Public Tasks aligned with expectedPublicTasks.",
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
			result := tasktestutil.RunTask(t, root, npmDryRunEnv(t), args...)
			tasktestutil.AssertExitCode(t, result, 0)
			out := strings.ToLower(result.Combined())
			tasktestutil.AssertNotContains(t, out, "task not found")
			tasktestutil.AssertNotContains(t, out, "unknown task")
			tasktestutil.AssertNotContains(t, out, "cannot find")
			tasktestutil.AssertNotContains(t, out, "missing required")
		})
	}
}

func TestOptionalVersionTasksDryRunWithoutVersion(t *testing.T) {
	root := tasktestutil.ModuleRoot(t)
	tf := tasktestutil.LoadTaskfile(t)
	for _, spec := range expectedPublicTasks {
		spec := spec
		t.Run(spec.Name, func(t *testing.T) {
			t.Parallel()
			if !spec.MustDryRunWithoutArgs {
				return
			}
			result := tasktestutil.RunTask(t, root, npmDryRunEnv(t), "--dry", "--yes", spec.Name)
			tasktestutil.AssertExitCode(t, result, 0)
			out := strings.ToLower(result.Combined())
			tasktestutil.AssertNotContains(t, out, "missing required")
			tasktestutil.AssertNotContains(t, out, "required variable")
			if len(spec.ExpectedDefaultTokens) > 0 {
				varsText := tasktestutil.NodeText(tf.Root.Field("vars"))
				for _, token := range spec.ExpectedDefaultTokens {
					tasktestutil.AssertContains(t, varsText, token)
				}
			}
		})
	}
}

func TestCommandsDoNotContainDangerousPatterns(t *testing.T) {
	tf := tasktestutil.LoadTaskfile(t)
	for taskName, task := range tf.Tasks {
		taskName := taskName
		for _, command := range tasktestutil.CollectCommandStrings(task.Node) {
			for _, pattern := range tasktestutil.DangerousCommandPatterns() {
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

func TestRunTaskRequiresScriptVariable(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("stub npm tests target Unix-like systems")
	}
	t.Parallel()
	result := tasktestutil.RunTask(t, tasktestutil.ModuleRoot(t), npmDryRunEnv(t), "--yes", "run")
	if result.Err == nil {
		t.Fatal("expected task run to fail without SCRIPT variable but it succeeded")
	}
}

func TestVersionTaskExitsSuccessfully(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("stub npm tests target Unix-like systems")
	}
	t.Parallel()
	result := tasktestutil.RunTask(t, tasktestutil.ModuleRoot(t), npmDryRunEnv(t), "--yes", "version")
	tasktestutil.AssertExitCode(t, result, 0)
	tasktestutil.AssertContains(t, result.Combined(), "stub")
}

func TestInstallTaskExitsSuccessfully(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("stub npm tests target Unix-like systems")
	}
	t.Parallel()
	result := tasktestutil.RunTask(t, tasktestutil.ModuleRoot(t), npmDryRunEnv(t), "--yes", "install")
	tasktestutil.AssertExitCode(t, result, 0)
}

func TestCiTaskExitsSuccessfully(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("stub npm tests target Unix-like systems")
	}
	t.Parallel()
	result := tasktestutil.RunTask(t, tasktestutil.ModuleRoot(t), npmDryRunEnv(t), "--yes", "ci")
	tasktestutil.AssertExitCode(t, result, 0)
}

func TestBuildTaskExitsSuccessfully(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("stub npm tests target Unix-like systems")
	}
	t.Parallel()
	result := tasktestutil.RunTask(t, tasktestutil.ModuleRoot(t), npmDryRunEnv(t), "--yes", "build")
	tasktestutil.AssertExitCode(t, result, 0)
}

func TestRunTaskExitsSuccessfully(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("stub npm tests target Unix-like systems")
	}
	t.Parallel()
	result := tasktestutil.RunTask(t, tasktestutil.ModuleRoot(t), npmDryRunEnv(t), "--yes", "run", "SCRIPT=build")
	tasktestutil.AssertExitCode(t, result, 0)
	tasktestutil.AssertContains(t, result.Combined(), "build")
}

func TestCleanTaskSkipsWhenNodeModulesAbsent(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("stub npm tests target Unix-like systems")
	}
	t.Parallel()
	result := tasktestutil.RunTask(t, tasktestutil.ModuleRoot(t), npmDryRunEnv(t), "--yes", "clean")
	tasktestutil.AssertExitCode(t, result, 0)
}

func TestOutdatedTaskExitsSuccessfully(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("stub npm tests target Unix-like systems")
	}
	t.Parallel()
	result := tasktestutil.RunTask(t, tasktestutil.ModuleRoot(t), npmDryRunEnv(t), "--yes", "outdated")
	tasktestutil.AssertExitCode(t, result, 0)
}

func TestOutdatedStrictTaskExitsSuccessfully(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("stub npm tests target Unix-like systems")
	}
	t.Parallel()
	result := tasktestutil.RunTask(t, tasktestutil.ModuleRoot(t), npmDryRunEnv(t), "--yes", "outdated:strict")
	tasktestutil.AssertExitCode(t, result, 0)
}

func TestAuditReportTaskExitsSuccessfully(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("stub npm tests target Unix-like systems")
	}
	t.Parallel()
	result := tasktestutil.RunTask(t, tasktestutil.ModuleRoot(t), npmDryRunEnv(t), "--yes", "audit:report")
	tasktestutil.AssertExitCode(t, result, 0)
}

func TestRunTaskForwardsCliArgs(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("stub npm tests target Unix-like systems")
	}
	t.Parallel()
	result := tasktestutil.RunTask(t, tasktestutil.ModuleRoot(t), npmDryRunEnv(t), "--yes", "run", "SCRIPT=test", "--", "--watch")
	tasktestutil.AssertExitCode(t, result, 0)
}

func TestRunTaskCliArgsWiredInYaml(t *testing.T) {
	tf := tasktestutil.LoadTaskfile(t)
	task := tasktestutil.MustTask(t, tf, "_run:unix")
	cmds := task.Field("cmds")
	if cmds == nil {
		t.Fatal("_run:unix task has no cmds")
	}
	if !strings.Contains(tasktestutil.NodeText(cmds), "CLI_ARGS") {
		t.Fatal("_run:unix cmds do not reference CLI_ARGS; extra arguments after -- will not be forwarded to npm run")
	}
}

func TestInvalidNodeManagerFails(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("stub npm tests target Unix-like systems")
	}
	t.Parallel()
	result := tasktestutil.RunTask(t, tasktestutil.ModuleRoot(t), npmDryRunEnv(t), "--yes", "install", "NODE_MANAGER=yarn")
	if result.Err == nil {
		t.Fatal("expected task to fail with invalid NODE_MANAGER but it succeeded")
	}
	out := result.Combined()
	if !strings.Contains(strings.ToLower(out), "node_manager") &&
		!strings.Contains(strings.ToLower(out), "fnm") &&
		!strings.Contains(strings.ToLower(out), "nvm") {
		t.Fatalf("expected error message to mention NODE_MANAGER or valid values, got:\n%s", out)
	}
}

func TestDevTaskExitsSuccessfully(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("stub npm tests target Unix-like systems")
	}
	t.Parallel()
	result := tasktestutil.RunTask(t, tasktestutil.ModuleRoot(t), npmDryRunEnv(t), "--yes", "dev")
	tasktestutil.AssertExitCode(t, result, 0)
}

func TestInstallFailsOutsideProjectRoot(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("stub npm tests target Unix-like systems")
	}
	t.Parallel()
	taskfileDir := tasktestutil.ModuleRoot(t)
	projectDir := t.TempDir()
	result := tasktestutil.RunTask(t, projectDir, npmDryRunEnv(t),
		"--taskfile", filepath.Join(taskfileDir, "Taskfile.yml"),
		"--yes", "install",
	)
	if result.Err == nil {
		t.Fatal("expected task install to fail outside a project root but it succeeded")
	}
	if !strings.Contains(strings.ToLower(result.Combined()), "package.json") {
		t.Fatalf("expected error mentioning package.json, got:\n%s", result.Combined())
	}
}

func TestCiFailsWithoutLockfile(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("stub npm tests target Unix-like systems")
	}
	t.Parallel()
	taskfileDir := tasktestutil.ModuleRoot(t)
	projectDir := t.TempDir()
	if err := os.WriteFile(
		filepath.Join(projectDir, "package.json"),
		[]byte(`{"name":"test","version":"1.0.0"}`),
		0644,
	); err != nil {
		t.Fatalf("failed to create package.json: %v", err)
	}
	result := tasktestutil.RunTask(t, projectDir, npmDryRunEnv(t),
		"--taskfile", filepath.Join(taskfileDir, "Taskfile.yml"),
		"--yes", "ci",
	)
	if result.Err == nil {
		t.Fatal("expected task ci to fail without package-lock.json but it succeeded")
	}
	out := strings.ToLower(result.Combined())
	if !strings.Contains(out, "package-lock.json") && !strings.Contains(out, "lockfile") {
		t.Fatalf("expected error mentioning lockfile, got:\n%s", result.Combined())
	}
}

func TestRunTaskRejectsUnsafeScript(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("stub npm tests target Unix-like systems")
	}
	t.Parallel()
	result := tasktestutil.RunTask(t, tasktestutil.ModuleRoot(t), npmDryRunEnv(t), "--yes", "run", "SCRIPT=dev; rm -rf /")
	if result.Err == nil {
		t.Fatal("expected task run to reject unsafe SCRIPT but it succeeded")
	}
	out := strings.ToLower(result.Combined())
	if !strings.Contains(out, "invalid") && !strings.Contains(out, "script") {
		t.Fatalf("expected error about invalid SCRIPT characters, got:\n%s", result.Combined())
	}
}

func TestRealNpmFlowOnlyWhenExplicitlyEnabled(t *testing.T) {
	if os.Getenv("RUN_INSTALLER_TESTS") != "1" {
		t.Skip("set RUN_INSTALLER_TESTS=1 to run real npm install/build/test tests")
	}
	if runtime.GOOS == "windows" {
		t.Skip("real npm flow tests target Unix-like systems")
	}
	root := tasktestutil.ModuleRoot(t)
	env := tasktestutil.IsolatedEnv(t)
	result := tasktestutil.RunTaskTimeout(t, root, env, 10*time.Minute, "--yes", "version")
	tasktestutil.AssertExitCode(t, result, 0)
	tasktestutil.AssertNotEmpty(t, result.Combined(), "version output is empty")
}

// npmDryRunEnv returns an isolated environment with stub fnm, node, npm, and
// corepack binaries so all preconditions pass without real installations.
func npmDryRunEnv(t *testing.T) []string {
	t.Helper()

	env := tasktestutil.IsolatedEnv(t)
	home := tasktestutil.EnvValue(env, "HOME")

	binDir := filepath.Join(home, ".local", "bin")
	if err := os.MkdirAll(binDir, 0755); err != nil {
		t.Fatalf("failed to create stub bin dir: %v", err)
	}

	tasktestutil.WriteStub(t, binDir, "fnm",
		"#!/usr/bin/env bash\ncase \"$1\" in\n  --version) echo \"fnm 1.37.1 stub\" ;;\n  env) echo \"# fnm env stub\" ;;\n  use) echo \"Using Node.js stub\" ;;\n  *) exit 0 ;;\nesac\n",
	)
	tasktestutil.WriteStub(t, binDir, "node",
		"#!/usr/bin/env bash\ncase \"$1\" in\n  --version) echo \"v20.11.0 stub\" ;;\n  *) exit 0 ;;\nesac\n",
	)
	tasktestutil.WriteStub(t, binDir, "npm",
		"#!/usr/bin/env bash\ncase \"$1\" in\n  --version) echo \"10.9.0 stub\" ;;\n  *) echo \"npm $* stub\"; exit 0 ;;\nesac\n",
	)
	tasktestutil.WriteStub(t, binDir, "corepack", "#!/usr/bin/env bash\necho \"corepack $* stub\"\n")

	bashrc := filepath.Join(home, ".bashrc")
	if err := os.WriteFile(bashrc, []byte("export PATH=\"$HOME/.local/bin:$PATH\"\n"), 0644); err != nil {
		t.Fatalf("failed to pre-populate shell profile: %v", err)
	}

	path := tasktestutil.EnvValue(env, "PATH")
	return tasktestutil.SetEnv(env, "PATH", binDir+":"+path)
}

// readmePublicTaskNames parses the npm README and returns sorted task names
// from the Public Tasks table.
func readmePublicTaskNames(content string) []string {
	return tasktestutil.ReadmePublicTaskNames(content)
}
