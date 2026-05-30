package fnm_test

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
		Name:               "ls",
		MustDryRunWithArgs: true,
		RequiresSummary:    true,
	},
	{
		Name:                  "node:install",
		Args:                  map[string]string{"VERSION": "24.0.0"},
		MustDryRunWithoutArgs: true,
		MustDryRunWithArgs:    true,
		ExpectedDefaultTokens: []string{"--lts"},
		RequiresGroupOutput:   true,
		RequiresSummary:       true,
	},
	{
		Name:                "node:uninstall",
		Args:                map[string]string{"VERSION": "24.0.0"},
		MustDryRunWithArgs:  true,
		RequiresGroupOutput: true,
		RequiresPrompt:      true,
		RequiresSummary:     true,
	},
	{
		Name:                  "node:use",
		Args:                  map[string]string{"VERSION": "24.0.0"},
		MustDryRunWithoutArgs: true,
		MustDryRunWithArgs:    true,
		ExpectedDefaultTokens: []string{"--lts"},
		RequiresSummary:       true,
	},
	{
		Name:               "node:version",
		MustDryRunWithArgs: true,
		RequiresSummary:    true,
	},
	{
		Name:                "shell:setup",
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
			result := tasktestutil.RunTask(t, root, fnmDryRunEnv(t), args...)
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
			result := tasktestutil.RunTask(t, root, fnmDryRunEnv(t), "--dry", "--yes", spec.Name)
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
	for _, p := range []struct{ task, undoAlias, undoTarget string }{
		{"node:install", "node:install:undo", "node:uninstall"},
		{"node:uninstall", "node:uninstall:undo", "node:install"},
	} {
		if _, ok := tf.Tasks[p.task]; !ok {
			t.Fatalf("task %q is missing", p.task)
		}
		target, ok := tf.Tasks[p.undoTarget]
		if !ok {
			t.Fatalf("undo target %q is missing for task %q", p.undoTarget, p.task)
		}
		if !tasktestutil.HasAlias(target, p.undoAlias) {
			t.Fatalf("task %q is missing alias %q (undo of %q)", p.undoTarget, p.undoAlias, p.task)
		}
	}
}

func TestAliasesDryRun(t *testing.T) {
	root := tasktestutil.ModuleRoot(t)
	for _, tc := range []struct {
		alias string
		args  []string
	}{
		{"list", nil},
		{"uninstall", nil},
		{"node:install:undo", []string{"VERSION=24.0.0"}},
		{"node:uninstall:undo", []string{"VERSION=24.0.0"}},
		{"node:current", nil},
		{"node:active", nil},
	} {
		tc := tc
		t.Run(tc.alias, func(t *testing.T) {
			t.Parallel()
			args := append([]string{"--dry", "--yes", tc.alias}, tc.args...)
			result := tasktestutil.RunTask(t, root, fnmDryRunEnv(t), args...)
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

func TestVersionTaskExitsSuccessfully(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("stub fnm tests target Unix-like systems")
	}
	t.Parallel()
	result := tasktestutil.RunTask(t, tasktestutil.ModuleRoot(t), fnmDryRunEnv(t), "--yes", "version")
	tasktestutil.AssertExitCode(t, result, 0)
}

func TestLsTaskExitsSuccessfully(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("stub fnm tests target Unix-like systems")
	}
	t.Parallel()
	result := tasktestutil.RunTask(t, tasktestutil.ModuleRoot(t), fnmDryRunEnv(t), "--yes", "ls")
	tasktestutil.AssertExitCode(t, result, 0)
}

func TestInstallIsIdempotentWithStubFnm(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("stub fnm tests target Unix-like systems")
	}
	t.Parallel()
	root := tasktestutil.ModuleRoot(t)
	env := fnmDryRunEnv(t)
	tasktestutil.AssertExitCode(t, tasktestutil.RunTask(t, root, env, "--yes", "install"), 0)
	tasktestutil.AssertExitCode(t, tasktestutil.RunTask(t, root, env, "--yes", "install"), 0)
}

func TestInstallUndoRemovesFnmBinary(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("stub fnm tests target Unix-like systems")
	}
	t.Parallel()
	root := tasktestutil.ModuleRoot(t)
	env := fnmDryRunEnv(t)
	stubBin := filepath.Join(tasktestutil.EnvValue(env, "HOME"), ".local", "bin", "fnm")
	tasktestutil.AssertFileExists(t, stubBin)
	tasktestutil.AssertExitCode(t, tasktestutil.RunTask(t, root, env, "--yes", "install:undo"), 0)
	if _, err := os.Stat(stubBin); !os.IsNotExist(err) {
		t.Fatalf("expected fnm binary at %s to be removed after install:undo", stubBin)
	}
}

func TestNodeInstallWithVersionPrintsVersionInOutput(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("stub fnm tests target Unix-like systems")
	}
	t.Parallel()
	result := tasktestutil.RunTask(t, tasktestutil.ModuleRoot(t), fnmDryRunEnv(t), "--yes", "node:install", "VERSION=18.0.0")
	tasktestutil.AssertExitCode(t, result, 0)
	tasktestutil.AssertContains(t, result.Combined(), "18.0.0")
}

