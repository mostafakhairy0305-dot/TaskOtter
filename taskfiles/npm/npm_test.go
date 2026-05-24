package npm_test

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
	Name                  string
	Args                  map[string]string
	MustDryRunWithoutArgs bool
	MustDryRunWithArgs    bool
	ExpectedDefaultTokens []string
	RequiresGroupOutput   bool
	RequiresPrompt        bool
	RequiresSummary       bool
}

var expectedPublicTasks = []PublicTaskSpec{
	{
		Name:               "add",
		Args:               map[string]string{"PACKAGES": "prettier"},
		MustDryRunWithArgs: true,
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
		Name:               "exec",
		Args:               map[string]string{"BINARY": "prettier"},
		MustDryRunWithArgs: true,
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

func TestReadmePublicTaskTableDoesNotDrift(t *testing.T) {
	content := readFile(t, filepath.Join(filepath.Dir(taskfilePath(t)), "README.md"))

	expected := expectedPublicTaskNames()
	actual := readmePublicTaskNames(t, content)

	if !slices.Equal(expected, actual) {
		t.Fatalf(
			"README public task table drift detected\n\nexpected:\n%s\n\nactual:\n%s\n\nKeep README.md Public Tasks aligned with expectedPublicTasks.",
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

			result := runTask(t, root, npmDryRunEnv(t), args...)

			assertExitCode(t, result, 0)

			out := result.combined()
			assertNotContains(t, strings.ToLower(out), "task not found")
			assertNotContains(t, strings.ToLower(out), "unknown task")
			assertNotContains(t, strings.ToLower(out), "cannot find")
			assertNotContains(t, strings.ToLower(out), "missing required")
		})
	}
}

func TestOptionalVersionTasksDryRunWithoutVersion(t *testing.T) {
	root := repoRoot(t)
	tf := loadTaskfile(t)

	for _, spec := range expectedPublicTasks {
		spec := spec

		t.Run(spec.Name, func(t *testing.T) {
			t.Parallel()
			if !spec.MustDryRunWithoutArgs {
				return
			}

			result := runTask(t, root, npmDryRunEnv(t), "--dry", "--yes", spec.Name)

			assertExitCode(t, result, 0)

			out := result.combined()
			assertNotContains(t, strings.ToLower(out), "missing required")
			assertNotContains(t, strings.ToLower(out), "required variable")

			if len(spec.ExpectedDefaultTokens) > 0 {
				vars := tf.Root.field("vars")
				varsText := nodeText(vars)
				for _, token := range spec.ExpectedDefaultTokens {
					assertContains(t, varsText, token)
				}
			}
		})
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

func TestRunTaskRequiresScriptVariable(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("stub npm tests target Unix-like systems")
	}
	t.Parallel()

	root := repoRoot(t)
	// Run without --dry so that requires: vars checks are enforced. SCRIPT
	// defaults to '' which triggers the requires check and must fail.
	result := runTask(t, root, npmDryRunEnv(t), "--yes", "run")

	if result.err == nil {
		t.Fatal("expected task run to fail without SCRIPT variable but it succeeded")
	}
}

func TestVersionTaskExitsSuccessfully(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("stub npm tests target Unix-like systems")
	}
	t.Parallel()

	root := repoRoot(t)
	result := runTask(t, root, npmDryRunEnv(t), "--yes", "version")

	assertExitCode(t, result, 0)
	assertContains(t, result.combined(), "stub")
}

func TestInstallTaskExitsSuccessfully(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("stub npm tests target Unix-like systems")
	}
	t.Parallel()

	root := repoRoot(t)
	result := runTask(t, root, npmDryRunEnv(t), "--yes", "install")

	assertExitCode(t, result, 0)
}

func TestCiTaskExitsSuccessfully(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("stub npm tests target Unix-like systems")
	}
	t.Parallel()

	root := repoRoot(t)
	result := runTask(t, root, npmDryRunEnv(t), "--yes", "ci")

	assertExitCode(t, result, 0)
}

