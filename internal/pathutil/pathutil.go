package pathutil

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

var (
	windowsAbsPath = regexp.MustCompile(`^[A-Za-z]:[\\/]`)
	taskNameRe     = regexp.MustCompile(`^[a-z0-9][a-z0-9-]*$`)
)

type PathError struct {
	Field   string
	Value   string
	Message string
}

func (e *PathError) Error() string {
	if e.Field != "" {
		return fmt.Sprintf("%s: %s", e.Field, e.Message)
	}
	return e.Message
}

func NormalizeSlashes(path string) string {
	path = strings.ReplaceAll(path, "\\", "/")
	path = strings.TrimSpace(path)
	for strings.Contains(path, "//") {
		path = strings.ReplaceAll(path, "//", "/")
	}
	path = strings.Trim(path, "/")
	return path
}

func ValidateTaskName(name string) error {
	if name == "" {
		return &PathError{Field: "tasks", Message: "task name must not be empty"}
	}
	if strings.ContainsAny(name, "/\\") || strings.Contains(name, "..") {
		return &PathError{Field: "tasks", Value: name, Message: fmt.Sprintf("unsafe task name %q", name)}
	}
	if !taskNameRe.MatchString(name) {
		return &PathError{Field: "tasks", Value: name, Message: fmt.Sprintf("invalid task name %q: must match ^[a-z0-9][a-z0-9-]*$", name)}
	}
	return nil
}

func ValidateTargetFolder(raw, workspace string) (string, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return "", &PathError{Field: "target-folder", Message: "must not be empty"}
	}

	if filepath.IsAbs(raw) || strings.HasPrefix(raw, "/") || windowsAbsPath.MatchString(raw) {
		return "", &PathError{Field: "target-folder", Value: raw, Message: "must be a relative path"}
	}

	normalized := NormalizeSlashes(raw)
	if normalized == "" {
		return "", &PathError{Field: "target-folder", Value: raw, Message: "must not be empty after normalization"}
	}

	for _, part := range strings.Split(normalized, "/") {
		if part == ".." {
			return "", &PathError{Field: "target-folder", Value: raw, Message: "must not contain .. path components"}
		}
	}

	if normalized == ".git" || strings.HasPrefix(normalized, ".git/") {
		return "", &PathError{Field: "target-folder", Value: raw, Message: "must not point to .git"}
	}
	if normalized == ".github/actions" || strings.HasPrefix(normalized, ".github/actions/") {
		return "", &PathError{Field: "target-folder", Value: raw, Message: "must not point inside .github/actions"}
	}

	absWorkspace, err := filepath.Abs(workspace)
	if err != nil {
		return "", fmt.Errorf("resolve workspace: %w", err)
	}
	evalWorkspace, err := filepath.EvalSymlinks(absWorkspace)
	if err != nil {
		evalWorkspace = absWorkspace
	}

	targetAbs := filepath.Join(evalWorkspace, filepath.FromSlash(normalized))
	rel, err := filepath.Rel(evalWorkspace, filepath.Clean(targetAbs))
	if err != nil {
		return "", &PathError{Field: "target-folder", Value: raw, Message: err.Error()}
	}
	if rel == ".." || strings.HasPrefix(rel, ".."+string(os.PathSeparator)) {
		return "", &PathError{Field: "target-folder", Value: raw, Message: "path resolves outside workspace"}
	}

	if err := validatePathComponents(evalWorkspace, normalized, raw); err != nil {
		return "", err
	}

	return normalized, nil
}

func validatePathComponents(evalWorkspace, normalized, raw string) error {
	current := evalWorkspace
	for _, part := range strings.Split(normalized, "/") {
		current = filepath.Join(current, part)
		info, err := os.Lstat(current)
		if err != nil {
			if os.IsNotExist(err) {
				continue
			}
			return &PathError{Field: "target-folder", Value: raw, Message: err.Error()}
		}
		if info.Mode()&os.ModeSymlink != 0 {
			linkTarget, err := filepath.EvalSymlinks(current)
			if err != nil {
				return &PathError{Field: "target-folder", Value: raw, Message: "invalid symlink target"}
			}
			linkRel, err := filepath.Rel(evalWorkspace, linkTarget)
			if err != nil || linkRel == ".." || strings.HasPrefix(linkRel, ".."+string(os.PathSeparator)) {
				return &PathError{Field: "target-folder", Value: raw, Message: "must not escape through symlinks"}
			}
			current = linkTarget
		}
	}
	return nil
}

func JoinRelative(base string, parts ...string) string {
	all := append([]string{base}, parts...)
	return NormalizeSlashes(filepath.ToSlash(filepath.Join(toOSParts(all)...)))
}

func toOSParts(parts []string) []string {
	out := make([]string, len(parts))
	for i, p := range parts {
		out[i] = filepath.FromSlash(p)
	}
	return out
}

func WorkspacePath(workspace, rel string) string {
	return filepath.Join(workspace, filepath.FromSlash(rel))
}

func IsDocPath(rel string) bool {
	rel = NormalizeSlashes(rel)
	if rel == "README.md" {
		return true
	}
	return strings.HasPrefix(rel, "docs/") || strings.Contains(rel, "/docs/")
}

func HasFolderPrefix(path, folder string) bool {
	path = NormalizeSlashes(path)
	folder = NormalizeSlashes(folder)
	if path == folder {
		return true
	}
	return strings.HasPrefix(path, folder+"/")
}
