// internal/github/client.go
package github

import (
	"context"
	"errors"
	"os"

	"github.com/google/go-github/v58/github"
	"golang.org/x/oauth2"
)

// Client defines methods required from GitHub API.
type Client interface {
	CreateRelease(ctx context.Context, owner, repo string, release *github.RepositoryRelease) (*github.RepositoryRelease, error)
	UploadReleaseAsset(ctx context.Context, owner, repo string, releaseID int64, opts *github.UploadOptions, file *os.File) (*github.ReleaseAsset, error)
	GetLatestRelease(ctx context.Context, owner, repo string) (*github.RepositoryRelease, error)
	ListReleases(ctx context.Context, owner, repo string, opts *github.ListOptions) ([]*github.RepositoryRelease, *github.Response, error)
	UpdateRelease(ctx context.Context, owner, repo string, releaseID int64, release *github.RepositoryRelease) (*github.RepositoryRelease, error)
}

type realClient struct {
	*github.Client
}

// NewRealClient creates a GitHub API client using the provided OAuth token.
func NewRealClient(token string) (Client, error) {
	if token == "" {
		return nil, errors.New("GitHub token is required")
	}
	ctx := context.Background()
	ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: token})
	tc := oauth2.NewClient(ctx, ts)
	client := github.NewClient(tc)
	return &realClient{client}, nil
}

func (c *realClient) CreateRelease(ctx context.Context, owner, repo string, release *github.RepositoryRelease) (*github.RepositoryRelease, error) {
	rel, _, err := c.Repositories.CreateRelease(ctx, owner, repo, release)
	return rel, err
}

func (c *realClient) UploadReleaseAsset(ctx context.Context, owner, repo string, releaseID int64, opts *github.UploadOptions, file *os.File) (*github.ReleaseAsset, error) {
	asset, _, err := c.Repositories.UploadReleaseAsset(ctx, owner, repo, releaseID, opts, file)
	return asset, err
}

func (c *realClient) GetLatestRelease(ctx context.Context, owner, repo string) (*github.RepositoryRelease, error) {
	rel, _, err := c.Repositories.GetLatestRelease(ctx, owner, repo)
	return rel, err
}

func (c *realClient) ListReleases(ctx context.Context, owner, repo string, opts *github.ListOptions) ([]*github.RepositoryRelease, *github.Response, error) {
	return c.Repositories.ListReleases(ctx, owner, repo, opts)
}

func (c *realClient) UpdateRelease(ctx context.Context, owner, repo string, releaseID int64, release *github.RepositoryRelease) (*github.RepositoryRelease, error) {
	rel, _, err := c.Repositories.EditRelease(ctx, owner, repo, releaseID, release)
	return rel, err
}