func TestBuildTaskExitsSuccessfully(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("stub npm tests target Unix-like systems")
	}
	t.Parallel()

	root := repoRoot(t)
	result := runTask(t, root, npmDryRunEnv(t), "--yes", "build")

	assertExitCode(t, result, 0)
}

func TestRunTaskExitsSuccessfully(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("stub npm tests target Unix-like systems")
	}
	t.Parallel()

	root := repoRoot(t)
	result := runTask(t, root, npmDryRunEnv(t), "--yes", "run", "SCRIPT=build")

	assertExitCode(t, result, 0)
	assertContains(t, result.combined(), "build")
}

func TestCleanTaskSkipsWhenNodeModulesAbsent(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("stub npm tests target Unix-like systems")
	}
	t.Parallel()

	root := repoRoot(t)
	result := runTask(t, root, npmDryRunEnv(t), "--yes", "clean")

	assertExitCode(t, result, 0)
}

// TestOutdatedTaskExitsSuccessfully verifies that outdated succeeds even when
// the npm stub exits 0 — the task has ignore_error: true so it never fails.
func TestOutdatedTaskExitsSuccessfully(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("stub npm tests target Unix-like systems")
	}
	t.Parallel()

	root := repoRoot(t)
	result := runTask(t, root, npmDryRunEnv(t), "--yes", "outdated")

	assertExitCode(t, result, 0)
}

func TestOutdatedStrictTaskExitsSuccessfully(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("stub npm tests target Unix-like systems")
	}
	t.Parallel()

	root := repoRoot(t)
	result := runTask(t, root, npmDryRunEnv(t), "--yes", "outdated:strict")

	assertExitCode(t, result, 0)
}

func TestAuditReportTaskExitsSuccessfully(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("stub npm tests target Unix-like systems")
	}
	t.Parallel()

	root := repoRoot(t)
	result := runTask(t, root, npmDryRunEnv(t), "--yes", "audit:report")

	assertExitCode(t, result, 0)
}

// TestRunTaskForwardsCliArgs verifies that extra CLI arguments (after --) do
// not cause the run task to fail, and that CLI_ARGS is wired into the ARGS
// template so it would be forwarded to npm run.
func TestRunTaskForwardsCliArgs(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("stub npm tests target Unix-like systems")
	}
	t.Parallel()

	root := repoRoot(t)
	// Run with the stub — npm stub exits 0 regardless of args, so success here
	// confirms the CLI_ARGS plumbing does not break the task.
	result := runTask(t, root, npmDryRunEnv(t), "--yes", "run", "SCRIPT=test", "--", "--watch")

	assertExitCode(t, result, 0)
}

// TestRunTaskCliArgsWiredInYaml verifies that CLI_ARGS is wired into the
// internal _run:unix task so extra arguments after -- are forwarded to npm run.
func TestRunTaskCliArgsWiredInYaml(t *testing.T) {
	tf := loadTaskfile(t)

	task := mustTask(t, tf, "_run:unix")
	cmds := task.field("cmds")
	if cmds == nil {
		t.Fatal("_run:unix task has no cmds")
	}

	cmdText := nodeText(cmds)
	if !strings.Contains(cmdText, "CLI_ARGS") {
		t.Fatal("_run:unix cmds do not reference CLI_ARGS; extra arguments after -- will not be forwarded to npm run")
	}
}

// TestInvalidNodeManagerFails verifies the NODE_MANAGER precondition rejects
// unsupported values before attempting to invoke npm.
func TestInvalidNodeManagerFails(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("stub npm tests target Unix-like systems")
	}
	t.Parallel()

	root := repoRoot(t)
	result := runTask(t, root, npmDryRunEnv(t), "--yes", "install", "NODE_MANAGER=yarn")

	if result.err == nil {
		t.Fatal("expected task to fail with invalid NODE_MANAGER but it succeeded")
	}

	out := result.combined()
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

	root := repoRoot(t)
	result := runTask(t, root, npmDryRunEnv(t), "--yes", "dev")

	assertExitCode(t, result, 0)
}

