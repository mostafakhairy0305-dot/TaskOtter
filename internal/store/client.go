package store

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/mostafakhairy0305-dot/taskotter-sync-action/internal/archive"
	"github.com/mostafakhairy0305-dot/taskotter-sync-action/internal/config"
	"gopkg.in/yaml.v3"
)

const (
	storeOwner = "mostafakhairy0305-dot"
	storeRepo  = "TaskOtter-store"
)

type RefInfo struct {
	Repository       string
	RequestedVersion string
	SourceRef        string
	ResolvedCommit   string
	DefaultBranch    string
}

type Snapshot struct {
	RootDir string
	Catalog map[string]struct{}
	Deps    map[string][]string
	Ref     RefInfo
	cleanup func() error
}

func (s *Snapshot) Close() error {
	if s.cleanup != nil {
		return s.cleanup()
	}
	return nil
}

func (s *Snapshot) ModuleDir(sourceModule string) string {
	return filepath.Join(s.RootDir, "taskfiles", sourceModule)
}

type HTTPDoer interface {
	Do(req *http.Request) (*http.Response, error)
}

type Client struct {
	httpClient HTTPDoer
	token      string
	baseURL    string
}

func NewClient(token string) *Client {
	return &Client{
		httpClient: &http.Client{Timeout: 60 * time.Second},
		token:      token,
		baseURL:    "https://api.github.com",
	}
}

func NewClientWithHTTP(token string, httpClient HTTPDoer) *Client {
	return &Client{
		httpClient: httpClient,
		token:      token,
		baseURL:    "https://api.github.com",
	}
}

func (c *Client) WithBaseURL(baseURL string) *Client {
	c.baseURL = strings.TrimRight(baseURL, "/")
	return c
}

func (c *Client) apiURL(path string) string {
	return c.baseURL + path
}

func (c *Client) ResolveRef(ctx context.Context, requestedVersion string) (RefInfo, error) {
	defaultBranch, err := c.getDefaultBranch(ctx)
	if err != nil {
		return RefInfo{}, err
	}

	info := RefInfo{
		Repository:       config.StoreRepository,
		RequestedVersion: requestedVersion,
		DefaultBranch:    defaultBranch,
	}

	if requestedVersion == "" {
		sha, err := c.resolveBranchHead(ctx, defaultBranch)
		if err != nil {
			return RefInfo{}, err
		}
		info.SourceRef = "refs/heads/" + defaultBranch
		info.ResolvedCommit = sha
		return info, nil
	}

	sha, err := c.resolveTag(ctx, requestedVersion)
	if err != nil {
		return RefInfo{}, err
	}
	info.SourceRef = "refs/tags/" + requestedVersion
	info.ResolvedCommit = sha
	return info, nil
}

func (c *Client) DownloadSnapshot(ctx context.Context, ref RefInfo) (*Snapshot, error) {
	url := c.apiURL(fmt.Sprintf("/repos/%s/%s/tarball/%s", storeOwner, storeRepo, ref.ResolvedCommit))
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	c.applyHeaders(req)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("download store archive: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusUnauthorized || resp.StatusCode == http.StatusForbidden {
		return nil, fmt.Errorf("authentication failed downloading store archive (HTTP %d)", resp.StatusCode)
	}
	if resp.StatusCode == http.StatusTooManyRequests {
		return nil, fmt.Errorf("GitHub rate limit exceeded downloading store archive")
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("download store archive failed with HTTP %d", resp.StatusCode)
	}

	tmpDir, err := os.MkdirTemp("", "taskotter-store-*")
	if err != nil {
		return nil, err
	}

	cleanup := func() error { return os.RemoveAll(tmpDir) }
	root, err := archive.ExtractTarGz(resp.Body, tmpDir)
	if err != nil {
		_ = cleanup()
		return nil, err
	}

	catalog, err := loadCatalog(root)
	if err != nil {
		_ = cleanup()
		return nil, err
	}
	deps, err := loadDeps(root, catalog)
	if err != nil {
		_ = cleanup()
		return nil, err
	}

	return &Snapshot{
		RootDir: root,
		Catalog: catalog,
		Deps:    deps,
		Ref:     ref,
		cleanup: cleanup,
	}, nil
}

func loadCatalog(root string) (map[string]struct{}, error) {
	taskfilesDir := filepath.Join(root, "taskfiles")
	entries, err := os.ReadDir(taskfilesDir)
	if err != nil {
		return nil, fmt.Errorf("load module catalog: %w", err)
	}
	catalog := make(map[string]struct{})
	for _, entry := range entries {
		if entry.IsDir() {
			catalog[entry.Name()] = struct{}{}
		}
	}
	return catalog, nil
}

