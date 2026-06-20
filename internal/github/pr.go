package github

import "context"

// PRClient creates and updates TaskOtter sync pull requests.
type PRClient interface {
	FindOpenPR(ctx context.Context, branch, base string) (*PullRequest, error)
	CreatePR(ctx context.Context, branch, base, body string) (*PullRequest, error)
	UpdatePRBody(ctx context.Context, number int, body string) error
}
