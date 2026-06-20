package syncer

import (
	"fmt"
	"os"

	"github.com/mostafakhairy0305-dot/taskotter-sync-action/internal/config"
	"github.com/mostafakhairy0305-dot/taskotter-sync-action/internal/pathutil"
	"gopkg.in/yaml.v3"
)

func LoadMetadata(path string) (*Metadata, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var meta Metadata
	if err := yaml.Unmarshal(data, &meta); err != nil {
		return nil, fmt.Errorf("parse metadata %q: %w", path, err)
	}
	return &meta, nil
}

func LoadLock(path string) (*LockFile, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var lock LockFile
	if err := yaml.Unmarshal(data, &lock); err != nil {
		return nil, fmt.Errorf("parse lock file %q: %w", path, err)
	}
	return &lock, nil
}

func loadPreviousState(workspace string, cfg *config.Config) (*LockFile, *Metadata, string, error) {
	metaPath := pathutil.WorkspacePath(workspace, cfg.MetadataPath())
	oldMeta, err := LoadMetadata(metaPath)
	if err != nil && !os.IsNotExist(err) {
		return nil, nil, "", err
	}

	oldLockPath := cfg.LockFilePath()
	oldTarget := ""
	if oldMeta != nil {
		if oldMeta.LockFile != "" {
			oldLockPath = oldMeta.LockFile
		}
		oldTarget = oldMeta.TargetFolder
	}

	lockPath := pathutil.WorkspacePath(workspace, oldLockPath)
	oldLock, err := LoadLock(lockPath)
	if err != nil && !os.IsNotExist(err) {
		return nil, nil, "", err
	}
	if oldLock != nil && oldTarget == "" {
		oldTarget = oldLock.Configuration.TargetFolder
	}
	return oldLock, oldMeta, oldTarget, nil
}
