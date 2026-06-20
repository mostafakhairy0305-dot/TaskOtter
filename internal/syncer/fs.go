package syncer

import (
	"io"
	"os"
	"path/filepath"
	"sort"
)

// CopyFileToHook, when non-nil, replaces copyFileTo (for tests).
var CopyFileToHook func(path string, entry FileEntry) error

func writeFileAtomic(path string, data []byte, mode os.FileMode) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	tmp, err := os.CreateTemp(dir, ".taskotter-*")
	if err != nil {
		return err
	}
	tmpPath := tmp.Name()
	defer os.Remove(tmpPath)
	if _, err := tmp.Write(data); err != nil {
		tmp.Close()
		return err
	}
	if err := tmp.Chmod(mode); err != nil {
		tmp.Close()
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	return os.Rename(tmpPath, path)
}

func copyFileTo(path string, entry FileEntry) error {
	if CopyFileToHook != nil {
		return CopyFileToHook(path, entry)
	}
	return writeFileAtomic(path, entry.Data, entry.Mode)
}

func CopyFile(src, dst string, mode os.FileMode) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()
	data, err := io.ReadAll(in)
	if err != nil {
		return err
	}
	return writeFileAtomic(dst, data, mode)
}

func sortedModuleRecords(requested map[string]ModuleRecord, deps []ModuleRecord) []ModuleRecord {
	tasks := make([]string, 0, len(requested))
	for task := range requested {
		tasks = append(tasks, task)
	}
	sort.Strings(tasks)
	out := make([]ModuleRecord, 0, len(requested)+len(deps))
	for _, task := range tasks {
		out = append(out, requested[task])
	}
	return append(out, deps...)
}

func preserveMode(mode os.FileMode) os.FileMode {
	perm := mode.Perm()
	if perm&0o111 != 0 {
		return perm
	}
	return 0o644
}
