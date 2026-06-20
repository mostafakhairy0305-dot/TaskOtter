package syncer

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"github.com/mostafakhairy0305-dot/TaskOtter/internal/config"
	"github.com/mostafakhairy0305-dot/TaskOtter/internal/pathutil"
	"github.com/mostafakhairy0305-dot/TaskOtter/internal/taskfile"
	"gopkg.in/yaml.v3"
)

func mustMarshalMetadata(meta Metadata) []byte {
	data, err := yaml.Marshal(meta)
	if err != nil {
		return nil
	}
	return data
}

func BuildPlan(in SyncInput) (*Plan, error) {
	workspace := in.Config.Workspace
	oldLock, _, oldTarget, err := loadPreviousState(workspace, in.Config)
	if err != nil {
		return nil, err
	}

	plannedFiles, moduleContents, err := planManagedFiles(in, oldLock)
	if err != nil {
		return nil, err
	}

	rootPath := pathutil.WorkspacePath(workspace, "Taskfile.yml")
	rootBytes, err := os.ReadFile(rootPath)
	rootExisted := true
	if err != nil {
		if os.IsNotExist(err) {
			rootBytes = taskfile.NewRootTemplate()
			rootExisted = false
		} else {
			return nil, err
		}
	}

	managedTasks := in.Config.Tasks
	if oldLock != nil {
		managedTasks = oldLock.Configuration.Tasks
	}

	newRoot, err := taskfile.UpdateRootTaskfile(rootBytes, taskfile.RootUpdateInput{
		Tasks:        in.Config.Tasks,
		TargetFolder: in.Config.TargetFolder,
		DestByTask:   in.DestByTask,
		ManagedTasks: managedTasks,
	})
	if err != nil {
		return nil, err
	}

	lock := buildLock(in, plannedFiles)
	meta := Metadata{
		TargetFolder:      in.Config.TargetFolder,
		LockFile:          in.Config.LockFilePath(),
		ConfigurationHash: in.Config.ConfigurationHash,
	}

	plan := &Plan{
		Requested:       in.Requested,
		Dependencies:    in.Dependencies,
		ManagedFiles:    plannedFiles,
		ModuleContents:  moduleContents,
		RootTaskfile:    newRoot,
		Lock:            lock,
		Metadata:        meta,
		OldLock:         oldLock,
		OldTargetFolder: oldTarget,
	}

	oldRootForDiff := rootBytes
	if !rootExisted {
		oldRootForDiff = nil
	}

	plan.Added, plan.Updated, plan.Removed, err = diffFiles(plan, workspace, oldRootForDiff, in.Config.MetadataPath(), mustMarshalMetadata(meta))
	if err != nil {
		return nil, err
	}
	plan.Changed = len(plan.Added) > 0 || len(plan.Updated) > 0 || len(plan.Removed) > 0
	plan.StagePaths = buildStagePaths(plan, in.Config.MetadataPath())
	return plan, nil
}

