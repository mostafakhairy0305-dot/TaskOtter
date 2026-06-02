package bruno_test

import (
	"testing"

	"github.com/mostafakhairy0305-dot/TaskOtter/internal/tasktest"
)

var publicTasks = []string{
	"ci",
	"help",
	"install",
	"run",
	"version",
}

var publicVars = []string{
	"COLLECTION",
	"ENV",
	"EXTRA_ARGS",
	"PM",
}

func TestTaskfileModuleContract(t *testing.T) {
	tasktest.AssertModule(t, "bruno", publicTasks, publicVars)
}

func TestRepresentativeDryRuns(t *testing.T) {
	tasktest.AssertDryRunContains(t, "bruno",
		[]string{"run", "PM=pnpm"},
		"js:pnpm:exec",
		`run .`,
	)

	tasktest.AssertDryRunContains(t, "bruno",
		[]string{"ci", "PM=npm", "ENV=staging"},
		"js:npm:exec",
		"--bail",
		"staging",
	)

	tasktest.AssertDryRunContains(t, "bruno",
		[]string{"run", "PM=bun", "COLLECTION=./api", "--", "--reporter-json results.json"},
		"js:bun:exec",
		"./api",
	)
}
