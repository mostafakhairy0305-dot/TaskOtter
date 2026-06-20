package github

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/google/go-github/v69/github"
	"github.com/mostafakhairy0305-dot/TaskOtter/internal/config"
	"github.com/mostafakhairy0305-dot/TaskOtter/internal/repo"
	"github.com/mostafakhairy0305-dot/TaskOtter/internal/store"
	"github.com/mostafakhairy0305-dot/TaskOtter/internal/syncer"
	"golang.org/x/oauth2"
)

const (
	prTitle = "chore(taskotter): sync taskfiles"
)

type PullRequest struct {
	Number int
	URL    string
}

type Client struct {
	api   *github.Client
	owner string
	repo  string
}

func NewClient(token, repository string) (*Client, error) {
	owner, repoName, err := repo.Parse(repository)
	if err != nil {
		return nil, err
	}
	ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: token})
	return &Client{
		api:   github.NewClient(oauth2.NewClient(context.Background(), ts)),
		owner: owner,
		repo:  repoName,
	}, nil
}

func (c *Client) FindOpenPR(ctx context.Context, branch, base string) (*PullRequest, error) {
	head := fmt.Sprintf("%s:%s", c.owner, branch)
	prs, _, err := c.api.PullRequests.List(ctx, c.owner, c.repo, &github.PullRequestListOptions{
		State: "open",
		Head:  head,
		Base:  base,
	})
	if err != nil {
		return nil, fmt.Errorf("list pull requests: %w", err)
	}
	if len(prs) == 0 {
		return nil, nil
	}
	pr := prs[0]
	return &PullRequest{Number: pr.GetNumber(), URL: pr.GetHTMLURL()}, nil
}

func (c *Client) CreatePR(ctx context.Context, branch, base, body string) (*PullRequest, error) {
	newPR := &github.NewPullRequest{
		Title: github.Ptr(prTitle),
		Head:  github.Ptr(branch),
		Base:  github.Ptr(base),
		Body:  github.Ptr(body),
	}
	pr, _, err := c.api.PullRequests.Create(ctx, c.owner, c.repo, newPR)
	if err != nil {
		return nil, fmt.Errorf("create pull request: %w", err)
	}
	return &PullRequest{Number: pr.GetNumber(), URL: pr.GetHTMLURL()}, nil
}

func (c *Client) UpdatePRBody(ctx context.Context, number int, body string) error {
	_, _, err := c.api.PullRequests.Edit(ctx, c.owner, c.repo, number, &github.PullRequest{
		Body: github.Ptr(body),
	})
	if err != nil {
		return fmt.Errorf("update pull request: %w", err)
	}
	return nil
}

type storeRef struct {
	SourceRef      string
	ResolvedCommit string
	DefaultBranch  string
}

func StoreRefFrom(ref store.RefInfo) storeRef {
	return storeRef{
		SourceRef:      ref.SourceRef,
		ResolvedCommit: ref.ResolvedCommit,
		DefaultBranch:  ref.DefaultBranch,
	}
}

func BuildPRBody(cfg *config.Config, plan *syncer.Plan, ref storeRef) string {
	var b strings.Builder
	b.WriteString("## TaskOtter\n\n")
	b.WriteString(fmt.Sprintf("- Source: `%s`\n", config.StoreRepository))
	b.WriteString(fmt.Sprintf("- Requested version: `%s`\n", emptyDash(cfg.StoreVersion)))
	b.WriteString(fmt.Sprintf("- Source reference: `%s`\n", ref.SourceRef))
	b.WriteString(fmt.Sprintf("- Resolved commit: `%s`\n", ref.ResolvedCommit))
	b.WriteString(fmt.Sprintf("- Default branch: `%s`\n", ref.DefaultBranch))
	b.WriteString(fmt.Sprintf("- Target folder: `%s`\n", cfg.TargetFolder))
	b.WriteString(fmt.Sprintf("- Documentation included: `%t`\n", cfg.IncludesDoc))
	b.WriteString(fmt.Sprintf("- JS runtime: `%s`\n", emptyDash(string(cfg.JSRuntime))))
	if cfg.JSRuntime == config.JSRuntimeNodeJS {
		b.WriteString(fmt.Sprintf("- Package manager: `%s`\n", cfg.NodePackageManager))
		b.WriteString(fmt.Sprintf("- Version manager: `%s`\n", cfg.NodeVersionManager))
	}
	b.WriteString("\n")

	b.WriteString("### Requested modules\n\n")
	b.WriteString("| Task | Source module | Destination |\n")
	b.WriteString("|---|---|---|\n")
	for _, task := range cfg.Tasks {
		rec := plan.Requested[task]
		b.WriteString(fmt.Sprintf("| %s | `%s` | `%s` |\n", task, rec.SourceModule, rec.Path))
	}

	b.WriteString("\n### Dependencies\n\n")
	b.WriteString("| Source module | Destination |\n")
	b.WriteString("|---|---|\n")
	for _, dep := range plan.Dependencies {
		b.WriteString(fmt.Sprintf("| `%s` | `%s` |\n", dep.SourceModule, dep.Path))
	}

	b.WriteString("\n### File changes\n\n")
	b.WriteString(fmt.Sprintf("- Added: %d\n", len(plan.Added)))
	for _, p := range plan.Added {
		b.WriteString(fmt.Sprintf("  - `%s`\n", p))
	}
	b.WriteString(fmt.Sprintf("- Updated: %d\n", len(plan.Updated)))
	for _, p := range plan.Updated {
		b.WriteString(fmt.Sprintf("  - `%s`\n", p))
	}
	b.WriteString(fmt.Sprintf("- Removed: %d\n", len(plan.Removed)))
	for _, p := range plan.Removed {
		b.WriteString(fmt.Sprintf("  - `%s`\n", p))
	}
	return b.String()
}

func emptyDash(v string) string {
	if v == "" {
		return ""
	}
	return v
}

func WriteOutputs(path string, values map[string]string) error {
	if path == "" {
		return nil
	}
	var b strings.Builder
	for key, value := range values {
		if strings.Contains(value, "\n") {
			b.WriteString(key)
			b.WriteString("<<EOF\n")
			b.WriteString(value)
			b.WriteString("\nEOF\n")
			continue
		}
		b.WriteString(key)
		b.WriteString("=")
		b.WriteString(value)
		b.WriteString("\n")
	}
	return os.WriteFile(path, []byte(b.String()), 0o644)
}
