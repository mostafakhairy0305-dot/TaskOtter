package knip_test

import (
	"testing"

	"github.com/mostafakhairy0305-dot/TaskOtter/taskfiles/internal/tasktest"
)

var publicTasks = []string{
	"check",
	"ci",
	"config:init",
	"dependencies",
	"dev-dependencies",
	"exports",
	"files",
	"fix",
	"help",
	"init",
	"install",
	"production",
	"version",
}

var publicVars = []string{
	"CONFIG",
	"EXTRA_ARGS",
	"PM",
}

func TestTaskfileModuleContract(t *testing.T) {
	tasktest.AssertModule(t, "knip", publicTasks, publicVars)
}

func TestRepresentativeDryRuns(t *testing.T) {
	tasktest.AssertDryRunContains(t, "knip",
		[]string{"check", "PM=pnpm", "--", "--debug"},
		`pnpm:exec BINARY="knip"`,
		"--debug",
	)

	tasktest.AssertDryRunContains(t, "knip",
		[]string{"production", "PM=bun", "CONFIG=knip.json"},
		`bun:exec BINARY="knip"`,
		"--production",
		"knip.json",
	)
}
