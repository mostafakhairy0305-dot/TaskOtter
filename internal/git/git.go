package git

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

const SyncCommitMessage = "chore(taskotter): sync taskfiles"

type GitOps interface {
	EnsureSafeDirectory(ctx context.Context) error
	HasUnrelatedChanges(ctx context.Context, allowed map[string]struct{}) (bool, error)
	CheckoutBranch(ctx context.Context, branch string, create bool) error
	BranchExists(ctx context.Context, branch string) (bool, error)
	LastCommitMessage(ctx context.Context, branch string) (string, error)
	Stage(ctx context.Context, paths []string) error
	Commit(ctx context.Context, message string) error
	Push(ctx context.Context, branch string, forceWithLease bool) error
	DefaultBranch(ctx context.Context) (string, error)
}

type Client struct {
	workspace string
}

func NewClient(workspace string) *Client {
	return &Client{workspace: workspace}
}

func (c *Client) EnsureSafeDirectory(ctx context.Context) error {
	_ = ctx
	// Safe directory is applied per command via -c; global/local config files are
	// not writable in GitHub Actions Docker containers (non-root user, read-only HOME).
	return nil
}

func (c *Client) gitArgs(args ...string) []string {
	return append([]string{"-c", "safe.directory=" + c.workspace}, args...)
}

func (c *Client) DefaultBranch(ctx context.Context) (string, error) {
	out, err := c.output(ctx, "symbolic-ref", "--short", "refs/remotes/origin/HEAD")
	if err != nil {
		return "", fmt.Errorf("detect default branch: %w", err)
	}
	branch := strings.TrimSpace(out)
	return strings.TrimPrefix(branch, "origin/"), nil
}

func (c *Client) HasUnrelatedChanges(ctx context.Context, allowed map[string]struct{}) (bool, error) {
	out, err := c.output(ctx, "status", "--porcelain")
	if err != nil {
		return false, err
	}
	lines := strings.Split(strings.TrimSpace(out), "\n")
	for _, line := range lines {
		if line == "" {
			continue
		}
		if len(line) < 4 {
			continue
		}
		path := strings.TrimSpace(line[3:])
		if path == "" {
			continue
		}
		if _, ok := allowed[path]; ok {
			continue
		}
		if isAllowedPath(path, allowed) {
			continue
		}
		return true, nil
	}
	return false, nil
}

func isAllowedPath(path string, allowed map[string]struct{}) bool {
	for allowedPath := range allowed {
		if path == allowedPath || strings.HasPrefix(path, allowedPath+"/") {
			return true
		}
	}
	return false
}

func (c *Client) CheckoutBranch(ctx context.Context, branch string, create bool) error {
	args := []string{"checkout"}
	if create {
		args = append(args, "-B")
	}
	args = append(args, branch)
	return c.run(ctx, args...)
}

func (c *Client) BranchExists(ctx context.Context, branch string) (bool, error) {
	_, err := c.output(ctx, "rev-parse", "--verify", branch)
	if err == nil {
		return true, nil
	}
	_, err = c.output(ctx, "rev-parse", "--verify", "refs/heads/"+branch)
	if err == nil {
		return true, nil
	}
	return false, nil
}

func (c *Client) LastCommitMessage(ctx context.Context, branch string) (string, error) {
	out, err := c.output(ctx, "log", "-1", "--format=%s", branch)
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(out), nil
}

func (c *Client) Stage(ctx context.Context, paths []string) error {
	if len(paths) == 0 {
		return nil
	}
	args := append([]string{"add", "--"}, paths...)
	return c.run(ctx, args...)
}

func (c *Client) Commit(ctx context.Context, message string) error {
	if err := c.run(ctx, "commit", "-m", message); err != nil {
		if strings.Contains(err.Error(), "nothing to commit") {
			return nil
		}
		return err
	}
	return nil
}

func (c *Client) Push(ctx context.Context, branch string, forceWithLease bool) error {
	args := []string{"push"}
	if forceWithLease {
		args = append(args, "--force-with-lease")
	}
	args = append(args, "origin", branch)
	return c.run(ctx, args...)
}

func (c *Client) run(ctx context.Context, args ...string) error {
	cmd := exec.CommandContext(ctx, "git", c.gitArgs(args...)...)
	cmd.Dir = c.workspace
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("git %s: %w: %s", strings.Join(args, " "), err, strings.TrimSpace(stderr.String()))
	}
	return nil
}

func (c *Client) output(ctx context.Context, args ...string) (string, error) {
	cmd := exec.CommandContext(ctx, "git", c.gitArgs(args...)...)
	cmd.Dir = c.workspace
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("git %s: %w: %s", strings.Join(args, " "), err, strings.TrimSpace(stderr.String()))
	}
	return stdout.String(), nil
}

func AllowedPathSet(paths []string) map[string]struct{} {
	out := make(map[string]struct{}, len(paths))
	for _, p := range paths {
		out[filepath.ToSlash(p)] = struct{}{}
	}
	return out
}

func WriteLocalIdentity(ctx context.Context, c GitOps) error {
	if client, ok := c.(*Client); ok {
		if err := client.run(ctx, "config", "user.email", "taskotter@users.noreply.github.com"); err != nil {
			return err
		}
		return client.run(ctx, "config", "user.name", "TaskOtter")
	}
	return nil
}

func IsGitRepo(workspace string) bool {
	_, err := os.Stat(filepath.Join(workspace, ".git"))
	return err == nil
}

func EnsureBranchOwned(ctx context.Context, g GitOps, branch string) error {
	exists, err := g.BranchExists(ctx, branch)
	if err != nil {
		return err
	}
	if !exists {
		return nil
	}
	msg, err := g.LastCommitMessage(ctx, branch)
	if err != nil {
		return err
	}
	if msg != SyncCommitMessage {
		return fmt.Errorf("branch %q exists but is not owned by TaskOtter", branch)
	}
	return nil
}
