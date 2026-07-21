package syncer_test

import (
	"os"
	"path/filepath"
	"slices"
	"testing"

	"github.com/mostafakhairy0305-dot/TaskOtter/internal/app"
	"github.com/mostafakhairy0305-dot/TaskOtter/internal/config"
	"github.com/mostafakhairy0305-dot/TaskOtter/internal/dependency"
	"github.com/mostafakhairy0305-dot/TaskOtter/internal/normalizer"
	"github.com/mostafakhairy0305-dot/TaskOtter/internal/resolver"
	"github.com/mostafakhairy0305-dot/TaskOtter/internal/store"
	"github.com/mostafakhairy0305-dot/TaskOtter/internal/syncer"
)

func fixtureStore(t *testing.T) *store.Snapshot {
	t.Helper()

	root := filepath.Join("..", "..", "tests", "fixtures", "store")

	snap, err := store.LocalSnapshot(root, store.RefInfo{
		Repository:       config.StoreRepository,
		RequestedVersion: "",
		SourceRef:        "refs/heads/main",
		ResolvedCommit:   "abc123",
		DefaultBranch:    "main",
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

	err := os.WriteFile(filepath.Join(workspace, testTaskfileName), content, 0o644)
	if err != nil {
		t.Fatal(err)
	}
}

func dependencySources(t *testing.T, sources []string, snap *store.Snapshot) ([]string, error) {
	t.Helper()

	return dependency.ResolveTransitive(sources, snap.Deps)
}

func TestBuildPlanInitialSync(t *testing.T) {
	t.Parallel()

	workspace := t.TempDir()
	writeRootTaskfile(t, workspace)
	snap := fixtureStore(t)

	cfg := testConfig(workspace, func(cfg *config.Config) {
		cfg.Tasks = []string{testModuleEslint, "go"}
		cfg.NodePackageManager = config.PMPnpm
		cfg.NodeVersionManager = config.VMFnm
		cfg.IncludesDoc = true
	})

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

	syncInput, err := app.PrepareSyncInput(cfg, snap, resolutions, depSources)
	if err != nil {
		t.Fatal(err)
	}

	plan, err := syncer.BuildPlan(syncInput)
	if err != nil {
		t.Fatal(err)
	}

	if !plan.Changed {
		t.Fatal("expected changes on initial sync")
	}
}

func TestBuildPlanCreatesRootTaskfile(t *testing.T) {
	t.Parallel()

	workspace := t.TempDir()
	snap := fixtureStore(t)

	cfg := testConfig(workspace, func(cfg *config.Config) {
		cfg.Tasks = []string{"go"}
		cfg.IncludesDoc = true
	})

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

	syncInput, err := app.PrepareSyncInput(cfg, snap, resolutions, depSources)
	if err != nil {
		t.Fatal(err)
	}

	plan, err := syncer.BuildPlan(syncInput)
	if err != nil {
		t.Fatal(err)
	}

	if !plan.Changed {
		t.Fatal("expected changes on initial sync")
	}

	if !containsRootTaskfile(plan.Added) {
		t.Fatalf("expected root Taskfile.yml in added files, got added=%v", plan.Added)
	}
}

func containsRootTaskfile(list []string) bool {
	return slices.Contains(list, testTaskfileName)
}

func TestUnmanagedDestinationConflict(t *testing.T) {
	t.Parallel()

	workspace := t.TempDir()
	writeRootTaskfile(t, workspace)

	err := os.MkdirAll(filepath.Join(workspace, config.DefaultTargetFolder, testModuleEslint), 0o755)
	if err != nil {
		t.Fatal(err)
	}

	err = os.WriteFile(
		filepath.Join(workspace, config.DefaultTargetFolder, testModuleEslint, "user.txt"),
		[]byte("keep"),
		0o644,
	)
	if err != nil {
		t.Fatal(err)
	}

	snap := fixtureStore(t)
	cfg := testConfig(workspace, func(cfg *config.Config) {
		cfg.Tasks = []string{testModuleEslint}
		cfg.NodePackageManager = config.PMPnpm
		cfg.NodeVersionManager = config.VMFnm
		cfg.IncludesDoc = true
	})

	res, err := resolver.Resolve(testModuleEslint, snap.Catalog, cfg.NodePackageManager, cfg.NodeVersionManager)
	if err != nil {
		t.Fatal(err)
	}

	sourceToDest, err := normalizer.BuildDestinationMap([]string{res.SourceModule})
	if err != nil {
		t.Fatal(err)
	}

	eslintPath := config.DefaultTargetFolder + "/" + testModuleEslint

	_, err = syncer.BuildPlan(syncer.SyncInput{
		Config:   cfg,
		Snapshot: snap,
		Requested: map[string]syncer.ModuleRecord{
			testModuleEslint: {
				SourceModule:      res.SourceModule,
				DestinationModule: testModuleEslint,
				Path:              eslintPath,
			},
		},
		Dependencies: nil,
		SourceToDest: sourceToDest,
		DestByTask:   map[string]string{testModuleEslint: testModuleEslint},
	})
	if err == nil {
		t.Fatal("expected unmanaged destination conflict")
	}
}

func TestIncludesDocFalseSkipsReadme(t *testing.T) {
	t.Parallel()

	workspace := t.TempDir()
	writeRootTaskfile(t, workspace)
	snap := fixtureStore(t)
	cfg := testConfig(workspace, func(cfg *config.Config) {
		cfg.Tasks = []string{testModuleEslint}
		cfg.NodePackageManager = config.PMPnpm
		cfg.NodeVersionManager = config.VMFnm
		cfg.IncludesDoc = false
	})

	res, err := resolver.Resolve(testModuleEslint, snap.Catalog, cfg.NodePackageManager, cfg.NodeVersionManager)
	if err != nil {
		t.Fatal(err)
	}

	sourceToDest := map[string]string{res.SourceModule: testModuleEslint}

	eslintPath := config.DefaultTargetFolder + "/" + testModuleEslint

	plan, err := syncer.BuildPlan(syncer.SyncInput{
		Config:   cfg,
		Snapshot: snap,
		Requested: map[string]syncer.ModuleRecord{
			testModuleEslint: {
				SourceModule:      res.SourceModule,
				DestinationModule: testModuleEslint,
				Path:              eslintPath,
			},
		},
		Dependencies: nil,
		SourceToDest: sourceToDest,
		DestByTask:   map[string]string{testModuleEslint: testModuleEslint},
	})
	if err != nil {
		t.Fatal(err)
	}

	for _, managed := range plan.ManagedFiles {
		if filepath.Base(managed.Path) == testReadmeName {
			t.Fatal("README should be excluded when includes-doc=false")
		}
	}
}
