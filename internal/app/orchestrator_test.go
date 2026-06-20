package app_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/mostafakhairy0305-dot/taskotter-sync-action/internal/app"
	"github.com/mostafakhairy0305-dot/taskotter-sync-action/internal/config"
	"github.com/mostafakhairy0305-dot/taskotter-sync-action/internal/git"
	gh "github.com/mostafakhairy0305-dot/taskotter-sync-action/internal/github"
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

type mockGitOps struct {
	unrelated          bool
	defaultBranch      string
	defaultBranchCalls int
}

func (m *mockGitOps) EnsureSafeDirectory(context.Context) error { return nil }
func (m *mockGitOps) HasUnrelatedChanges(context.Context, map[string]struct{}) (bool, error) {
	return m.unrelated, nil
}
func (m *mockGitOps) CheckoutBranch(context.Context, string, bool) error { return nil }
func (m *mockGitOps) BranchExists(context.Context, string) (bool, error) {
	return false, nil
}
func (m *mockGitOps) LastCommitMessage(context.Context, string) (string, error) {
	return "", nil
}
func (m *mockGitOps) Stage(context.Context, []string) error    { return nil }
func (m *mockGitOps) Commit(context.Context, string) error     { return nil }
func (m *mockGitOps) Push(context.Context, string, bool) error { return nil }
func (m *mockGitOps) DefaultBranch(context.Context) (string, error) {
	m.defaultBranchCalls++
	if m.defaultBranch != "" {
		return m.defaultBranch, nil
	}
	return "main", nil
}

type mockPR struct {
	find        *gh.PullRequest
	create      *gh.PullRequest
	updated     int
	lastBase    string
	lastHead    string
	createdBase string
}

func (m *mockPR) FindOpenPR(_ context.Context, branch, base string) (*gh.PullRequest, error) {
	m.lastHead = branch
	m.lastBase = base
	return m.find, nil
}
func (m *mockPR) CreatePR(_ context.Context, branch, base, _ string) (*gh.PullRequest, error) {
	m.createdBase = base
	if m.create != nil {
		return m.create, nil
	}
	return &gh.PullRequest{Number: 99, URL: "https://example/pr/99"}, nil
}
func (m *mockPR) UpdatePRBody(context.Context, int, string) error {
	m.updated++
	return nil
}

func fixtureRoot(t *testing.T) string {
	t.Helper()
	root, err := filepath.Abs(filepath.Join("..", "..", "tests", "fixtures", "store"))
	if err != nil {
		t.Fatal(err)
	}
	return root
}

func workspaceWithRootTaskfile(t *testing.T) string {
	t.Helper()
	workspace := t.TempDir()
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
	return workspace
}

func TestOrchestratorNoChangeAfterApply(t *testing.T) {
	workspace := workspaceWithRootTaskfile(t)
	cfg := &config.Config{
		Tasks:        []string{"go"},
		IncludesDoc:  true,
		TargetFolder: "taskfiles",
		Workspace:    workspace,
	}
	o := &app.Orchestrator{StoreClient: &localStore{root: fixtureRoot(t)}}
	first, err := o.Run(context.Background(), cfg)
	if err != nil {
		t.Fatal(err)
	}
	if !first.Changed {
		t.Fatal("expected first run to change files")
	}
	second, err := o.Run(context.Background(), cfg)
	if err != nil {
		t.Fatal(err)
	}
	if second.Changed {
		t.Fatal("expected no changes on second run")
	}
}

func TestOrchestratorUnrelatedDirtyTreeFails(t *testing.T) {
	workspace := workspaceWithRootTaskfile(t)
	cfg := &config.Config{
		Tasks:        []string{"go"},
		IncludesDoc:  true,
		TargetFolder: "taskfiles",
		Workspace:    workspace,
	}
	o := &app.Orchestrator{
		StoreClient: &localStore{root: fixtureRoot(t)},
		GitOps:      &mockGitOps{unrelated: true},
	}
	if err := os.MkdirAll(filepath.Join(workspace, ".git"), 0o755); err != nil {
		t.Fatal(err)
	}
	_, err := o.Run(context.Background(), cfg)
	if err == nil {
		t.Fatal("expected unrelated changes error")
	}
	if _, statErr := os.Stat(filepath.Join(workspace, "taskfiles/go/Taskfile.yml")); statErr == nil {
		t.Fatal("workspace should not be mutated when git preconditions fail")
	}
}

func TestOrchestratorUpdatesExistingPR(t *testing.T) {
	workspace := workspaceWithRootTaskfile(t)
	cfg := &config.Config{
		Tasks:        []string{"go"},
		IncludesDoc:  true,
		TargetFolder: "taskfiles",
		Workspace:    workspace,
		Repository:   "owner/repo",
	}
	pr := &mockPR{find: &gh.PullRequest{Number: 7, URL: "https://example/pr/7"}}
	gitOps := &mockGitOps{defaultBranch: "main"}
	o := &app.Orchestrator{
		StoreClient: &localStore{root: fixtureRoot(t)},
		GitOps:      gitOps,
		PRClient:    pr,
	}
	if err := os.MkdirAll(filepath.Join(workspace, ".git"), 0o755); err != nil {
		t.Fatal(err)
	}
	result, err := o.Run(context.Background(), cfg)
	if err != nil {
		t.Fatal(err)
	}
	if pr.updated != 1 {
		t.Fatalf("expected PR body update, got %d", pr.updated)
	}
	if pr.lastBase != "main" {
		t.Fatalf("PR base = %q, want main", pr.lastBase)
	}
	if gitOps.defaultBranchCalls == 0 {
		t.Fatal("expected DefaultBranch before PR")
	}
	if result.PullRequestNumber != "7" {
		t.Fatalf("got PR number %q", result.PullRequestNumber)
	}
}

func TestOrchestratorCreatesPRWithResolvedBase(t *testing.T) {
	workspace := workspaceWithRootTaskfile(t)
	cfg := &config.Config{
		Tasks:        []string{"go"},
		IncludesDoc:  true,
		TargetFolder: "taskfiles",
		Workspace:    workspace,
		Repository:   "owner/repo",
	}
	pr := &mockPR{}
	gitOps := &mockGitOps{defaultBranch: "develop"}
	o := &app.Orchestrator{
		StoreClient: &localStore{root: fixtureRoot(t)},
		GitOps:      gitOps,
		PRClient:    pr,
	}
	if err := os.MkdirAll(filepath.Join(workspace, ".git"), 0o755); err != nil {
		t.Fatal(err)
	}
	if _, err := o.Run(context.Background(), cfg); err != nil {
		t.Fatal(err)
	}
	if pr.createdBase != "develop" {
		t.Fatalf("created PR base = %q, want develop", pr.createdBase)
	}
}

func TestNewOrchestratorInvalidRepository(t *testing.T) {
	cfg := &config.Config{Repository: "not-a-valid-repo"}
	_, err := app.NewOrchestrator(cfg)
	if err == nil {
		t.Fatal("expected repository parse error")
	}
}

var _ git.GitOps = (*mockGitOps)(nil)
var _ gh.PRClient = (*mockPR)(nil)
