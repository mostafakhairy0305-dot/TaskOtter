package taskfile_test

import (
	"os"
	"strings"
	"testing"

	"github.com/mostafakhairy0305-dot/TaskOtter/internal/taskfile"
)

func TestRewriteIncludes(t *testing.T) {
	input := []byte(`version: "3"
includes:
  pnpm:
    taskfile: ../pnpm-fnm/Taskfile.yml
tasks:
  lint:
    cmds:
      - echo ../pnpm-fnm/Taskfile.yml
`)
	sourceToDest := map[string]string{
		"pnpm-fnm": "pnpm",
	}
	out, err := taskfile.RewriteIncludes(input, sourceToDest)
	if err != nil {
		t.Fatal(err)
	}
	text := string(out)
	if !strings.Contains(text, "../pnpm/Taskfile.yml") {
		t.Fatalf("include not rewritten: %s", text)
	}
	if !strings.Contains(text, "../pnpm-fnm/Taskfile.yml") {
		t.Fatalf("command string should remain unchanged: %s", text)
	}
}

func TestUpdateRootTaskfileFromTemplate(t *testing.T) {
	out, err := taskfile.UpdateRootTaskfile(taskfile.NewRootTemplate(), taskfile.RootUpdateInput{
		Tasks:        []string{"go"},
		TargetFolder: "taskfiles",
		DestByTask:   map[string]string{"go": "go"},
		ModuleTaskfiles: map[string][]byte{
			"go": []byte(`version: "3"
vars:
  GO_VERSION: ""
  GO_CMD_UNIX: /usr/local/go/bin/go
`),
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	text := string(out)
	if !strings.Contains(text, "taskfiles/go/Taskfile.yml") {
		t.Fatalf("missing go include: %s", text)
	}
	if !strings.Contains(text, "GO_VERSION") {
		t.Fatalf("missing include vars: %s", text)
	}
	if !strings.Contains(text, "GO_CMD_UNIX") {
		t.Fatalf("missing include vars: %s", text)
	}
}

func TestUpdateRootTaskfilePreservesExistingIncludeVars(t *testing.T) {
	root := []byte(`version: "3"
includes:
  go:
    taskfile: taskfiles/go/Taskfile.yml
    vars:
      GO_VERSION: go1.22.0
`)
	module := []byte(`version: "3"
vars:
  GO_VERSION: ""
  GO_CMD_UNIX: /usr/local/go/bin/go
`)
	out, err := taskfile.UpdateRootTaskfile(root, taskfile.RootUpdateInput{
		Tasks:        []string{"go"},
		TargetFolder: "taskfiles",
		DestByTask:   map[string]string{"go": "go"},
		ManagedTasks: []string{"go"},
		ModuleTaskfiles: map[string][]byte{
			"go": module,
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	text := string(out)
	if !strings.Contains(text, "go1.22.0") {
		t.Fatalf("existing include var override removed: %s", text)
	}
	if !strings.Contains(text, "GO_CMD_UNIX") {
		t.Fatalf("missing newly added module var: %s", text)
	}
}

func TestUpdateRootTaskfile(t *testing.T) {
	root := []byte(`version: "3"
includes:
  custom:
    taskfile: custom/Taskfile.yml
tasks:
  hello:
    cmds:
      - echo hi
`)
	out, err := taskfile.UpdateRootTaskfile(root, taskfile.RootUpdateInput{
		Tasks:        []string{"go", "eslint"},
		TargetFolder: "taskfiles",
		DestByTask: map[string]string{
			"go":     "go",
			"eslint": "eslint",
		},
		ManagedTasks: []string{},
	})
	if err != nil {
		t.Fatal(err)
	}
	text := string(out)
	if !strings.Contains(text, "taskfiles/go/Taskfile.yml") {
		t.Fatalf("missing go include: %s", text)
	}
	if !strings.Contains(text, "taskfiles/eslint/Taskfile.yml") {
		t.Fatalf("missing eslint include: %s", text)
	}
	if !strings.Contains(text, "custom/Taskfile.yml") {
		t.Fatalf("user include removed: %s", text)
	}
}

func TestManagedIncludeDifferentPathConflict(t *testing.T) {
	root := []byte(`version: "3"
includes:
  eslint:
    taskfile: custom/eslint/Taskfile.yml
tasks:
  hello:
    cmds:
      - echo hi
`)
	_, err := taskfile.UpdateRootTaskfile(root, taskfile.RootUpdateInput{
		Tasks:        []string{"eslint"},
		TargetFolder: "taskfiles",
		DestByTask:   map[string]string{"eslint": "eslint"},
		ManagedTasks: []string{"eslint"},
	})
	if err == nil {
		t.Fatal("expected conflict when alias path differs from managed path")
	}
}

func TestRootTaskfileAliasConflict(t *testing.T) {
	root := []byte(`version: "3"
includes:
  go:
    taskfile: legacy/go/Taskfile.yml
`)
	_, err := taskfile.UpdateRootTaskfile(root, taskfile.RootUpdateInput{
		Tasks:        []string{"go"},
		TargetFolder: "taskfiles",
		DestByTask:   map[string]string{"go": "go"},
		ManagedTasks: []string{},
	})
	if err == nil {
		t.Fatal("expected alias conflict")
	}
}

func TestScalarIncludeWrongPathConflict(t *testing.T) {
	root := []byte(`version: "3"
includes:
  go: legacy/go/Taskfile.yml
`)
	_, err := taskfile.UpdateRootTaskfile(root, taskfile.RootUpdateInput{
		Tasks:        []string{"go"},
		TargetFolder: "taskfiles",
		DestByTask:   map[string]string{"go": "go"},
		ManagedTasks: []string{},
	})
	if err == nil {
		t.Fatal("expected scalar include path conflict")
	}
}

func TestRewriteUsesRealStoreSnippet(t *testing.T) {
	data, err := os.ReadFile("../../tests/fixtures/store/taskfiles/eslint-pnpm-fnm/Taskfile.yml")
	if err != nil {
		t.Fatal(err)
	}
	out, err := taskfile.RewriteIncludes(data, map[string]string{
		"pnpm-fnm": "pnpm",
	})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(out), "../pnpm/Taskfile.yml") {
		t.Fatalf("rewrite failed: %s", out)
	}
}
