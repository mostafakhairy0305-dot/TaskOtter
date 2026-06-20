package app

import (
	"github.com/mostafakhairy0305-dot/taskotter-sync-action/internal/config"
	"github.com/mostafakhairy0305-dot/taskotter-sync-action/internal/normalizer"
	"github.com/mostafakhairy0305-dot/taskotter-sync-action/internal/pathutil"
	"github.com/mostafakhairy0305-dot/taskotter-sync-action/internal/resolver"
	"github.com/mostafakhairy0305-dot/taskotter-sync-action/internal/store"
	"github.com/mostafakhairy0305-dot/taskotter-sync-action/internal/syncer"
)

func PrepareSyncInput(cfg *config.Config, snapshot *store.Snapshot, resolutions []resolver.Resolution, depSources []string) (syncer.SyncInput, error) {
	requestedSources := make([]string, 0, len(resolutions))
	for _, res := range resolutions {
		requestedSources = append(requestedSources, res.SourceModule)
	}
	allSources := append(append([]string{}, requestedSources...), depSources...)
	sourceToDest, err := normalizer.BuildDestinationMap(allSources)
	if err != nil {
		return syncer.SyncInput{}, err
	}

	requestedRecords := make(map[string]syncer.ModuleRecord)
	destByTask := make(map[string]string)
	for _, res := range resolutions {
		dest := sourceToDest[res.SourceModule]
		requestedRecords[res.LogicalTask] = syncer.ModuleRecord{
			SourceModule:      res.SourceModule,
			DestinationModule: dest,
			Path:              pathutil.JoinRelative(cfg.TargetFolder, dest),
		}
		destByTask[res.LogicalTask] = dest
	}

	dependencyRecords := make([]syncer.ModuleRecord, 0, len(depSources))
	for _, dep := range depSources {
		dest := sourceToDest[dep]
		dependencyRecords = append(dependencyRecords, syncer.ModuleRecord{
			SourceModule:      dep,
			DestinationModule: dest,
			Path:              pathutil.JoinRelative(cfg.TargetFolder, dest),
		})
	}

	return syncer.SyncInput{
		Config:       cfg,
		Snapshot:     snapshot,
		Requested:    requestedRecords,
		Dependencies: dependencyRecords,
		SourceToDest: sourceToDest,
		DestByTask:   destByTask,
	}, nil
}
