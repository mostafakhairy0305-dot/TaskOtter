package stylelint_test

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
	"install",
	"version",
}

var publicVars = []string{
	"ALLOW_EMPTY_INPUT",
	"CACHE",
	"CONFIG",
	"EXTRA_ARGS",
	"PM",
	"TARGETS",
}

func TestTaskfileModuleContract(t *testing.T) {
	tasktest.AssertModule(t, "stylelint", publicTasks, publicVars)
}

func TestRepresentativeDryRuns(t *testing.T) {
	tasktest.AssertDryRunContains(t, "stylelint",
		[]string{"fix", "PM=yarn", "TARGETS=src/**/*.scss", "--", "--formatter", "verbose"},
		"js:yarn:exec",
		"src/**/*.scss --fix",
		"--formatter verbose",
	)

	tasktest.AssertDryRunContains(t, "stylelint",
		[]string{"ci", "PM=pnpm", "CACHE=false", "ALLOW_EMPTY_INPUT=false"},
		"js:pnpm:exec",
		"--max-warnings=0",
	)
}
