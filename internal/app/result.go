package app

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"

	"github.com/mostafakhairy0305-dot/taskotter-sync-action/internal/config"
	gh "github.com/mostafakhairy0305-dot/taskotter-sync-action/internal/github"
	"github.com/mostafakhairy0305-dot/taskotter-sync-action/internal/logging"
	"github.com/mostafakhairy0305-dot/taskotter-sync-action/internal/store"
	"github.com/mostafakhairy0305-dot/taskotter-sync-action/internal/syncer"
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
