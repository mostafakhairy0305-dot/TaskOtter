package syncer

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"github.com/mostafakhairy0305-dot/TaskOtter/internal/pathutil"
	"gopkg.in/yaml.v3"
)

type stagedFile struct {
	finalRel string
	entry    FileEntry
}

func buildStagedFiles(plan *Plan, syncInput SyncInput) ([]stagedFile, error) {
	var staged []stagedFile

	for _, mod := range sortedModuleRecords(plan.Requested, plan.Dependencies) {
		contents := plan.ModuleContents[mod.SourceModule]

		rels := make([]string, 0, len(contents))
		for rel := range contents {
			rels = append(rels, rel)
		}

		sort.Strings(rels)

		destDirRel := pathutil.JoinRelative(syncInput.Config.TargetFolder, mod.DestinationModule)
		for _, rel := range rels {
			finalRel := pathutil.JoinRelative(destDirRel, rel)
			staged = append(staged, stagedFile{finalRel: finalRel, entry: contents[rel]})
		}
	}

	if syncInput.Config.SyncRoot {
		staged = append(staged, stagedFile{
			finalRel: rootTaskfileName,
			entry:    FileEntry{Data: plan.RootTaskfile, Mode: fileModeRegular},
		})
	}

	lockBytes, err := yaml.Marshal(plan.Lock)
	if err != nil {
		return nil, fmt.Errorf("marshal lock file: %w", err)
	}

	staged = append(staged, stagedFile{
		finalRel: plan.Metadata.LockFile,
		entry:    FileEntry{Data: lockBytes, Mode: fileModeRegular},
	})

	metaBytes, err := yaml.Marshal(plan.Metadata)
	if err != nil {
		return nil, fmt.Errorf("marshal metadata: %w", err)
	}

	staged = append(staged, stagedFile{
		finalRel: syncInput.Config.MetadataPath(),
		entry:    FileEntry{Data: metaBytes, Mode: fileModeRegular},
	})

	return staged, nil
}

// ApplyPlan writes planned files atomically and removes obsolete managed paths.
func ApplyPlan(plan *Plan, syncInput SyncInput) error {
	workspace := syncInput.Config.Workspace

	stagingParent := pathutil.WorkspacePath(workspace, ".taskotter/staging")

	err := os.MkdirAll(stagingParent, dirModePerm)
	if err != nil {
		return fmt.Errorf("create staging directory: %w", err)
	}

	stagingRoot, err := os.MkdirTemp(stagingParent, "apply-*")
	if err != nil {
		return fmt.Errorf("create staging directory: %w", err)
	}

	defer func() { _ = os.RemoveAll(stagingRoot) }()

	staged, err := buildStagedFiles(plan, syncInput)
	if err != nil {
		return err
	}

	for _, stagedEntry := range staged {
		stagePath := filepath.Join(stagingRoot, filepath.FromSlash(stagedEntry.finalRel))

		err = copyFileTo(stagePath, stagedEntry.entry)
		if err != nil {
			return fmt.Errorf("stage %q: %w", stagedEntry.finalRel, err)
		}
	}

	err = validateGeneratedYAML(staged)
	if err != nil {
		return err
	}

	for _, stagedEntry := range staged {
		finalPath := pathutil.WorkspacePath(workspace, stagedEntry.finalRel)

		err = os.MkdirAll(filepath.Dir(finalPath), dirModePerm)
		if err != nil {
			return fmt.Errorf("prepare %q: %w", stagedEntry.finalRel, err)
		}

		err = copyFileTo(finalPath, stagedEntry.entry)
		if err != nil {
			return fmt.Errorf("write %q: %w", stagedEntry.finalRel, err)
		}
	}

	return removeObsolete(plan, workspace)
}

func validateGeneratedYAML(staged []stagedFile) error {
	for _, stagedEntry := range staged {
		switch {
		case stagedEntry.finalRel == rootTaskfileName:
			var node yaml.Node

			err := yaml.Unmarshal(stagedEntry.entry.Data, &node)
			if err != nil {
				return fmt.Errorf("validate root Taskfile.yml: %w", err)
			}
		case filepath.Base(stagedEntry.finalRel) == ".taskotter-lock.yml":
			var lock LockFile

			err := yaml.Unmarshal(stagedEntry.entry.Data, &lock)
			if err != nil {
				return fmt.Errorf("validate lock file: %w", err)
			}
		case filepath.Base(stagedEntry.finalRel) == "metadata.yml":
			var meta Metadata

			err := yaml.Unmarshal(stagedEntry.entry.Data, &meta)
			if err != nil {
				return fmt.Errorf("validate metadata: %w", err)
			}
		}
	}

	return nil
}

func removeObsolete(plan *Plan, workspace string) error {
	currentPaths := make(map[string]struct{}, len(plan.ManagedFiles))
	for _, managed := range plan.ManagedFiles {
		currentPaths[managed.Path] = struct{}{}
	}

	if plan.OldLock == nil {
		return nil
	}

	for _, old := range plan.OldLock.ManagedFiles {
		if _, ok := currentPaths[old.Path]; ok {
			continue
		}

		abs := pathutil.WorkspacePath(workspace, old.Path)

		err := os.Remove(abs)
		if err != nil && !os.IsNotExist(err) {
			return fmt.Errorf("remove obsolete file %q: %w", old.Path, err)
		}
	}

	return cleanupOldTarget(plan, workspace)
}

func cleanupOldTarget(plan *Plan, workspace string) error {
	if plan.OldLock == nil || plan.OldTargetFolder == "" || plan.OldTargetFolder == plan.Metadata.TargetFolder {
		return nil
	}

	for _, old := range plan.OldLock.ManagedFiles {
		if !pathutil.HasFolderPrefix(old.Path, plan.OldTargetFolder) {
			continue
		}

		abs := pathutil.WorkspacePath(workspace, old.Path)

		err := os.Remove(abs)
		if err != nil && !os.IsNotExist(err) {
			return fmt.Errorf("remove old target file %q: %w", old.Path, err)
		}
	}

	oldLock := pathutil.WorkspacePath(workspace, pathutil.JoinRelative(plan.OldTargetFolder, ".taskotter-lock.yml"))

	err := os.Remove(oldLock)
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("remove old lock file: %w", err)
	}

	return nil
}
