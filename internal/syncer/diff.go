package syncer

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"os"
	"sort"

	"github.com/mostafakhairy0305-dot/TaskOtter/internal/pathutil"
	"gopkg.in/yaml.v3"
)

func lockContentChanged(old *LockFile, newLock *LockFile) (bool, error) {
	oldNorm, err := marshalLockForCompare(old)
	if err != nil {
		return false, err
	}
	newNorm, err := marshalLockForCompare(newLock)
	if err != nil {
		return false, err
	}
	return !bytes.Equal(oldNorm, newNorm), nil
}

func marshalLockForCompare(lock *LockFile) ([]byte, error) {
	if lock == nil {
		return nil, nil
	}
	cloned := *lock
	cloned.Source.ResolvedCommit = ""
	return yaml.Marshal(cloned)
}

func diffFiles(plan *Plan, workspace string, oldRoot []byte, metadataPath string, plannedMeta []byte) (added, updated, removed []string, err error) {
	current := make(map[string]ManagedFile, len(plan.ManagedFiles))
	for _, mf := range plan.ManagedFiles {
		current[mf.Path] = mf
	}

	if plan.OldLock != nil {
		for _, mf := range plan.OldLock.ManagedFiles {
			if _, ok := current[mf.Path]; !ok {
				removed = append(removed, mf.Path)
			}
		}
	}

	for path, mf := range current {
		abs := pathutil.WorkspacePath(workspace, path)
		data, readErr := os.ReadFile(abs)
		if readErr != nil {
			if os.IsNotExist(readErr) {
				added = append(added, path)
				continue
			}
			return nil, nil, nil, readErr
		}
		sum := sha256.Sum256(data)
		if hex.EncodeToString(sum[:]) != mf.SHA256 {
			updated = append(updated, path)
		}
	}

	if !bytes.Equal(oldRoot, plan.RootTaskfile) {
		if len(oldRoot) == 0 {
			added = append(added, "Taskfile.yml")
		} else {
			updated = append(updated, "Taskfile.yml")
		}
	}

	lockPath := plan.Metadata.LockFile
	lockAbs := pathutil.WorkspacePath(workspace, lockPath)
	oldLockBytes, readErr := os.ReadFile(lockAbs)
	if readErr != nil && !os.IsNotExist(readErr) {
		return nil, nil, nil, readErr
	}

	changed, err := lockContentChanged(plan.OldLock, &plan.Lock)
	if err != nil {
		return nil, nil, nil, err
	}
	if changed {
		if len(oldLockBytes) == 0 {
			added = append(added, lockPath)
		} else {
			updated = append(updated, lockPath)
		}
	}

	metaAbs := pathutil.WorkspacePath(workspace, metadataPath)
	oldMetaBytes, readErr := os.ReadFile(metaAbs)
	if readErr != nil && !os.IsNotExist(readErr) {
		return nil, nil, nil, readErr
	}
	if !bytes.Equal(oldMetaBytes, plannedMeta) {
		if len(oldMetaBytes) == 0 {
			added = append(added, metadataPath)
		} else {
			updated = append(updated, metadataPath)
		}
	}

	sort.Strings(added)
	sort.Strings(updated)
	sort.Strings(removed)
	return added, updated, removed, nil
}

func buildStagePaths(plan *Plan, metadataPath string) []string {
	paths := make(map[string]struct{})
	for _, mf := range plan.ManagedFiles {
		paths[mf.Path] = struct{}{}
	}
	for _, rm := range plan.Removed {
		paths[rm] = struct{}{}
	}
	paths["Taskfile.yml"] = struct{}{}
	paths[plan.Metadata.LockFile] = struct{}{}
	paths[metadataPath] = struct{}{}

	if plan.OldLock != nil && plan.OldTargetFolder != "" && plan.OldTargetFolder != plan.Metadata.TargetFolder {
		for _, mf := range plan.OldLock.ManagedFiles {
			if pathutil.HasFolderPrefix(mf.Path, plan.OldTargetFolder) {
				paths[mf.Path] = struct{}{}
			}
		}
		oldLockPath := pathutil.JoinRelative(plan.OldTargetFolder, ".taskotter-lock.yml")
		paths[oldLockPath] = struct{}{}
	}

	out := make([]string, 0, len(paths))
	for p := range paths {
		out = append(out, p)
	}
	sort.Strings(out)
	return out
}
