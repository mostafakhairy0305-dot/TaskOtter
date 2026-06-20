package syncer

import (
	"os"

	"github.com/mostafakhairy0305-dot/TaskOtter/internal/config"
	"github.com/mostafakhairy0305-dot/TaskOtter/internal/store"
)

type ModuleRecord struct {
	SourceModule      string `yaml:"source_module"`
	DestinationModule string `yaml:"destination_module"`
	Path              string `yaml:"path,omitempty"`
}

type ManagedFile struct {
	SourceModule      string `yaml:"source_module"`
	DestinationModule string `yaml:"destination_module"`
	SourcePath        string `yaml:"source_path"`
	Path              string `yaml:"path"`
	SHA256            string `yaml:"sha256"`
}

type LockFile struct {
	Source struct {
		Repository       string `yaml:"repository"`
		RequestedVersion string `yaml:"requested_version"`
		SourceRef        string `yaml:"source_ref"`
		ResolvedCommit   string `yaml:"resolved_commit"`
		DefaultBranch    string `yaml:"default_branch"`
	} `yaml:"source"`
	Configuration struct {
		TargetFolder       string   `yaml:"target_folder"`
		Tasks              []string `yaml:"tasks"`
		NodePackageManager string   `yaml:"node_package_manager"`
		NodeVersionManager string   `yaml:"node_version_manager"`
		IncludesDoc        bool     `yaml:"includes_doc"`
	} `yaml:"configuration"`
	ResolvedModules struct {
		Requested    orderedRequested `yaml:"requested"`
		Dependencies []ModuleRecord   `yaml:"dependencies"`
	} `yaml:"resolved_modules"`
	ManagedFiles []ManagedFile `yaml:"managed_files"`
}

type Metadata struct {
	TargetFolder      string `yaml:"target_folder"`
	LockFile          string `yaml:"lock_file"`
	ConfigurationHash string `yaml:"configuration_hash"`
}

type FileEntry struct {
	Data []byte
	Mode os.FileMode
}

type Plan struct {
	Requested       map[string]ModuleRecord
	Dependencies    []ModuleRecord
	ManagedFiles    []ManagedFile
	ModuleContents  map[string]map[string]FileEntry
	RootTaskfile    []byte
	Lock            LockFile
	Metadata        Metadata
	Added           []string
	Updated         []string
	Removed         []string
	Changed         bool
	OldLock         *LockFile
	OldTargetFolder string
	StagePaths      []string
}

type SyncInput struct {
	Config       *config.Config
	Snapshot     *store.Snapshot
	Requested    map[string]ModuleRecord
	Dependencies []ModuleRecord
	SourceToDest map[string]string
	DestByTask   map[string]string
}

type SyncError struct {
	Message string
}

func (e *SyncError) Error() string {
	return e.Message
}