func loadDeps(root string, catalog map[string]struct{}) (map[string][]string, error) {
	path := filepath.Join(root, ".deps.yml")
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read .deps.yml: %w", err)
	}
	var raw map[string][]string
	if err := yaml.Unmarshal(data, &raw); err != nil {
		return nil, fmt.Errorf("parse .deps.yml: %w", err)
	}
	for module, deps := range raw {
		if _, ok := catalog[module]; !ok {
			return nil, fmt.Errorf(".deps.yml references missing module %q", module)
		}
		for _, dep := range deps {
			if _, ok := catalog[dep]; !ok {
				return nil, fmt.Errorf(".deps.yml references missing dependency %q for module %q", dep, module)
			}
		}
	}
	return raw, nil
}

func (c *Client) applyHeaders(req *http.Request) {
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("X-GitHub-Api-Version", "2022-11-28")
	req.Header.Set("User-Agent", "taskotter-sync-action")
	if c.token != "" {
		req.Header.Set("Authorization", "Bearer "+c.token)
	}
}

func (c *Client) getJSON(ctx context.Context, path string, v any) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.apiURL(path), nil)
	if err != nil {
		return err
	}
	c.applyHeaders(req)
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("GitHub API %s failed with HTTP %d", path, resp.StatusCode)
	}
	return decodeJSON(resp.Body, v)
}

func (c *Client) getDefaultBranch(ctx context.Context) (string, error) {
	var payload struct {
		DefaultBranch string `json:"default_branch"`
	}
	path := fmt.Sprintf("/repos/%s/%s", storeOwner, storeRepo)
	if err := c.getJSON(ctx, path, &payload); err != nil {
		return "", fmt.Errorf("fetch store repository metadata: %w", err)
	}
	if payload.DefaultBranch == "" {
		return "", fmt.Errorf("store repository default branch is empty")
	}
	return payload.DefaultBranch, nil
}

func (c *Client) resolveBranchHead(ctx context.Context, branch string) (string, error) {
	var payload struct {
		SHA string `json:"sha"`
	}
	path := fmt.Sprintf("/repos/%s/%s/commits/%s", storeOwner, storeRepo, url.PathEscape(branch))
	if err := c.getJSON(ctx, path, &payload); err != nil {
		return "", fmt.Errorf("resolve branch %q: %w", branch, err)
	}
	return payload.SHA, nil
}

func (c *Client) resolveTag(ctx context.Context, tag string) (string, error) {
	var payload struct {
		Object struct {
			SHA  string `json:"sha"`
			Type string `json:"type"`
		} `json:"object"`
	}
	path := fmt.Sprintf("/repos/%s/%s/git/ref/tags/%s", storeOwner, storeRepo, url.PathEscape(tag))
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.apiURL(path), nil)
	if err != nil {
		return "", err
	}
	c.applyHeaders(req)
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusNotFound {
		return "", fmt.Errorf("store tag %q does not exist", tag)
	}
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("resolve tag %q failed with HTTP %d", tag, resp.StatusCode)
	}
	if err := decodeJSON(resp.Body, &payload); err != nil {
		return "", err
	}
	sha := payload.Object.SHA
	if payload.Object.Type == "tag" {
		sha, err = c.peelAnnotatedTag(ctx, sha)
		if err != nil {
			return "", err
		}
	}
	return sha, nil
}

func (c *Client) peelAnnotatedTag(ctx context.Context, sha string) (string, error) {
	var payload struct {
		Object struct {
			SHA string `json:"sha"`
		} `json:"object"`
	}
	path := fmt.Sprintf("/repos/%s/%s/git/tags/%s", storeOwner, storeRepo, sha)
	if err := c.getJSON(ctx, path, &payload); err != nil {
		return "", fmt.Errorf("resolve annotated tag: %w", err)
	}
	return payload.Object.SHA, nil
}

func decodeJSON(r io.Reader, v any) error {
	return json.NewDecoder(r).Decode(v)
}

// LocalSnapshot creates a snapshot from an on-disk directory (tests/fixtures).
func LocalSnapshot(root string, ref RefInfo) (*Snapshot, error) {
	catalog, err := loadCatalog(root)
	if err != nil {
		return nil, err
	}
	deps, err := loadDeps(root, catalog)
	if err != nil {
		return nil, err
	}
	return &Snapshot{RootDir: root, Catalog: catalog, Deps: deps, Ref: ref}, nil
}
