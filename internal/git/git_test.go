package git_test

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/mostafakhairy0305-dot/TaskOtter/internal/git"
)

type mockGitOps struct {
	branchExists bool
	lastMessage  string
}

func (m *mockGitOps) EnsureSafeDirectory(context.Context) error { return nil }
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
func (m *mockGitOps) Stage(context.Context, []string) error         { return nil }
func (m *mockGitOps) Commit(context.Context, string) error          { return nil }
func (m *mockGitOps) Push(context.Context, string, bool) error      { return nil }
func (m *mockGitOps) DefaultBranch(context.Context) (string, error) { return "main", nil }

func runGit(t *testing.T, dir string, args ...string) string {
	t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("git %s in %s: %v\n%s", strings.Join(args, " "), dir, err, out)
	}
	return string(out)
}

func setupRemoteRepo(t *testing.T) (bareDir, cloneDir string) {
	t.Helper()
	root := t.TempDir()
	bareDir = filepath.Join(root, "bare.git")
	cloneDir = filepath.Join(root, "clone")
	if err := os.MkdirAll(bareDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(cloneDir, 0o755); err != nil {
		t.Fatal(err)
	}

	runGit(t, bareDir, "init", "--bare", "-b", "main")
	runGit(t, cloneDir, "init", "-b", "main")
	runGit(t, cloneDir, "config", "user.email", "test@test.com")
	runGit(t, cloneDir, "config", "user.name", "Test")
	if err := os.WriteFile(filepath.Join(cloneDir, "README.md"), []byte("init\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	runGit(t, cloneDir, "add", "README.md")
	runGit(t, cloneDir, "commit", "-m", "init")
	runGit(t, cloneDir, "remote", "add", "origin", bareDir)
	runGit(t, cloneDir, "push", "-u", "origin", "main")
	runGit(t, cloneDir, "fetch", "origin")
	return bareDir, cloneDir
}

func TestDefaultBranchSymbolicRef(t *testing.T) {
	_, cloneDir := setupRemoteRepo(t)
	runGit(t, cloneDir, "remote", "set-head", "origin", "-a")

	client := git.NewClient(cloneDir)
	branch, err := client.DefaultBranch(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if branch != "main" {
		t.Fatalf("branch = %q, want main", branch)
	}
}

func TestDefaultBranchMissingOriginHEAD(t *testing.T) {
	_, cloneDir := setupRemoteRepo(t)
	originHEAD := filepath.Join(cloneDir, ".git", "refs", "remotes", "origin", "HEAD")
	if err := os.Remove(originHEAD); err != nil && !os.IsNotExist(err) {
		t.Fatal(err)
	}

	client := git.NewClient(cloneDir)
	branch, err := client.DefaultBranch(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if branch != "main" {
		t.Fatalf("branch = %q, want main", branch)
	}
}

func TestDefaultBranchDirectRef(t *testing.T) {
	_, cloneDir := setupRemoteRepo(t)
	mainSHA := strings.TrimSpace(runGit(t, cloneDir, "rev-parse", "refs/remotes/origin/main"))
	originHEAD := filepath.Join(cloneDir, ".git", "refs", "remotes", "origin", "HEAD")
	if err := os.WriteFile(originHEAD, []byte(mainSHA+"\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	client := git.NewClient(cloneDir)
	branch, err := client.DefaultBranch(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if branch != "main" {
		t.Fatalf("branch = %q, want main", branch)
	}
}

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