func TestNodeInstallDefaultVersionUsesLts(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("stub fnm tests target Unix-like systems")
	}
	t.Parallel()
	result := tasktestutil.RunTask(t, tasktestutil.ModuleRoot(t), fnmDryRunEnv(t), "--yes", "node:install")
	tasktestutil.AssertExitCode(t, result, 0)
	tasktestutil.AssertContains(t, result.Combined(), "--lts")
}

func TestNodeUninstallWithInstalledVersionPrintsVersionInOutput(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("stub fnm tests target Unix-like systems")
	}
	t.Parallel()
	result := tasktestutil.RunTask(t, tasktestutil.ModuleRoot(t), fnmDryRunEnv(t), "--yes", "node:uninstall", "VERSION=18.0.0")
	tasktestutil.AssertExitCode(t, result, 0)
	tasktestutil.AssertContains(t, result.Combined(), "18.0.0")
}

func TestShellSetupAddsActivationToProfile(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("stub fnm tests target Unix-like systems")
	}
	t.Parallel()
	root := tasktestutil.ModuleRoot(t)
	env := fnmFreshProfileEnv(t)
	home := tasktestutil.EnvValue(env, "HOME")
	tasktestutil.AssertExitCode(t, tasktestutil.RunTask(t, root, env, "--yes", "shell:setup"), 0)
	if !profileContains(home, "fnm env") {
		t.Fatal("expected at least one shell profile to contain 'fnm env' after task shell:setup")
	}
}

func TestShellSetupIsIdempotent(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("stub fnm tests target Unix-like systems")
	}
	t.Parallel()
	root := tasktestutil.ModuleRoot(t)
	env := fnmFreshProfileEnv(t)
	home := tasktestutil.EnvValue(env, "HOME")
	tasktestutil.AssertExitCode(t, tasktestutil.RunTask(t, root, env, "--yes", "shell:setup"), 0)
	tasktestutil.AssertExitCode(t, tasktestutil.RunTask(t, root, env, "--yes", "shell:setup"), 0)
	for _, p := range shellProfilePaths(home) {
		content, err := os.ReadFile(p)
		if err != nil {
			continue
		}
		if count := strings.Count(string(content), "fnm env"); count > 1 {
			t.Fatalf("profile %s contains 'fnm env' %d times after two shell:setup runs; want at most 1", p, count)
		}
	}
}

func TestInstallAlsoConfiguresShellActivation(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("stub fnm tests target Unix-like systems")
	}
	t.Parallel()
	root := tasktestutil.ModuleRoot(t)
	env := fnmFreshProfileEnv(t)
	home := tasktestutil.EnvValue(env, "HOME")
	tasktestutil.AssertExitCode(t, tasktestutil.RunTask(t, root, env, "--yes", "install"), 0)
	if !profileContains(home, "fnm env") {
		t.Fatal("expected task install to configure shell activation but no profile contains 'fnm env'")
	}
}

func TestRealInstallerFlowOnlyWhenExplicitlyEnabled(t *testing.T) {
	if os.Getenv("RUN_INSTALLER_TESTS") != "1" {
		t.Skip("set RUN_INSTALLER_TESTS=1 to run real install/uninstall tests")
	}
	if runtime.GOOS == "windows" {
		t.Skip("real fnm shell installer tests are intended for Unix-like systems")
	}
	root := tasktestutil.ModuleRoot(t)
	env := tasktestutil.IsolatedEnv(t)
	home := tasktestutil.EnvValue(env, "HOME")
	fnmBin := filepath.Join(home, ".local", "bin", "fnm")

	tasktestutil.AssertExitCode(t, tasktestutil.RunTaskTimeout(t, root, env, 10*time.Minute, "--yes", "install"), 0)
	tasktestutil.AssertFileExists(t, fnmBin)
	tasktestutil.AssertExitCode(t, tasktestutil.RunTaskTimeout(t, root, env, 10*time.Minute, "version"), 0)
	tasktestutil.AssertExitCode(t, tasktestutil.RunTaskTimeout(t, root, env, 10*time.Minute, "--yes", "install:undo"), 0)
	if _, err := os.Stat(fnmBin); !os.IsNotExist(err) {
		t.Fatalf("expected fnm binary to be removed after install:undo: %s", fnmBin)
	}
}

