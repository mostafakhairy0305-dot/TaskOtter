package github

import "context"

type PRClient interface {
	FindOpenPR(ctx context.Context, branch, base string) (*PullRequest, error)
	CreatePR(ctx context.Context, branch, base, body string) (*PullRequest, error)
	UpdatePRBody(ctx context.Context, number int, body string) error
}
