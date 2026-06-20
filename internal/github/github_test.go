package github_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/mostafakhairy0305-dot/TaskOtter/internal/config"
	gh "github.com/mostafakhairy0305-dot/TaskOtter/internal/github"
	"github.com/mostafakhairy0305-dot/TaskOtter/internal/store"
	"github.com/mostafakhairy0305-dot/TaskOtter/internal/syncer"
)

func TestWriteOutputsMultilineJSON(t *testing.T) {
	path := filepath.Join(t.TempDir(), "output")
	values := map[string]string{
		"changed":        "true",
		"resolved-tasks": "{\n  \"go\": {}\n}\n",
	}
	if err := gh.WriteOutputs(path, values); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	text := string(data)
	if !strings.Contains(text, "resolved-tasks<<EOF") {
		t.Fatalf("expected heredoc output: %s", text)
	}
}

func TestBuildPRBody(t *testing.T) {
	cfg := &config.Config{
		Tasks:              []string{"eslint"},
		TargetFolder:       "taskfiles",
		NodePackageManager: config.PMPnpm,
		NodeVersionManager: config.VMFnm,
		IncludesDoc:        true,
	}
	plan := &syncer.Plan{
		Requested: map[string]syncer.ModuleRecord{
			"eslint": {SourceModule: "eslint-pnpm-fnm", DestinationModule: "eslint", Path: "taskfiles/eslint"},
		},
		Dependencies: []syncer.ModuleRecord{
			{SourceModule: "pnpm-fnm", DestinationModule: "pnpm", Path: "taskfiles/pnpm"},
		},
	}
	body := gh.BuildPRBody(cfg, plan, gh.StoreRefFrom(store.RefInfo{
		SourceRef: "refs/heads/main", ResolvedCommit: "abc", DefaultBranch: "main",
	}))
	if !strings.Contains(body, "eslint-pnpm-fnm") {
		t.Fatalf("missing module info: %s", body)
	}
}
