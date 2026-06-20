package syncer

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"github.com/mostafakhairy0305-dot/taskotter-sync-action/internal/pathutil"
	"gopkg.in/yaml.v3"
)

type stagedFile struct {
	finalRel string
	entry    FileEntry
}

func buildStagedFiles(plan *Plan, in SyncInput) ([]stagedFile, error) {
	var staged []stagedFile

	for _, mod := range sortedModuleRecords(plan.Requested, plan.Dependencies) {
		contents := plan.ModuleContents[mod.SourceModule]
		rels := make([]string, 0, len(contents))
		for rel := range contents {
			rels = append(rels, rel)
		}
		sort.Strings(rels)
		destDirRel := pathutil.JoinRelative(in.Config.TargetFolder, mod.DestinationModule)
		for _, rel := range rels {
			finalRel := pathutil.JoinRelative(destDirRel, rel)
			staged = append(staged, stagedFile{finalRel: finalRel, entry: contents[rel]})
		}
	}

	staged = append(staged, stagedFile{finalRel: "Taskfile.yml", entry: FileEntry{Data: plan.RootTaskfile, Mode: 0o644}})

	lockBytes, err := yaml.Marshal(plan.Lock)
	if err != nil {
		return nil, fmt.Errorf("marshal lock file: %w", err)
	}
	staged = append(staged, stagedFile{finalRel: plan.Metadata.LockFile, entry: FileEntry{Data: lockBytes, Mode: 0o644}})

	metaBytes, err := yaml.Marshal(plan.Metadata)
	if err != nil {
		return nil, fmt.Errorf("marshal metadata: %w", err)
	}
	staged = append(staged, stagedFile{finalRel: in.Config.MetadataPath(), entry: FileEntry{Data: metaBytes, Mode: 0o644}})

	return staged, nil
}

func ApplyPlan(plan *Plan, in SyncInput) error {
	workspace := in.Config.Workspace
	stagingParent := pathutil.WorkspacePath(workspace, ".taskotter/staging")
	if err := os.MkdirAll(stagingParent, 0o755); err != nil {
		return fmt.Errorf("create staging directory: %w", err)
	}
	stagingRoot, err := os.MkdirTemp(stagingParent, "apply-*")
	if err != nil {
		return fmt.Errorf("create staging directory: %w", err)
	}
	defer func() { _ = os.RemoveAll(stagingRoot) }()

	staged, err := buildStagedFiles(plan, in)
	if err != nil {
		return err
	}

	for _, sf := range staged {
		stagePath := filepath.Join(stagingRoot, filepath.FromSlash(sf.finalRel))
		if err := copyFileTo(stagePath, sf.entry); err != nil {
			return fmt.Errorf("stage %q: %w", sf.finalRel, err)
		}
	}

	if err := validateGeneratedYAML(staged); err != nil {
		return err
	}

	for _, sf := range staged {
		finalPath := pathutil.WorkspacePath(workspace, sf.finalRel)
		if err := os.MkdirAll(filepath.Dir(finalPath), 0o755); err != nil {
			return fmt.Errorf("prepare %q: %w", sf.finalRel, err)
		}
		if err := copyFileTo(finalPath, sf.entry); err != nil {
			return fmt.Errorf("write %q: %w", sf.finalRel, err)
		}
	}

	return removeObsolete(plan, workspace)
}

func validateGeneratedYAML(staged []stagedFile) error {
	for _, sf := range staged {
		switch {
		case sf.finalRel == "Taskfile.yml":
			var node yaml.Node
			if err := yaml.Unmarshal(sf.entry.Data, &node); err != nil {
				return fmt.Errorf("validate root Taskfile.yml: %w", err)
			}
		case filepath.Base(sf.finalRel) == ".taskotter-lock.yml":
			var lock LockFile
			if err := yaml.Unmarshal(sf.entry.Data, &lock); err != nil {
				return fmt.Errorf("validate lock file: %w", err)
			}
		case filepath.Base(sf.finalRel) == "metadata.yml":
			var meta Metadata
			if err := yaml.Unmarshal(sf.entry.Data, &meta); err != nil {
				return fmt.Errorf("validate metadata: %w", err)
			}
		}
	}
	return nil
}

func removeObsolete(plan *Plan, workspace string) error {
	currentPaths := make(map[string]struct{}, len(plan.ManagedFiles))
	for _, mf := range plan.ManagedFiles {
		currentPaths[mf.Path] = struct{}{}
	}
	if plan.OldLock == nil {
		return nil
	}
	for _, old := range plan.OldLock.ManagedFiles {
		if _, ok := currentPaths[old.Path]; ok {
			continue
		}
		abs := pathutil.WorkspacePath(workspace, old.Path)
		if err := os.Remove(abs); err != nil && !os.IsNotExist(err) {
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
		if err := os.Remove(abs); err != nil && !os.IsNotExist(err) {
			return fmt.Errorf("remove old target file %q: %w", old.Path, err)
		}
	}
	oldLock := pathutil.WorkspacePath(workspace, pathutil.JoinRelative(plan.OldTargetFolder, ".taskotter-lock.yml"))
	if err := os.Remove(oldLock); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("remove old lock file: %w", err)
	}
	return nil
}
