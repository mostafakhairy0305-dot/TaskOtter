package app

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"

	"github.com/mostafakhairy0305-dot/TaskOtter/internal/config"
	gh "github.com/mostafakhairy0305-dot/TaskOtter/internal/github"
	"github.com/mostafakhairy0305-dot/TaskOtter/internal/logging"
	"github.com/mostafakhairy0305-dot/TaskOtter/internal/store"
	"github.com/mostafakhairy0305-dot/TaskOtter/internal/syncer"
)

type Result struct {
	Changed              bool
	StoreVersion         string
	SourceRef            string
	SourceSHA            string
	TargetFolder         string
	ResolvedTasksJSON    string
	ResolvedDependencies string
	PullRequestNumber    string
	PullRequestURL       string
	Plan                 *syncer.Plan
	Ref                  store.RefInfo
}

type ResolvedTask struct {
	SourceModule      string `json:"source_module"`
	DestinationModule string `json:"destination_module"`
	Path              string `json:"path"`
}

func buildResolvedTasksJSON(requested map[string]syncer.ModuleRecord) (string, error) {
	out := make(map[string]ResolvedTask, len(requested))
	for task, rec := range requested {
		out[task] = ResolvedTask{
			SourceModule:      rec.SourceModule,
			DestinationModule: rec.DestinationModule,
			Path:              rec.Path,
		}
	}
	data, err := json.MarshalIndent(out, "", "  ")
	if err != nil {
		return "", fmt.Errorf("marshal resolved tasks: %w", err)
	}
	return string(data), nil
}

func buildResolvedDependenciesJSON(deps []syncer.ModuleRecord) (string, error) {
	data, err := json.MarshalIndent(deps, "", "  ")
	if err != nil {
		return "", fmt.Errorf("marshal resolved dependencies: %w", err)
	}
	return string(data), nil
}

func buildResult(cfg *config.Config, plan *syncer.Plan, ref store.RefInfo) (*Result, error) {
	result := &Result{
		Changed:      plan.Changed,
		StoreVersion: cfg.StoreVersion,
		SourceRef:    ref.SourceRef,
		SourceSHA:    ref.ResolvedCommit,
		TargetFolder: cfg.TargetFolder,
		Plan:         plan,
		Ref:          ref,
	}
	var err error
	result.ResolvedTasksJSON, err = buildResolvedTasksJSON(plan.Requested)
	if err != nil {
		return nil, err
	}
	result.ResolvedDependencies, err = buildResolvedDependenciesJSON(plan.Dependencies)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func printSummary(log *logging.Logger, cfg *config.Config, plan *syncer.Plan, result *Result, prURL string) {
	log.Printf("Requested tasks: %v", cfg.Tasks)
	for _, task := range cfg.Tasks {
		rec := plan.Requested[task]
		log.Printf("Source module %s -> %s", rec.SourceModule, rec.Path)
	}
	for _, dep := range plan.Dependencies {
		log.Printf("Dependency %s -> %s", dep.SourceModule, dep.Path)
	}
	log.Printf("Store version: %s", empty(result.StoreVersion))
	log.Printf("Source SHA: %s", result.SourceSHA)
	log.Printf("Target folder: %s", result.TargetFolder)
	log.Printf("Files added: %d", len(plan.Added))
	log.Printf("Files updated: %d", len(plan.Updated))
	log.Printf("Files removed: %d", len(plan.Removed))
	if prURL != "" {
		log.Printf("Pull request: %s", prURL)
	} else {
		log.Printf("Pull request result: none")
	}
}

func empty(v string) string {
	if v == "" {
		return "(latest default branch)"
	}
	return v
}

func SyncRequired(result *Result) bool {
	return result.Changed
}

func ReportSyncRequired(result *Result) {
	var summary string
	if result.PullRequestURL != "" {
		prNumber := result.PullRequestNumber
		if prNumber == "" {
			prNumber = "unknown"
		}
		summary = fmt.Sprintf("TaskOtter opened sync PR #%s: %s", prNumber, result.PullRequestURL)
	} else {
		summary = "TaskOtter synced taskfile changes but did not return a pull request URL."
	}
	fmt.Fprintf(os.Stderr, "::error title=TaskOtter sync required::%s Merge the sync pull request to update taskfiles, then re-run this workflow.\n", summary)
	fmt.Fprintln(os.Stderr, "::notice title=What happened::TaskOtter compared managed files with the store and found drift. This job fails intentionally until the sync PR is merged.")
}

func ReportSyncUpToDate(result *Result) {
	fmt.Fprintf(os.Stdout, "::notice title=TaskOtter sync up to date::Managed taskfiles match the store. No sync pull request was created.\n")
	fmt.Fprintf(os.Stdout, "Store source SHA: %s\n", result.SourceSHA)
}

func WriteActionOutputs(cfg *config.Config, result *Result) error {
	values := map[string]string{
		"changed":               fmt.Sprintf("%t", result.Changed),
		"store-version":         result.StoreVersion,
		"source-ref":            result.SourceRef,
		"source-sha":            result.SourceSHA,
		"target-folder":         result.TargetFolder,
		"resolved-tasks":        result.ResolvedTasksJSON,
		"resolved-dependencies": result.ResolvedDependencies,
		"pull-request-number":   result.PullRequestNumber,
		"pull-request-url":      result.PullRequestURL,
	}
	if cfg.GitHubOutput != "" {
		return gh.WriteOutputs(cfg.GitHubOutput, values)
	}
	keys := make([]string, 0, len(values))
	for k := range values {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		fmt.Fprintf(os.Stdout, "%s=%s\n", k, values[k])
	}
	return nil
}
