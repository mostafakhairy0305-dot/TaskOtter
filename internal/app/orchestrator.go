package app

import (
	"context"
	"fmt"

	"github.com/mostafakhairy0305-dot/TaskOtter/internal/config"
	"github.com/mostafakhairy0305-dot/TaskOtter/internal/dependency"
	"github.com/mostafakhairy0305-dot/TaskOtter/internal/git"
	gh "github.com/mostafakhairy0305-dot/TaskOtter/internal/github"
	"github.com/mostafakhairy0305-dot/TaskOtter/internal/logging"
	"github.com/mostafakhairy0305-dot/TaskOtter/internal/resolver"
	"github.com/mostafakhairy0305-dot/TaskOtter/internal/store"
	"github.com/mostafakhairy0305-dot/TaskOtter/internal/syncer"
)

type StoreClient interface {
	ResolveRef(ctx context.Context, requestedVersion string) (store.RefInfo, error)
	DownloadSnapshot(ctx context.Context, ref store.RefInfo) (*store.Snapshot, error)
}

type Orchestrator struct {
	Logger      *logging.Logger
	StoreClient StoreClient
	GitOps      git.GitOps
	PRClient    gh.PRClient
}

func Run(ctx context.Context, cfg *config.Config) (*Result, error) {
	o, err := NewOrchestrator(cfg)
	if err != nil {
		return nil, err
	}
	return o.Run(ctx, cfg)
}

func NewOrchestrator(cfg *config.Config) (*Orchestrator, error) {
	o := &Orchestrator{
		Logger:      logging.New(),
		StoreClient: store.NewClient(cfg.GitHubToken),
		GitOps:      git.NewClient(cfg.Workspace),
	}
	if cfg.Repository != "" {
		prClient, err := gh.NewClient(cfg.GitHubToken, cfg.Repository)
		if err != nil {
			return nil, err
		}
		o.PRClient = prClient
	}
	return o, nil
}

func (o *Orchestrator) wireDefaults(cfg *config.Config) error {
	if o.Logger == nil {
		o.Logger = logging.New()
	}
	if o.StoreClient == nil {
		o.StoreClient = store.NewClient(cfg.GitHubToken)
	}
	if o.GitOps == nil {
		o.GitOps = git.NewClient(cfg.Workspace)
	}
	if o.PRClient == nil && cfg.Repository != "" {
		prClient, err := gh.NewClient(cfg.GitHubToken, cfg.Repository)
		if err != nil {
			return err
		}
		o.PRClient = prClient
	}
	return nil
}

func (o *Orchestrator) Run(ctx context.Context, cfg *config.Config) (*Result, error) {
	if err := o.wireDefaults(cfg); err != nil {
		return nil, err
	}
	return o.run(ctx, cfg)
}

func (o *Orchestrator) runGitPreconditions(ctx context.Context, cfg *config.Config, plan *syncer.Plan) (string, error) {
	if err := o.GitOps.EnsureSafeDirectory(ctx); err != nil {
		return "", err
	}
	if err := git.WriteLocalIdentity(ctx, o.GitOps); err != nil {
		return "", err
	}
	if err := git.EnsureBranchOwned(ctx, o.GitOps, cfg.BranchName); err != nil {
		return "", err
	}
	allowed := git.AllowedPathSet(plan.StagePaths)
	unrelated, err := o.GitOps.HasUnrelatedChanges(ctx, allowed)
	if err != nil {
		return "", err
	}
	if unrelated {
		return "", fmt.Errorf("unrelated uncommitted changes detected in workspace")
	}
	return o.GitOps.DefaultBranch(ctx)
}

func (o *Orchestrator) runGitSync(ctx context.Context, cfg *config.Config, plan *syncer.Plan) error {
	if err := o.GitOps.CheckoutBranch(ctx, cfg.BranchName, true); err != nil {
		return err
	}
	if err := o.GitOps.Stage(ctx, plan.StagePaths); err != nil {
		return err
	}
	if err := o.GitOps.Commit(ctx, git.SyncCommitMessage); err != nil {
		return err
	}
	return o.GitOps.Push(ctx, cfg.BranchName, true)
}

func (o *Orchestrator) runPR(ctx context.Context, cfg *config.Config, plan *syncer.Plan, ref store.RefInfo, defaultBranch string, result *Result) error {
	body := gh.BuildPRBody(cfg, plan, gh.StoreRefFrom(ref))
	existing, err := o.PRClient.FindOpenPR(ctx, cfg.BranchName, defaultBranch)
	if err != nil {
		return err
	}
	if existing != nil {
		if err := o.PRClient.UpdatePRBody(ctx, existing.Number, body); err != nil {
			return err
		}
		result.PullRequestNumber = fmt.Sprintf("%d", existing.Number)
		result.PullRequestURL = existing.URL
		o.Logger.Printf("Updated pull request #%d", existing.Number)
		return nil
	}
	pr, err := o.PRClient.CreatePR(ctx, cfg.BranchName, defaultBranch, body)
	if err != nil {
		return err
	}
	result.PullRequestNumber = fmt.Sprintf("%d", pr.Number)
	result.PullRequestURL = pr.URL
	o.Logger.Printf("Created pull request #%d", pr.Number)
	return nil
}

