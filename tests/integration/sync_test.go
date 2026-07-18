package integration_test

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/mostafakhairy0305-dot/TaskOtter/internal/app"
	"github.com/mostafakhairy0305-dot/TaskOtter/internal/config"
	"github.com/mostafakhairy0305-dot/TaskOtter/internal/logging"
	"github.com/mostafakhairy0305-dot/TaskOtter/internal/store"
)

type localStore struct {
	root string
}

func (localStore *localStore) ResolveRef(_ context.Context, requestedVersion string) (store.RefInfo, error) {
	return store.RefInfo{
		Repository:       config.StoreRepository,
		RequestedVersion: requestedVersion,
		SourceRef:        "refs/heads/main",
		ResolvedCommit:   "deadbeef",
		DefaultBranch:    "main",
	}, nil
}

func (localStore *localStore) DownloadSnapshot(_ context.Context, ref store.RefInfo) (*store.Snapshot, error) {
	return store.LocalSnapshot(localStore.root, ref)
}

func TestIntegrationSyncNoGit(t *testing.T) {
	t.Parallel()

	workspace := t.TempDir()

	rootTaskfile := []byte(`version: "3"
includes: {}
`)

	err := os.WriteFile(filepath.Join(workspace, "Taskfile.yml"), rootTaskfile, 0o644)
	if err != nil {
		t.Fatal(err)
	}

	fixtureRoot, err := filepath.Abs(filepath.Join("..", "fixtures", "store"))
	if err != nil {
		t.Fatal(err)
	}

	cfg := &config.Config{
		Tasks:              []string{"go"},
		JSRuntime:          "",
		NodePackageManager: "",
		NodeVersionManager: "",
		IncludesDoc:        true,
		SyncRoot:           true,
		FailOnChanges:      false,
		StoreVersion:       "",
		TargetFolder:       "taskfiles",
		RootTaskfile:       "taskfiles/Taskfile.yml",
		GitHubToken:        "",
		Workspace:          workspace,
		Repository:         "",
		GitHubOutput:       "",
		BaseBranch:         "",
		ConfigurationHash:  "",
		BranchName:         "",
	}

	orchestrator := &app.Orchestrator{
		Logger:      logging.New(),
		StoreClient: &localStore{root: fixtureRoot},
		GitOps:      nil,
		PRClient:    nil,
	}

	result, err := orchestrator.Run(context.Background(), cfg)
	if err != nil {
		t.Fatal(err)
	}

	if !result.Changed {
		t.Fatal("expected changes")
	}

	_, err = os.Stat(filepath.Join(workspace, "taskfiles/go/Taskfile.yml"))
	if err != nil {
		t.Fatal(err)
	}

	// The aggregator Taskfile lands inside the target folder (default
	// <target-folder>/Taskfile.yml) and references modules with folder-relative
	// include paths rather than repeating the target folder prefix.
	rootData, err := os.ReadFile(filepath.Join(workspace, "taskfiles/Taskfile.yml"))
	if err != nil {
		t.Fatalf("aggregator Taskfile not written to target folder: %v", err)
	}

	if !strings.Contains(string(rootData), "taskfile: go/Taskfile.yml") {
		t.Fatalf("expected folder-relative include, got:\n%s", rootData)
	}

	if strings.Contains(string(rootData), "taskfiles/go/Taskfile.yml") {
		t.Fatalf("include should not repeat the target folder, got:\n%s", rootData)
	}
}
