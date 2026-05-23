package depcheck_test

import (
	"testing"

	"github.com/mostafakhairy0305-dot/TaskOtter/taskfiles/internal/tasktest"
)

var publicTasks = []string{
	"check",
	"ci",
	"help",
	"ignores",
	"install",
	"json",
	"skip-missing",
	"version",
}

var publicVars = []string{
	"EXTRA_ARGS",
	"IGNORE_PACKAGES",
	"PM",
	"PROJECT_PATH",
	"TARGETS",
}

func TestTaskfileModuleContract(t *testing.T) {
	tasktest.AssertModule(t, "depcheck", publicTasks, publicVars)
}

func TestRepresentativeDryRuns(t *testing.T) {
	tasktest.AssertDryRunContains(t, "depcheck",
		[]string{"check", "PM=npm", "PROJECT_PATH=packages/app", "--", "--ignores=@types/*,eslint-*"},
		"npx --no-install depcheck",
		"packages/app",
		"--ignores=@types/*,eslint-*",
	)

	tasktest.AssertDryRunContains(t, "depcheck",
		[]string{"ignores", "PM=pnpm", "IGNORE_PACKAGES=@types/*,eslint-*"},
		"pnpm exec depcheck",
		"--ignores=\"@types/*,eslint-*\"",
	)
}
