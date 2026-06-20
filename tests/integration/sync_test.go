package integration_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/mostafakhairy0305-dot/taskotter-sync-action/internal/app"
	"github.com/mostafakhairy0305-dot/taskotter-sync-action/internal/config"
	"github.com/mostafakhairy0305-dot/taskotter-sync-action/internal/store"
)

type localStore struct {
	root string
}

func (l *localStore) ResolveRef(_ context.Context, requestedVersion string) (store.RefInfo, error) {
	return store.RefInfo{
		Repository:       config.StoreRepository,
		RequestedVersion: requestedVersion,
		SourceRef:        "refs/heads/main",
		ResolvedCommit:   "deadbeef",
		DefaultBranch:    "main",
	}, nil
}

func (l *localStore) DownloadSnapshot(_ context.Context, ref store.RefInfo) (*store.Snapshot, error) {
	return store.LocalSnapshot(l.root, ref)
}

func TestIntegrationSyncNoGit(t *testing.T) {
	workspace := t.TempDir()
	rootTaskfile := []byte(`version: "3"
includes: {}
`)
	if err := os.WriteFile(filepath.Join(workspace, "Taskfile.yml"), rootTaskfile, 0o644); err != nil {
		t.Fatal(err)
	}

	fixtureRoot, err := filepath.Abs(filepath.Join("..", "fixtures", "store"))
	if err != nil {
		t.Fatal(err)
	}

	cfg := &config.Config{
		Tasks:        []string{"go"},
		IncludesDoc:  true,
		TargetFolder: "taskfiles",
		Workspace:    workspace,
	}

	o := &app.Orchestrator{
		StoreClient: &localStore{root: fixtureRoot},
	}

	result, err := o.Run(context.Background(), cfg)
	if err != nil {
		t.Fatal(err)
	}
	if !result.Changed {
		t.Fatal("expected changes")
	}
	if _, err := os.Stat(filepath.Join(workspace, "taskfiles/go/Taskfile.yml")); err != nil {
		t.Fatal(err)
	}
}
