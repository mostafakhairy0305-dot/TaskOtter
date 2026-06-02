package biome_test

import (
	"testing"

	"github.com/mostafakhairy0305-dot/TaskOtter/internal/tasktest"
)

var publicTasks = []string{
	"cache:clean",
	"check",
	"check:write",
	"ci",
	"config:init",
	"fix",
	"format",
	"format:write",
	"help",
	"init",
	"install",
	"lint",
	"lint:fix",
	"version",
}

var publicVars = []string{
	"CONFIG",
	"EXTRA_ARGS",
	"PM",
	"TARGETS",
}

func TestTaskfileModuleContract(t *testing.T) {
	tasktest.AssertModule(t, "biome", publicTasks, publicVars)
}

func TestRepresentativeDryRuns(t *testing.T) {
	tasktest.AssertDryRunContains(t, "biome",
		[]string{"ci", "PM=yarn", "CONFIG=biome.json", "TARGETS=src"},
		"js:yarn:exec",
		"biome.json",
		"ci",
		"src",
	)

	tasktest.AssertDryRunContains(t, "biome",
		[]string{"format:write", "PM=pnpm", "--", "--no-errors-on-unmatched"},
		"js:pnpm:exec",
		"format --write",
		"--no-errors-on-unmatched",
	)

	tasktest.AssertDryRunContains(t, "biome",
		[]string{"config:init", "PM=npm"},
		"js:npm:exec",
		"init",
	)
}