// TestInstallFailsOutsideProjectRoot verifies that npm tasks fail clearly when
// package.json is absent from the working directory.
func TestInstallFailsOutsideProjectRoot(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("stub npm tests target Unix-like systems")
	}
	t.Parallel()

	taskfileDir := repoRoot(t)
	projectDir := t.TempDir()
	env := npmDryRunEnv(t)

	result := runTask(t, projectDir, env,
		"--taskfile", filepath.Join(taskfileDir, "Taskfile.yml"),
		"--yes", "install",
	)

	if result.err == nil {
		t.Fatal("expected task install to fail outside a project root but it succeeded")
	}

	out := result.combined()
	if !strings.Contains(strings.ToLower(out), "package.json") {
		t.Fatalf("expected error mentioning package.json, got:\n%s", out)
	}
}

// TestCiFailsWithoutLockfile verifies that the ci task fails clearly when
// package-lock.json is missing even if package.json is present.
func TestCiFailsWithoutLockfile(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("stub npm tests target Unix-like systems")
	}
	t.Parallel()

	taskfileDir := repoRoot(t)
	projectDir := t.TempDir()
	env := npmDryRunEnv(t)

	if err := os.WriteFile(
		filepath.Join(projectDir, "package.json"),
		[]byte(`{"name":"test","version":"1.0.0"}`),
		0644,
	); err != nil {
		t.Fatalf("failed to create package.json: %v", err)
	}

	result := runTask(t, projectDir, env,
		"--taskfile", filepath.Join(taskfileDir, "Taskfile.yml"),
		"--yes", "ci",
	)

	if result.err == nil {
		t.Fatal("expected task ci to fail without package-lock.json but it succeeded")
	}

	out := result.combined()
	if !strings.Contains(strings.ToLower(out), "package-lock.json") &&
		!strings.Contains(strings.ToLower(out), "lockfile") {
		t.Fatalf("expected error mentioning lockfile, got:\n%s", out)
	}
}

// TestRunTaskRejectsUnsafeScript verifies that SCRIPT values containing shell
// metacharacters are rejected before npm is invoked.
func TestRunTaskRejectsUnsafeScript(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("stub npm tests target Unix-like systems")
	}
	t.Parallel()

	root := repoRoot(t)
	result := runTask(t, root, npmDryRunEnv(t), "--yes", "run", "SCRIPT=dev; rm -rf /")

	if result.err == nil {
		t.Fatal("expected task run to reject unsafe SCRIPT but it succeeded")
	}

	out := result.combined()
	if !strings.Contains(strings.ToLower(out), "invalid") &&
		!strings.Contains(strings.ToLower(out), "script") {
		t.Fatalf("expected error about invalid SCRIPT characters, got:\n%s", out)
	}
}

func TestRealNpmFlowOnlyWhenExplicitlyEnabled(t *testing.T) {
	if os.Getenv("RUN_INSTALLER_TESTS") != "1" {
		t.Skip("set RUN_INSTALLER_TESTS=1 to run real npm install/build/test tests")
	}

	if runtime.GOOS == "windows" {
		t.Skip("real npm flow tests target Unix-like systems")
	}

	root := repoRoot(t)
	env := isolatedEnv(t)

	result := runTaskTimeout(t, root, env, 10*time.Minute, "--yes", "version")
	assertExitCode(t, result, 0)
	assertNotEmpty(t, result.combined(), "version output is empty")
}