func (o *Orchestrator) run(ctx context.Context, cfg *config.Config) (*Result, error) {
	o.Logger.Group("Validate inputs", func() {
		o.Logger.Printf("Validated %d task(s)", len(cfg.Tasks))
		o.Logger.Printf("Target folder: %s", cfg.TargetFolder)
	})

	var ref store.RefInfo
	var refErr error
	o.Logger.Group("Resolve source version", func() {
		ref, refErr = o.StoreClient.ResolveRef(ctx, cfg.StoreVersion)
		if refErr != nil {
			return
		}
		o.Logger.Printf("Source ref: %s", ref.SourceRef)
		o.Logger.Printf("Resolved commit: %s", ref.ResolvedCommit)
	})
	if refErr != nil {
		return nil, refErr
	}

	var snapshot *store.Snapshot
	var snapErr error
	o.Logger.Group("Download store", func() {
		snapshot, snapErr = o.StoreClient.DownloadSnapshot(ctx, ref)
		if snapErr != nil {
			return
		}
		o.Logger.Printf("Loaded store snapshot from %s", ref.ResolvedCommit)
	})
	if snapErr != nil {
		return nil, snapErr
	}
	defer snapshot.Close()

	o.Logger.Group("Load module catalog", func() {
		o.Logger.Printf("Catalog modules: %d", len(snapshot.Catalog))
	})

	var resolutions []resolver.Resolution
	var resolveErr error
	o.Logger.Group("Resolve requested modules", func() {
		resolutions, resolveErr = resolver.ResolveAll(cfg.Tasks, snapshot.Catalog, cfg.NodePackageManager, cfg.NodeVersionManager)
		if resolveErr != nil {
			return
		}
		for _, res := range resolutions {
			o.Logger.Printf("%s -> %s", res.LogicalTask, res.SourceModule)
		}
	})
	if resolveErr != nil {
		return nil, resolveErr
	}

	var depSources []string
	var depErr error
	o.Logger.Group("Resolve dependencies", func() {
		requestedSources := make([]string, 0, len(resolutions))
		for _, res := range resolutions {
			requestedSources = append(requestedSources, res.SourceModule)
		}
		depSources, depErr = dependency.ResolveTransitive(requestedSources, snapshot.Deps)
		if depErr != nil {
			return
		}
		for _, dep := range depSources {
			o.Logger.Printf("dependency: %s", dep)
		}
	})
	if depErr != nil {
		return nil, depErr
	}

	syncInput, err := PrepareSyncInput(cfg, snapshot, resolutions, depSources)
	if err != nil {
		return nil, err
	}
	o.Logger.Group("Normalize destination names", func() {
		for source, dest := range syncInput.SourceToDest {
			o.Logger.Printf("%s -> %s", source, dest)
		}
	})

	var plan *syncer.Plan
	var planErr error
	o.Logger.Group("Compare managed files", func() {
		plan, planErr = syncer.BuildPlan(syncInput)
		if planErr != nil {
			return
		}
		o.Logger.Printf("Changed: %t", plan.Changed)
		o.Logger.Printf("Added: %d Updated: %d Removed: %d", len(plan.Added), len(plan.Updated), len(plan.Removed))
	})
	if planErr != nil {
		return nil, planErr
	}

	result, err := buildResult(cfg, plan, ref)
	if err != nil {
		return nil, err
	}

	if !plan.Changed {
		o.Logger.Group("Summary", func() {
			printSummary(o.Logger, cfg, plan, result, "")
		})
		return result, nil
	}

	var defaultBranch string
	if git.IsGitRepo(cfg.Workspace) {
		defaultBranch, err = o.runGitPreconditions(ctx, cfg, plan)
		if err != nil {
			return nil, err
		}
	}

	var applyErr error
	o.Logger.Group("Copy task modules", func() {
		applyErr = syncer.ApplyPlan(plan, syncInput)
		if applyErr != nil {
			return
		}
		o.Logger.Printf("Copied modules and validated generated YAML")
	})
	if applyErr != nil {
		return nil, applyErr
	}

	if git.IsGitRepo(cfg.Workspace) {
		var gitErr error
		o.Logger.Group("Create synchronization commit", func() {
			gitErr = o.runGitSync(ctx, cfg, plan)
		})
		if gitErr != nil {
			return nil, gitErr
		}
	}

	if git.IsGitRepo(cfg.Workspace) && o.PRClient != nil {
		var prErr error
		o.Logger.Group("Create or update pull request", func() {
			prErr = o.runPR(ctx, cfg, plan, ref, defaultBranch, result)
		})
		if prErr != nil {
			return nil, prErr
		}
	}

	o.Logger.Group("Summary", func() {
		printSummary(o.Logger, cfg, plan, result, result.PullRequestURL)
	})

	return result, nil
}
