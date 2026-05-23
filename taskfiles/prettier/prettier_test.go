package prettier_test

import (
	"testing"

	"github.com/mostafakhairy0305-dot/TaskOtter/taskfiles/internal/tasktest"
)

var publicTasks = []string{
	"check",
	"ci",
	"config:init",
	"fix",
	"help",
	"install",
	"version",
	"write",
}

var publicVars = []string{
	"CONFIG",
	"EXTRA_ARGS",
	"IGNORE_PATH",
	"PM",
	"TARGETS",
}

func TestTaskfileModuleContract(t *testing.T) {
	tasktest.AssertModule(t, "prettier", publicTasks, publicVars)
}

func TestRepresentativeDryRuns(t *testing.T) {
	tasktest.AssertDryRunContains(t, "prettier",
		[]string{"write", "PM=bun", "--", "--ignore-unknown"},
		"bunx prettier",
		". --write",
		"--ignore-unknown",
	)

	tasktest.AssertDryRunContains(t, "prettier",
		[]string{"check", "PM=pnpm", "TARGETS=src/**/*.ts", "CONFIG=.prettierrc.json"},
		"pnpm exec prettier",
		"src/**/*.ts --check",
		"--config \".prettierrc.json\"",
	)
}
