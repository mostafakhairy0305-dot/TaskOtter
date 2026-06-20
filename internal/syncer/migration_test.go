package syncer_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/mostafakhairy0305-dot/taskotter-sync-action/internal/config"
	"github.com/mostafakhairy0305-dot/taskotter-sync-action/internal/normalizer"
	"github.com/mostafakhairy0305-dot/taskotter-sync-action/internal/syncer"
)

func TestTargetFolderMigration(t *testing.T) {
	workspace := t.TempDir()
	writeRootTaskfile(t, workspace)

	oldManaged := filepath.Join(workspace, "task/go/Taskfile.yml")
	if err := os.MkdirAll(filepath.Dir(oldManaged), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(oldManaged, []byte("version: '3'\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	oldUser := filepath.Join(workspace, "task/go/user.txt")
	if err := os.WriteFile(oldUser, []byte("keep"), 0o644); err != nil {
		t.Fatal(err)
	}
	writeMinimalLock(t, workspace, "task", []syncer.ManagedFile{
		{DestinationModule: "go", Path: "task/go/Taskfile.yml"},
	})

	cfg := &config.Config{
		Tasks:        []string{"go"},
		IncludesDoc:  true,
		TargetFolder: "taskfiles",
		Workspace:    workspace,
	}
	in, plan := preparePlan(t, workspace, cfg)
	if err := syncer.ApplyPlan(plan, in); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(workspace, "taskfiles/go/Taskfile.yml")); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(oldManaged); err == nil {
		t.Fatal("old managed file under previous target folder should be removed")
	}
	if _, err := os.Stat(oldUser); err != nil {
		t.Fatal("unknown files outside managed set should be preserved")
	}
}

func TestPrefixSafetyPreservesUnrelatedPaths(t *testing.T) {
	workspace := t.TempDir()
	writeRootTaskfile(t, workspace)

	extra := filepath.Join(workspace, "taskfiles-extra/foo.txt")
	if err := os.MkdirAll(filepath.Dir(extra), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(extra, []byte("stay"), 0o644); err != nil {
		t.Fatal(err)
	}
	writeMinimalLock(t, workspace, "task", []syncer.ManagedFile{
		{DestinationModule: "go", Path: "task/go/Taskfile.yml"},
	})

	cfg := &config.Config{
		Tasks:        []string{"go"},
		IncludesDoc:  true,
		TargetFolder: "taskfiles",
		Workspace:    workspace,
	}
	in, plan := preparePlan(t, workspace, cfg)
	if err := syncer.ApplyPlan(plan, in); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(extra); err != nil {
		t.Fatal("taskfiles-extra must not match old folder prefix task")
	}
}

func TestPackageManagerSwitchSameDestination(t *testing.T) {
	for _, mod := range []string{"eslint-pnpm-fnm", "eslint-npm-fnm"} {
		sourceToDest, err := normalizer.BuildDestinationMap([]string{mod})
		if err != nil {
			t.Fatal(err)
		}
		if sourceToDest[mod] != "eslint" {
			t.Fatalf("%s should normalize to eslint, got %q", mod, sourceToDest[mod])
		}
	}
}
