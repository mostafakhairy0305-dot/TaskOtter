package eslint_test

import (
	"testing"

	"github.com/mostafakhairy0305-dot/TaskOtter/internal/tasktest"
)

var publicTasks = []string{
	"cache:clean",
	"check",
	"ci",
	"config:init",
	"fix",
	"help",
	"init",
	"install",
	"version",
}

var publicVars = []string{
	"CACHE",
	"CONFIG",
	"EXTRA_ARGS",
	"PM",
	"TARGETS",
}

func TestTaskfileModuleContract(t *testing.T) {
	tasktest.AssertModule(t, "eslint", publicTasks, publicVars)
}

func TestRepresentativeDryRuns(t *testing.T) {
	tasktest.AssertDryRunContains(t, "eslint",
		[]string{"check", "PM=pnpm", "TARGETS=src test", "--", "--quiet"},
		"js:pnpm:exec",
		"--cache --cache-location .cache/eslint/",
		"src test",
		"--quiet",
	)

	tasktest.AssertDryRunContains(t, "eslint",
		[]string{"ci", "PM=bun", "CONFIG=eslint.config.js", "CACHE=false"},
		"js:bun:exec",
		"eslint.config.js",
		"--max-warnings=0",
	)

	tasktest.AssertDryRunContains(t, "eslint",
		[]string{"config:init", "PM=npm"},
		"js:npm:exec",
		"--init",
	)
}
