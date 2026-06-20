package syncer_test

import (
	"github.com/mostafakhairy0305-dot/TaskOtter/internal/config"
)

const testModuleEslint = "eslint"

func testConfig(workspace string, mutate func(*config.Config)) *config.Config {
	cfg := &config.Config{
		Tasks:              nil,
		JSRuntime:          "",
		NodePackageManager: "",
		NodeVersionManager: "",
		IncludesDoc:        false,
		FailOnChanges:      false,
		StoreVersion:       "",
		TargetFolder:       config.DefaultTargetFolder,
		GitHubToken:        "",
		Workspace:          workspace,
		Repository:         "",
		GitHubOutput:       "",
		ConfigurationHash:  "",
		BranchName:         "",
	}
	if mutate != nil {
		mutate(cfg)
	}

	return cfg
}
