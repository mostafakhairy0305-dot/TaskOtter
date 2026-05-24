package taskotter_test

import (
	"context"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"
	"testing"
	"time"

	"gopkg.in/yaml.v3"
)

func TestEveryTaskfileDirectoryHasReadmeAndTest(t *testing.T) {
	err := filepath.WalkDir("taskfiles", func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() || entry.Name() != "Taskfile.yml" {
			return nil
		}

		dir := filepath.Dir(path)
		if !fileExists(filepath.Join(dir, "README.md")) {
			t.Fatalf("%s must include README.md", dir)
		}

		tests, err := filepath.Glob(filepath.Join(dir, "*_test.go"))
		if err != nil {
			return err
		}
		if len(tests) == 0 {
			t.Fatalf("%s must include a *_test.go file", dir)
		}

		return nil
	})
	if err != nil {
		t.Fatalf("walk taskfiles: %v", err)
	}
}

func TestRootIncludesEveryTaskfileModule(t *testing.T) {
	content, err := os.ReadFile("Taskfile.yml")
	if err != nil {
		t.Fatalf("read root Taskfile: %v", err)
	}

	var root struct {
		Includes map[string]any `yaml:"includes"`
	}
	if err := yaml.Unmarshal(content, &root); err != nil {
		t.Fatalf("parse root Taskfile: %v", err)
	}

	var expected []string
	err = filepath.WalkDir("taskfiles", func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() || entry.Name() != "Taskfile.yml" {
			return nil
		}
		expected = append(expected, filepath.Base(filepath.Dir(path)))
		return nil
	})
	if err != nil {
		t.Fatalf("walk taskfiles: %v", err)
	}

	var actual []string
	for name := range root.Includes {
		actual = append(actual, name)
	}

	slices.Sort(expected)
	slices.Sort(actual)
	if !slices.Equal(expected, actual) {
		t.Fatalf("root Taskfile include drift\nexpected: %v\nactual:   %v", expected, actual)
	}
}

func TestRootToolDefaultsStayNamespaced(t *testing.T) {
	output := rootDryRun(t, "prettier:check")

	for _, token := range []string{
		`BINARY="prettier"`,
		". --check",
		".prettierignore",
	} {
		if !strings.Contains(output, token) {
			t.Fatalf("root prettier dry-run missing %q\noutput:\n%s", token, output)
		}
	}

	if strings.Contains(output, "**/*.{css,scss,sass,less,vue,svelte,astro} --check") {
		t.Fatalf("prettier picked up stylelint targets from another include\noutput:\n%s", output)
	}
}

func TestRootAggregatesForwardCommonOverrides(t *testing.T) {
	output := rootDryRun(t, "check", "PM=pnpm", "TARGETS=src")

	for _, token := range []string{
		`pnpm:exec BINARY="eslint"`,
		`pnpm:exec BINARY="prettier"`,
		`pnpm:exec BINARY="biome"`,
		`pnpm:exec BINARY="stylelint"`,
		`pnpm:exec BINARY="knip"`,
		`pnpm:exec BINARY="depcheck"`,
	} {
		if !strings.Contains(output, token) {
			t.Fatalf("root aggregate dry-run missing %q\noutput:\n%s", token, output)
		}
	}
}

func rootDryRun(t *testing.T, args ...string) string {
	t.Helper()

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	allArgs := append([]string{"--dry", "--yes", "--verbose"}, args...)
	cmd := exec.CommandContext(ctx, "task", allArgs...)
	cmd.Env = os.Environ()

	output, err := cmd.CombinedOutput()
	if ctx.Err() == context.DeadlineExceeded {
		t.Fatalf("task command timed out: task %s", strings.Join(allArgs, " "))
	}
	if err != nil {
		t.Fatalf("task command failed: task %s\nerror: %v\noutput:\n%s", strings.Join(allArgs, " "), err, string(output))
	}

	return string(output)
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}
