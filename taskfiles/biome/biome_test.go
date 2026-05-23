package biome_test

import (
	"testing"

	"github.com/mostafakhairy0305-dot/TaskOtter/taskfiles/internal/tasktest"
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
		"yarn exec biome",
		"ci --config-path \"biome.json\" src",
	)

	tasktest.AssertDryRunContains(t, "biome",
		[]string{"format:write", "PM=pnpm", "--", "--no-errors-on-unmatched"},
		"pnpm exec biome",
		"format --write",
		"--no-errors-on-unmatched",
	)
}
