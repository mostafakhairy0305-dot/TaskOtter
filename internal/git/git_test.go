package git_test

import (
	"context"
	"testing"

	"github.com/mostafakhairy0305-dot/taskotter-sync-action/internal/git"
)

type mockGitOps struct {
	branchExists bool
	lastMessage  string
}

func (m *mockGitOps) EnsureSafeDirectory(context.Context) error              { return nil }
func (m *mockGitOps) HasUnrelatedChanges(context.Context, map[string]struct{}) (bool, error) {
	return false, nil
}
func (m *mockGitOps) CheckoutBranch(context.Context, string, bool) error { return nil }
func (m *mockGitOps) BranchExists(context.Context, string) (bool, error) {
	return m.branchExists, nil
}
func (m *mockGitOps) LastCommitMessage(context.Context, string) (string, error) {
	return m.lastMessage, nil
}
func (m *mockGitOps) Stage(context.Context, []string) error              { return nil }
func (m *mockGitOps) Commit(context.Context, string) error                 { return nil }
func (m *mockGitOps) Push(context.Context, string, bool) error           { return nil }
func (m *mockGitOps) DefaultBranch(context.Context) (string, error)      { return "main", nil }

func TestEnsureBranchOwnedAllowsNewBranch(t *testing.T) {
	if err := git.EnsureBranchOwned(context.Background(), &mockGitOps{branchExists: false}, "taskotter/sync-abc"); err != nil {
		t.Fatal(err)
	}
}

func TestEnsureBranchOwnedAllowsTaskOtterBranch(t *testing.T) {
	m := &mockGitOps{branchExists: true, lastMessage: git.SyncCommitMessage}
	if err := git.EnsureBranchOwned(context.Background(), m, "taskotter/sync-abc"); err != nil {
		t.Fatal(err)
	}
}

func TestEnsureBranchOwnedRejectsForeignBranch(t *testing.T) {
	m := &mockGitOps{branchExists: true, lastMessage: "feat: custom work"}
	err := git.EnsureBranchOwned(context.Background(), m, "taskotter/sync-abc")
	if err == nil {
		t.Fatal("expected branch ownership error")
	}
}
