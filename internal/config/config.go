package config

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/mostafakhairy0305-dot/TaskOtter/internal/pathutil"
)

const (
	DefaultTargetFolder = "taskfiles"
	StoreRepository     = "mostafakhairy0305-dot/TaskOtter-store"
)

var unsafeStoreVersion = regexp.MustCompile(`(?i)(^refs/|\.\./|/|\\|\^|~|\^{commit})`)

type ValidationError struct {
	Field   string
	Message string
}

func (e *ValidationError) Error() string {
	if e.Field != "" {
		return fmt.Sprintf("%s: %s", e.Field, e.Message)
	}
	return e.Message
}

// actionInput reads a GitHub Actions input from the environment.
// Docker container actions expose INPUT_<NAME> with hyphens preserved
// (for example INPUT_GITHUB-TOKEN). Other runners may use underscores
// (for example INPUT_GITHUB_TOKEN).
func actionInput(name string) string {
	upper := strings.ToUpper(name)
	for _, key := range []string{
		"INPUT_" + upper,
		"INPUT_" + strings.ReplaceAll(upper, "-", "_"),
	} {
		if v := strings.TrimSpace(os.Getenv(key)); v != "" {
			return v
		}
	}
	return ""
}

func missingActionInput(name string) *ValidationError {
	upper := strings.ToUpper(name)
	return &ValidationError{
		Field: name,
		Message: fmt.Sprintf(
			"is required (set %q in the workflow step; checked env vars INPUT_%s, INPUT_%s)",
			name,
			upper,
			strings.ReplaceAll(upper, "-", "_"),
		),
	}
}

type PackageManager string

const (
	PMNPM  PackageManager = "npm"
	PMYarn PackageManager = "yarn"
	PMPnpm PackageManager = "pnpm"
	PMBun  PackageManager = "bun"
)

type VersionManager string

const (
	VMFnm VersionManager = "fnm"
	VMNvm VersionManager = "nvm"
)

type Config struct {
	Tasks              []string
	NodePackageManager PackageManager
	NodeVersionManager VersionManager
	IncludesDoc        bool
	StoreVersion       string
	TargetFolder       string
	GitHubToken        string
	Workspace          string
	Repository         string
	GitHubOutput       string
	ConfigurationHash  string
	BranchName         string
}

type hashPayload struct {
	Tasks              []string `json:"tasks"`
	NodePackageManager string   `json:"node_package_manager"`
	NodeVersionManager string   `json:"node_version_manager"`
	TargetFolder       string   `json:"target_folder"`
	StoreVersion       string   `json:"store_version"`
	IncludesDoc        bool     `json:"includes_doc"`
}

func LoadFromEnv() (*Config, error) {
	tasksRaw := actionInput("tasks")
	pmRaw := actionInput("node-package-manager")
	vmRaw := actionInput("node-version-manager")
	includesDocRaw := actionInput("includes-doc")
	storeVersion := actionInput("store-version")
	targetFolderRaw := actionInput("target-folder")
	token := actionInput("github-token")
	if token == "" {
		token = strings.TrimSpace(os.Getenv("GITHUB_TOKEN"))
	}
	workspace := os.Getenv("GITHUB_WORKSPACE")
	repository := os.Getenv("GITHUB_REPOSITORY")
	githubOutput := os.Getenv("GITHUB_OUTPUT")

	if workspace == "" {
		return nil, &ValidationError{Field: "GITHUB_WORKSPACE", Message: "is required"}
	}
	if token == "" {
		return nil, missingActionInput("github-token")
	}

	tasks, err := parseTasks(tasksRaw)
	if err != nil {
		return nil, err
	}

	pm, err := parsePackageManager(pmRaw)
	if err != nil {
		return nil, err
	}
	vm, err := parseVersionManager(vmRaw)
	if err != nil {
		return nil, err
	}

	if pm == PMBun && vm != "" {
		return nil, &ValidationError{
			Field:   "node-version-manager",
			Message: `node-package-manager "bun" cannot be combined with node-version-manager; leave node-version-manager empty`,
		}
	}

	includesDoc, err := parseIncludesDoc(includesDocRaw)
	if err != nil {
		return nil, err
	}

	if err := validateStoreVersion(storeVersion); err != nil {
		return nil, err
	}

	targetFolder := DefaultTargetFolder
	if targetFolderRaw != "" {
		targetFolder = targetFolderRaw
	}
	normalizedTarget, err := pathutil.ValidateTargetFolder(targetFolder, workspace)
	if err != nil {
		return nil, err
	}

	hash, branch := computeConfigurationHash(hashPayload{
		Tasks:              tasks,
		NodePackageManager: string(pm),
		NodeVersionManager: string(vm),
		TargetFolder:       normalizedTarget,
		StoreVersion:       storeVersion,
		IncludesDoc:        includesDoc,
	})

	return &Config{
		Tasks:              tasks,
		NodePackageManager: pm,
		NodeVersionManager: vm,
		IncludesDoc:        includesDoc,
		StoreVersion:       storeVersion,
		TargetFolder:       normalizedTarget,
		GitHubToken:        token,
		Workspace:          workspace,
		Repository:         repository,
		GitHubOutput:       githubOutput,
		ConfigurationHash:  hash,
		BranchName:         branch,
	}, nil
}

