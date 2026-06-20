package syncer_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/mostafakhairy0305-dot/TaskOtter/internal/app"
	"github.com/mostafakhairy0305-dot/TaskOtter/internal/config"
	"github.com/mostafakhairy0305-dot/TaskOtter/internal/resolver"
	"github.com/mostafakhairy0305-dot/TaskOtter/internal/syncer"
	"gopkg.in/yaml.v3"
)

func TestBuildPlanCorruptLockFails(t *testing.T) {
	workspace := t.TempDir()
	writeRootTaskfile(t, workspace)
	if err := os.MkdirAll(filepath.Join(workspace, "taskfiles"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(workspace, "taskfiles/.taskotter-lock.yml"), []byte("{{not yaml"), 0o644); err != nil {
		t.Fatal(err)
	}
	meta := []byte(`target_folder: taskfiles
lock_file: taskfiles/.taskotter-lock.yml
configuration_hash: abc
`)
	if err := os.MkdirAll(filepath.Join(workspace, ".taskotter"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(workspace, ".taskotter/metadata.yml"), meta, 0o644); err != nil {
		t.Fatal(err)
	}

	cfg := &config.Config{
		Tasks:        []string{"go"},
		IncludesDoc:  true,
		TargetFolder: "taskfiles",
		Workspace:    workspace,
	}
	snap := fixtureStore(t)
	res, err := resolver.Resolve("go", snap.Catalog, "", "")
	if err != nil {
		t.Fatal(err)
	}
	in := syncer.SyncInput{
		Config:       cfg,
		Snapshot:     snap,
		Requested:    map[string]syncer.ModuleRecord{"go": {SourceModule: res.SourceModule, DestinationModule: "go", Path: "taskfiles/go"}},
		SourceToDest: map[string]string{res.SourceModule: "go"},
		DestByTask:   map[string]string{"go": "go"},
	}
	if _, err := syncer.BuildPlan(in); err == nil {
		t.Fatal("expected corrupt lock error")
	}
}

func TestBuildPlanCorruptMetadataFails(t *testing.T) {
	workspace := t.TempDir()
	writeRootTaskfile(t, workspace)
	if err := os.MkdirAll(filepath.Join(workspace, ".taskotter"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(workspace, ".taskotter/metadata.yml"), []byte("{{not yaml"), 0o644); err != nil {
		t.Fatal(err)
	}

	cfg := &config.Config{
		Tasks:        []string{"go"},
		IncludesDoc:  true,
		TargetFolder: "taskfiles",
		Workspace:    workspace,
	}
	snap := fixtureStore(t)
	res, err := resolver.Resolve("go", snap.Catalog, "", "")
	if err != nil {
		t.Fatal(err)
	}
	in := syncer.SyncInput{
		Config:       cfg,
		Snapshot:     snap,
		Requested:    map[string]syncer.ModuleRecord{"go": {SourceModule: res.SourceModule, DestinationModule: "go", Path: "taskfiles/go"}},
		SourceToDest: map[string]string{res.SourceModule: "go"},
		DestByTask:   map[string]string{"go": "go"},
	}
	if _, err := syncer.BuildPlan(in); err == nil {
		t.Fatal("expected corrupt metadata error")
	}
}

func TestMetadataOnlyChangeMarksChanged(t *testing.T) {
	workspace := t.TempDir()
	writeRootTaskfile(t, workspace)
	cfg := &config.Config{
		Tasks:             []string{"go"},
		IncludesDoc:       true,
		TargetFolder:      "taskfiles",
		Workspace:         workspace,
		ConfigurationHash: "hash-a",
	}
	in, plan := preparePlan(t, workspace, cfg)
	if err := syncer.ApplyPlan(plan, in); err != nil {
		t.Fatal(err)
	}

	cfg.ConfigurationHash = "hash-b"
	_, plan2 := preparePlan(t, workspace, cfg)
	if !plan2.Changed {
		t.Fatal("expected metadata-only configuration hash change to mark plan changed")
	}
}

func TestSHAOnlyLockChangeNotChanged(t *testing.T) {
	workspace := t.TempDir()
	writeRootTaskfile(t, workspace)
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

	snap := fixtureStore(t)
	snap.Ref.ResolvedCommit = "different-sha-only"
	resolutions, err := resolver.ResolveAll(cfg.Tasks, snap.Catalog, cfg.NodePackageManager, cfg.NodeVersionManager)
	if err != nil {
		t.Fatal(err)
	}
	in2, err := app.PrepareSyncInput(cfg, snap, resolutions, nil)
	if err != nil {
		t.Fatal(err)
	}
	plan2, err := syncer.BuildPlan(in2)
	if err != nil {
		t.Fatal(err)
	}
	if plan2.Changed {
		t.Fatalf("expected no file changes when only resolved commit differs: added=%v updated=%v removed=%v", plan2.Added, plan2.Updated, plan2.Removed)
	}
}

func TestConfigurationChangeMarksUpdated(t *testing.T) {
	workspace := t.TempDir()
	writeRootTaskfile(t, workspace)
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

	cfg.IncludesDoc = false
	_, plan2 := preparePlan(t, workspace, cfg)
	if !plan2.Changed {
		t.Fatal("expected changes when includes-doc toggles")
	}
}

func writeMinimalLock(t *testing.T, workspace, targetFolder string, files []syncer.ManagedFile) {
	t.Helper()
	lock := syncer.LockFile{}
	lock.Configuration.TargetFolder = targetFolder
	lock.ManagedFiles = files
	data, err := yaml.Marshal(lock)
	if err != nil {
		t.Fatal(err)
	}
	lockPath := filepath.Join(workspace, targetFolder, ".taskotter-lock.yml")
	if err := os.MkdirAll(filepath.Dir(lockPath), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(lockPath, data, 0o644); err != nil {
		t.Fatal(err)
	}
	meta := []byte("target_folder: " + targetFolder + "\nlock_file: " + targetFolder + "/.taskotter-lock.yml\nconfiguration_hash: x\n")
	if err := os.MkdirAll(filepath.Join(workspace, ".taskotter"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(workspace, ".taskotter/metadata.yml"), meta, 0o644); err != nil {
		t.Fatal(err)
	}
}
