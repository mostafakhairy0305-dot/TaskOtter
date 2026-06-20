package syncer_test

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/mostafakhairy0305-dot/TaskOtter/internal/app"
	"github.com/mostafakhairy0305-dot/TaskOtter/internal/config"
	"github.com/mostafakhairy0305-dot/TaskOtter/internal/resolver"
	"github.com/mostafakhairy0305-dot/TaskOtter/internal/syncer"
)

func preparePlan(t *testing.T, workspace string, cfg *config.Config) (syncer.SyncInput, *syncer.Plan) {
	t.Helper()
	snap := fixtureStore(t)
	resolutions, err := resolver.ResolveAll(cfg.Tasks, snap.Catalog, cfg.NodePackageManager, cfg.NodeVersionManager)
	if err != nil {
		t.Fatal(err)
	}
	sources := make([]string, 0, len(resolutions))
	for _, res := range resolutions {
		sources = append(sources, res.SourceModule)
	}
	depSources, err := dependencySources(t, sources, snap)
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
	return in, plan
}

func writeFileEntry(path string, entry syncer.FileEntry) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	return os.WriteFile(path, entry.Data, entry.Mode)
}

func TestApplyPlanWritesFiles(t *testing.T) {
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
	if _, err := os.Stat(filepath.Join(workspace, "taskfiles/go/Taskfile.yml")); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(workspace, ".taskotter/metadata.yml")); err != nil {
		t.Fatal(err)
	}
}

func TestApplyPlanPreservesExecutableMode(t *testing.T) {
	workspace := t.TempDir()
	writeRootTaskfile(t, workspace)
	setupPath := filepath.Join("..", "..", "tests", "fixtures", "store", "taskfiles", "go", "setup.sh")
	if err := os.Chmod(setupPath, 0o755); err != nil {
		t.Fatal(err)
	}

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
	info, err := os.Stat(filepath.Join(workspace, "taskfiles/go/setup.sh"))
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode().Perm()&0o111 == 0 {
		t.Fatalf("expected executable bit, got %o", info.Mode().Perm())
	}
}

func TestApplyPlanPromoteBeforeDelete(t *testing.T) {
	workspace := t.TempDir()
	writeRootTaskfile(t, workspace)

	obsolete := filepath.Join(workspace, "taskfiles/go/obsolete.txt")
	if err := os.MkdirAll(filepath.Dir(obsolete), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(obsolete, []byte("old"), 0o644); err != nil {
		t.Fatal(err)
	}
	writeMinimalLock(t, workspace, "taskfiles", []syncer.ManagedFile{
		{DestinationModule: "go", Path: "taskfiles/go/obsolete.txt"},
	})

	cfg := &config.Config{
		Tasks:        []string{"go"},
		IncludesDoc:  true,
		TargetFolder: "taskfiles",
		Workspace:    workspace,
	}
	in, plan := preparePlan(t, workspace, cfg)

	var promoted int
	syncer.CopyFileToHook = func(path string, entry syncer.FileEntry) error {
		if strings.Contains(path, filepath.Join(".taskotter", "staging")) {
			return writeFileEntry(path, entry)
		}
		promoted++
		if promoted == 1 {
			return fmt.Errorf("simulated promote failure")
		}
		return writeFileEntry(path, entry)
	}
	t.Cleanup(func() { syncer.CopyFileToHook = nil })

	if err := syncer.ApplyPlan(plan, in); err == nil {
		t.Fatal("expected promote failure")
	}
	if _, err := os.Stat(obsolete); err != nil {
		t.Fatal("obsolete file should remain when promote fails before delete")
	}
}

func TestApplyPlanWriteOrder(t *testing.T) {
	workspace := t.TempDir()
	writeRootTaskfile(t, workspace)
	cfg := &config.Config{
		Tasks:        []string{"go"},
		IncludesDoc:  true,
		TargetFolder: "taskfiles",
		Workspace:    workspace,
	}
	in, plan := preparePlan(t, workspace, cfg)

	var order []string
	stagingMarker := filepath.Join(".taskotter", "staging")
	syncer.CopyFileToHook = func(path string, entry syncer.FileEntry) error {
		if strings.Contains(path, stagingMarker) {
			return writeFileEntry(path, entry)
		}
		rel, err := filepath.Rel(workspace, path)
		if err != nil {
			return err
		}
		order = append(order, filepath.ToSlash(rel))
		return writeFileEntry(path, entry)
	}
	t.Cleanup(func() { syncer.CopyFileToHook = nil })

	if err := syncer.ApplyPlan(plan, in); err != nil {
		t.Fatal(err)
	}
	lockIdx := indexOfSuffix(order, ".taskotter-lock.yml")
	metaIdx := indexOfSuffix(order, "metadata.yml")
	moduleIdx := indexOfContains(order, "taskfiles/go/")
	if lockIdx < moduleIdx || metaIdx < lockIdx {
		t.Fatalf("expected modules before lock before metadata, got %v", order)
	}
}

func indexOfSuffix(paths []string, suffix string) int {
	for i, p := range paths {
		if strings.HasSuffix(p, suffix) {
			return i
		}
	}
	return -1
}

func indexOfContains(paths []string, part string) int {
	for i, p := range paths {
		if strings.Contains(p, part) {
			return i
		}
	}
	return -1
}