func parseTasks(raw string) ([]string, error) {
	raw = strings.ReplaceAll(raw, ",", "\n")
	lines := strings.Split(raw, "\n")
	seen := make(map[string]struct{})
	var tasks []string
	for _, line := range lines {
		name := strings.TrimSpace(line)
		if name == "" {
			continue
		}
		if err := pathutil.ValidateTaskName(name); err != nil {
			return nil, err
		}
		if _, ok := seen[name]; ok {
			continue
		}
		seen[name] = struct{}{}
		tasks = append(tasks, name)
	}
	if len(tasks) == 0 {
		return nil, &ValidationError{Field: "tasks", Message: "at least one task is required"}
	}
	return tasks, nil
}

func parsePackageManager(raw string) (PackageManager, error) {
	switch raw {
	case "":
		return "", nil
	case "npm", "yarn", "pnpm", "bun":
		return PackageManager(raw), nil
	default:
		return "", &ValidationError{
			Field:   "node-package-manager",
			Message: fmt.Sprintf("invalid value %q: allowed values are npm, yarn, pnpm, bun, or empty", raw),
		}
	}
}

func parseVersionManager(raw string) (VersionManager, error) {
	switch raw {
	case "":
		return "", nil
	case "fnm", "nvm":
		return VersionManager(raw), nil
	default:
		return "", &ValidationError{
			Field:   "node-version-manager",
			Message: fmt.Sprintf("invalid value %q: allowed values are fnm, nvm, or empty", raw),
		}
	}
}

func parseIncludesDoc(raw string) (bool, error) {
	if raw == "" {
		return true, nil
	}
	switch strings.ToLower(raw) {
	case "true":
		return true, nil
	case "false":
		return false, nil
	default:
		return false, &ValidationError{
			Field:   "includes-doc",
			Message: fmt.Sprintf("invalid value %q: allowed values are true or false", raw),
		}
	}
}

func validateStoreVersion(version string) error {
	if version == "" {
		return nil
	}
	if unsafeStoreVersion.MatchString(version) {
		return &ValidationError{
			Field:   "store-version",
			Message: fmt.Sprintf("unsafe revision expression %q", version),
		}
	}
	return nil
}

func computeConfigurationHash(payload hashPayload) (string, string) {
	data, _ := json.Marshal(payload)
	sum := sha256.Sum256(data)
	hash := hex.EncodeToString(sum[:])
	return hash, fmt.Sprintf("taskotter/sync-%s", hash[:12])
}

func (c *Config) LockFilePath() string {
	return pathutil.JoinRelative(c.TargetFolder, ".taskotter-lock.yml")
}

func (c *Config) MetadataPath() string {
	return ".taskotter/metadata.yml"
}