func planManagedFiles(in SyncInput, oldLock *LockFile) ([]ManagedFile, map[string]map[string]FileEntry, error) {
	var allModules []ModuleRecord
	for _, rec := range in.Requested {
		allModules = append(allModules, rec)
	}
	allModules = append(allModules, in.Dependencies...)
	sort.Slice(allModules, func(i, j int) bool {
		return allModules[i].SourceModule < allModules[j].SourceModule
	})

	moduleContents := make(map[string]map[string]FileEntry)
	var planned []ManagedFile

	for _, mod := range allModules {
		sourceDir := in.Snapshot.ModuleDir(mod.SourceModule)
		if _, err := os.Stat(sourceDir); err != nil {
			return nil, nil, &SyncError{Message: fmt.Sprintf("source module directory %q does not exist", mod.SourceModule)}
		}

		destDirRel := pathutil.JoinRelative(in.Config.TargetFolder, mod.DestinationModule)
		destDirAbs := pathutil.WorkspacePath(in.Config.Workspace, destDirRel)

		if err := validateDestination(destDirAbs, mod, oldLock); err != nil {
			return nil, nil, err
		}

		contents, err := collectModuleFiles(sourceDir, in.Config.IncludesDoc, in.SourceToDest)
		if err != nil {
			return nil, nil, err
		}
		moduleContents[mod.SourceModule] = contents

		for rel, entry := range contents {
			sum := sha256.Sum256(entry.Data)
			planned = append(planned, ManagedFile{
				SourceModule:      mod.SourceModule,
				DestinationModule: mod.DestinationModule,
				SourcePath:        pathutil.JoinRelative("taskfiles", mod.SourceModule, rel),
				Path:              pathutil.JoinRelative(destDirRel, rel),
				SHA256:            hex.EncodeToString(sum[:]),
			})
		}
	}

	sort.Slice(planned, func(i, j int) bool {
		if planned[i].Path == planned[j].Path {
			return planned[i].SourceModule < planned[j].SourceModule
		}
		return planned[i].Path < planned[j].Path
	})
	return planned, moduleContents, nil
}

func validateDestination(destDirAbs string, mod ModuleRecord, oldLock *LockFile) error {
	info, err := os.Stat(destDirAbs)
	if os.IsNotExist(err) {
		return nil
	}
	if err != nil {
		return err
	}
	if !info.IsDir() {
		return &SyncError{Message: fmt.Sprintf("destination %q exists and is not a directory", mod.Path)}
	}
	if oldLock == nil {
		return &SyncError{Message: fmt.Sprintf(`Cannot copy source module %q to %q: the destination exists but is not managed by TaskOtter.`, mod.SourceModule, mod.Path)}
	}
	for _, managed := range oldLock.ManagedFiles {
		if managed.DestinationModule == mod.DestinationModule {
			return nil
		}
	}
	return &SyncError{Message: fmt.Sprintf(`Cannot copy source module %q to %q: the destination exists but is not managed by TaskOtter.`, mod.SourceModule, mod.Path)}
}

func collectModuleFiles(sourceDir string, includesDoc bool, sourceToDest map[string]string) (map[string]FileEntry, error) {
	contents := make(map[string]FileEntry)
	err := filepath.WalkDir(sourceDir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		rel, err := filepath.Rel(sourceDir, path)
		if err != nil {
			return err
		}
		rel = filepath.ToSlash(rel)
		if !includesDoc && pathutil.IsDocPath(rel) {
			return nil
		}
		info, err := d.Info()
		if err != nil {
			return err
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		if rel == "Taskfile.yml" {
			data, err = taskfile.RewriteIncludes(data, sourceToDest)
			if err != nil {
				return err
			}
		}
		contents[rel] = FileEntry{Data: data, Mode: preserveMode(info.Mode())}
		return nil
	})
	return contents, err
}

func buildLock(in SyncInput, files []ManagedFile) LockFile {
	var lock LockFile
	lock.Source.Repository = config.StoreRepository
	lock.Source.RequestedVersion = in.Config.StoreVersion
	lock.Source.SourceRef = in.Snapshot.Ref.SourceRef
	lock.Source.ResolvedCommit = in.Snapshot.Ref.ResolvedCommit
	lock.Source.DefaultBranch = in.Snapshot.Ref.DefaultBranch
	lock.Configuration.TargetFolder = in.Config.TargetFolder
	lock.Configuration.Tasks = append([]string{}, in.Config.Tasks...)
	lock.Configuration.NodePackageManager = string(in.Config.NodePackageManager)
	lock.Configuration.NodeVersionManager = string(in.Config.NodeVersionManager)
	lock.Configuration.IncludesDoc = in.Config.IncludesDoc
	lock.ResolvedModules.Requested = orderedRequested(in.Requested)
	lock.ResolvedModules.Dependencies = append([]ModuleRecord{}, in.Dependencies...)
	lock.ManagedFiles = files
	return lock
}