// npmDryRunEnv returns an isolated environment with stub fnm, node, and npm
// binaries installed at $HOME/.local/bin. The stubs satisfy all precondition
// checks (fnm presence, command -v npm, command -v node) and respond to the
// subcommands used by task commands without performing real operations.
//
// fnm stub responds to:
//   - "env --shell bash" with a no-op comment (safe to eval in bash)
//   - "use [VERSION]"    with a status message and exit 0
//   - all other args     with exit 0
//
// This is sufficient for both dry-run tests (preconditions evaluate, commands
// print) and stub behavioral tests (commands run against stubs).
func npmDryRunEnv(t *testing.T) []string {
	t.Helper()

	env := isolatedEnv(t)
	home := envValue(env, "HOME")

	binDir := filepath.Join(home, ".local", "bin")
	if err := os.MkdirAll(binDir, 0755); err != nil {
		t.Fatalf("failed to create stub bin dir: %v", err)
	}

	// fnm stub: satisfies "test -f $HOME/.local/bin/fnm" precondition,
	// and returns a safe eval-able string for "fnm env --shell bash".
	fnmStub := "#!/usr/bin/env bash\n" +
		"case \"$1\" in\n" +
		"  --version) echo \"fnm 1.37.1 stub\" ;;\n" +
		"  env) echo \"# fnm env stub\" ;;\n" +
		"  use) echo \"Using Node.js stub\" ;;\n" +
		"  *) exit 0 ;;\n" +
		"esac\n"

	if err := os.WriteFile(filepath.Join(binDir, "fnm"), []byte(fnmStub), 0755); err != nil {
		t.Fatalf("failed to create stub fnm binary: %v", err)
	}

	nodeStub := "#!/usr/bin/env bash\n" +
		"case \"$1\" in\n" +
		"  --version) echo \"v20.11.0 stub\" ;;\n" +
		"  *) exit 0 ;;\n" +
		"esac\n"

	if err := os.WriteFile(filepath.Join(binDir, "node"), []byte(nodeStub), 0755); err != nil {
		t.Fatalf("failed to create stub node binary: %v", err)
	}

	npmStub := "#!/usr/bin/env bash\n" +
		"case \"$1\" in\n" +
		"  --version) echo \"10.9.0 stub\" ;;\n" +
		"  *) echo \"npm $* stub\"; exit 0 ;;\n" +
		"esac\n"

	if err := os.WriteFile(filepath.Join(binDir, "npm"), []byte(npmStub), 0755); err != nil {
		t.Fatalf("failed to create stub npm binary: %v", err)
	}

	corepackStub := "#!/usr/bin/env bash\n" +
		"echo \"corepack $* stub\"\n"

	if err := os.WriteFile(filepath.Join(binDir, "corepack"), []byte(corepackStub), 0755); err != nil {
		t.Fatalf("failed to create stub corepack binary: %v", err)
	}

	// Pre-populate .bashrc so the stub binaries are in PATH for any bash
	// subshell that task spawns (e.g. bash -c '...' in NODE_LOAD commands).
	bashrc := filepath.Join(home, ".bashrc")
	if err := os.WriteFile(bashrc, []byte("export PATH=\"$HOME/.local/bin:$PATH\"\n"), 0644); err != nil {
		t.Fatalf("failed to pre-populate shell profile: %v", err)
	}

	path := envValue(env, "PATH")
	env = setEnv(env, "PATH", binDir+":"+path)

	return env
}

// --- helpers (shared across all taskfile test packages) ---

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

func readmePublicTaskNames(t *testing.T, content string) []string {
	t.Helper()

	taskRow := regexp.MustCompile("^\\|\\s*`([^`]+)`\\s*\\|")
	lines := strings.Split(content, "\n")
	inPublicTasks := false
	names := []string{}

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "## Public Tasks" {
			inPublicTasks = true
			continue
		}

		if inPublicTasks && strings.HasPrefix(trimmed, "## ") {
			break
		}

		if !inPublicTasks {
			continue
		}

		if match := taskRow.FindStringSubmatch(trimmed); len(match) == 2 {
			names = append(names, match[1])
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
