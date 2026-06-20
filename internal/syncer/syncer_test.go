package syncer_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/mostafakhairy0305-dot/taskotter-sync-action/internal/app"
	"github.com/mostafakhairy0305-dot/taskotter-sync-action/internal/config"
	"github.com/mostafakhairy0305-dot/taskotter-sync-action/internal/dependency"
	"github.com/mostafakhairy0305-dot/taskotter-sync-action/internal/normalizer"
	"github.com/mostafakhairy0305-dot/taskotter-sync-action/internal/resolver"
	"github.com/mostafakhairy0305-dot/taskotter-sync-action/internal/store"
	"github.com/mostafakhairy0305-dot/taskotter-sync-action/internal/syncer"
)

func fixtureStore(t *testing.T) *store.Snapshot {
	t.Helper()
	root := filepath.Join("..", "..", "tests", "fixtures", "store")
	snap, err := store.LocalSnapshot(root, store.RefInfo{
		Repository:     config.StoreRepository,
		SourceRef:      "refs/heads/main",
		ResolvedCommit: "abc123",
		DefaultBranch:  "main",
	})
	if err != nil {
		t.Fatal(err)
	}
	return snap
}

func writeRootTaskfile(t *testing.T, workspace string) {
	t.Helper()
	content := []byte(`version: "3"
includes: {}
tasks:
  hello:
    cmds:
      - echo hello
`)
	if err := os.WriteFile(filepath.Join(workspace, "Taskfile.yml"), content, 0o644); err != nil {
		t.Fatal(err)
	}
}

func dependencySources(t *testing.T, sources []string, snap *store.Snapshot) ([]string, error) {
	t.Helper()
	return dependency.ResolveTransitive(sources, snap.Deps)
}

func TestBuildPlanInitialSync(t *testing.T) {
	workspace := t.TempDir()
	writeRootTaskfile(t, workspace)
	snap := fixtureStore(t)

	cfg := &config.Config{
		Tasks:              []string{"eslint", "go"},
		NodePackageManager: config.PMPnpm,
		NodeVersionManager: config.VMFnm,
		IncludesDoc:        true,
		TargetFolder:       "taskfiles",
		Workspace:          workspace,
	}

	resolutions, err := resolver.ResolveAll(cfg.Tasks, snap.Catalog, cfg.NodePackageManager, cfg.NodeVersionManager)
	if err != nil {
		t.Fatal(err)
	}
	sources := make([]string, 0, len(resolutions))
	for _, res := range resolutions {
		sources = append(sources, res.SourceModule)
	}
	depSources, err := dependency.ResolveTransitive(sources, snap.Deps)
	if err != nil {
		t.Fatal(err)
	}
	in, err := app.PrepareSyncInput(cfg, snap, resolutions, depSources)
	if err != nil {
		t.Fatal(err)
	}
	plan, err := syncer.BuildPlan(in)
	if err != nil {
		t.Fatal(err)
	}
	if !plan.Changed {
		t.Fatal("expected changes on initial sync")
	}
}

func TestUnmanagedDestinationConflict(t *testing.T) {
	workspace := t.TempDir()
	writeRootTaskfile(t, workspace)
	if err := os.MkdirAll(filepath.Join(workspace, "taskfiles/eslint"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(workspace, "taskfiles/eslint/user.txt"), []byte("keep"), 0o644); err != nil {
		t.Fatal(err)
	}

	snap := fixtureStore(t)
	cfg := &config.Config{
		Tasks:              []string{"eslint"},
		NodePackageManager: config.PMPnpm,
		NodeVersionManager: config.VMFnm,
		IncludesDoc:        true,
		TargetFolder:       "taskfiles",
		Workspace:          workspace,
	}
	res, err := resolver.Resolve("eslint", snap.Catalog, cfg.NodePackageManager, cfg.NodeVersionManager)
	if err != nil {
		t.Fatal(err)
	}
	sourceToDest, err := normalizer.BuildDestinationMap([]string{res.SourceModule})
	if err != nil {
		t.Fatal(err)
	}
	_, err = syncer.BuildPlan(syncer.SyncInput{
		Config:       cfg,
		Snapshot:     snap,
		Requested:    map[string]syncer.ModuleRecord{"eslint": {SourceModule: res.SourceModule, DestinationModule: "eslint", Path: "taskfiles/eslint"}},
		SourceToDest: sourceToDest,
		DestByTask:   map[string]string{"eslint": "eslint"},
	})
	if err == nil {
		t.Fatal("expected unmanaged destination conflict")
	}
}

func TestIncludesDocFalseSkipsReadme(t *testing.T) {
	workspace := t.TempDir()
	writeRootTaskfile(t, workspace)
	snap := fixtureStore(t)
	cfg := &config.Config{
		Tasks:              []string{"eslint"},
		NodePackageManager: config.PMPnpm,
		NodeVersionManager: config.VMFnm,
		IncludesDoc:        false,
		TargetFolder:       "taskfiles",
		Workspace:          workspace,
	}
	res, err := resolver.Resolve("eslint", snap.Catalog, cfg.NodePackageManager, cfg.NodeVersionManager)
	if err != nil {
		t.Fatal(err)
	}
	sourceToDest := map[string]string{res.SourceModule: "eslint"}
	plan, err := syncer.BuildPlan(syncer.SyncInput{
		Config:       cfg,
		Snapshot:     snap,
		Requested:    map[string]syncer.ModuleRecord{"eslint": {SourceModule: res.SourceModule, DestinationModule: "eslint", Path: "taskfiles/eslint"}},
		SourceToDest: sourceToDest,
		DestByTask:   map[string]string{"eslint": "eslint"},
	})
	if err != nil {
		t.Fatal(err)
	}
	for _, mf := range plan.ManagedFiles {
		if filepath.Base(mf.Path) == "README.md" {
			t.Fatal("README should be excluded when includes-doc=false")
		}
	}
}