func TestAllPublicTasksIntegration(t *testing.T) {
	if os.Getenv("RUN_INTEGRATION_TESTS") != "1" {
		t.Skip("set RUN_INTEGRATION_TESTS=1 to run integration tests (downloads and installs fnm and Node.js)")
	}
	if runtime.GOOS == "windows" {
		t.Skip("integration tests target Unix-like systems")
	}
	root := tasktestutil.ModuleRoot(t)
	env := tasktestutil.IsolatedEnv(t)
	home := tasktestutil.EnvValue(env, "HOME")
	fnmBin := filepath.Join(home, ".local", "bin", "fnm")
	fnmRoot := filepath.Join(home, ".local", "share", "fnm")

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

	step("install — fnm binary is present on disk", func(t *testing.T) {
		run(t, "--yes", "install")
		tasktestutil.AssertFileExists(t, fnmBin)
	})
	step("version — fnm version string is printed", func(t *testing.T) {
		result := run(t, "version")
		tasktestutil.AssertNotEmpty(t, result.Combined(), "version output is empty")
	})
	step("node:install — default LTS version directory is created", func(t *testing.T) {
		run(t, "--yes", "node:install")
		tasktestutil.AssertDirHasEntries(t, filepath.Join(fnmRoot, "node-versions"))
	})
	step("ls — installed versions appear in output", func(t *testing.T) {
		result := run(t, "ls")
		tasktestutil.AssertNotEmpty(t, result.Combined(), "ls output is empty")
	})
	const secondary = "18.0.0"
	step("node:install VERSION=18.0.0 — specific version directory is created", func(t *testing.T) {
		run(t, "--yes", "node:install", "VERSION="+secondary)
		tasktestutil.AssertDirExists(t, filepath.Join(fnmRoot, "node-versions", "v"+secondary))
	})
	step("node:uninstall VERSION=18.0.0 — specific version directory is removed", func(t *testing.T) {
		run(t, "--yes", "node:uninstall", "VERSION="+secondary)
		tasktestutil.AssertDirNotExists(t, filepath.Join(fnmRoot, "node-versions", "v"+secondary))
	})
	step("node:use — LTS is activated without error", func(t *testing.T) { run(t, "--yes", "node:use") })
	step("node:version — active node and npm version strings are printed", func(t *testing.T) {
		result := run(t, "node:version")
		tasktestutil.AssertContains(t, result.Combined(), "v")
	})
	step("install:undo — fnm binary is removed", func(t *testing.T) {
		run(t, "--yes", "install:undo")
		if _, err := os.Stat(fnmBin); !os.IsNotExist(err) {
			t.Fatalf("expected fnm binary to be removed: %s", fnmBin)
		}
	})
}

// fnmDryRunEnv returns an isolated environment with a stub fnm binary.
// .bashrc is pre-populated so the shell:setup status check exits 0 and
// shell:setup is skipped in tests that don't specifically test it.
func fnmDryRunEnv(t *testing.T) []string {
	t.Helper()

	env := tasktestutil.IsolatedEnv(t)
	home := tasktestutil.EnvValue(env, "HOME")

	binDir := filepath.Join(home, ".local", "bin")
	if err := os.MkdirAll(binDir, 0755); err != nil {
		t.Fatalf("failed to create stub bin dir: %v", err)
	}

	fnmRoot := filepath.Join(home, ".local", "share", "fnm")
	stub := "#!/usr/bin/env bash\n" +
		"case \"$1\" in\n" +
		"  --version) echo \"fnm 1.37.1 stub\" ;;\n" +
		"  root) echo \"" + fnmRoot + "\" ;;\n" +
		"  list|ls) echo \"* v20.0.0 default\" ;;\n" +
		"  current) echo \"v20.0.0\" ;;\n" +
		"  env) echo \"# fnm stub env\" ;;\n" +
		"  *) exit 0 ;;\n" +
		"esac\n"

	if err := os.WriteFile(filepath.Join(binDir, "fnm"), []byte(stub), 0755); err != nil {
		t.Fatalf("failed to create stub fnm binary: %v", err)
	}

	bashrc := filepath.Join(home, ".bashrc")
	bashrcContent := "# fnm (Fast Node Manager)\n" +
		"export PATH=\"$HOME/.local/share/fnm:$HOME/.local/bin:$PATH\"\n" +
		"eval \"$(fnm env --use-on-cd --shell bash)\"\n"
	if err := os.WriteFile(bashrc, []byte(bashrcContent), 0644); err != nil {
		t.Fatalf("failed to pre-populate shell profile: %v", err)
	}

	path := tasktestutil.EnvValue(env, "PATH")
	return tasktestutil.SetEnv(env, "PATH", binDir+":"+path)
}

// fnmFreshProfileEnv returns the same stub env as fnmDryRunEnv but with an
// empty .bashrc so shell:setup actually runs. Use only in shell:setup tests.
func fnmFreshProfileEnv(t *testing.T) []string {
	t.Helper()
	env := fnmDryRunEnv(t)
	home := tasktestutil.EnvValue(env, "HOME")
	if err := os.WriteFile(filepath.Join(home, ".bashrc"), []byte(""), 0644); err != nil {
		t.Fatalf("failed to clear shell profile: %v", err)
	}
	return env
}

func shellProfilePaths(home string) []string {
	return []string{
		filepath.Join(home, ".bashrc"),
		filepath.Join(home, ".bash_profile"),
		filepath.Join(home, ".profile"),
		filepath.Join(home, ".zshrc"),
		filepath.Join(home, ".zprofile"),
		filepath.Join(home, ".config", "fish", "config.fish"),
	}
}

func profileContains(home, token string) bool {
	for _, p := range shellProfilePaths(home) {
		content, err := os.ReadFile(p)
		if err != nil {
			continue
		}
		if strings.Contains(string(content), token) {
			return true
		}
	}
	return false
}
