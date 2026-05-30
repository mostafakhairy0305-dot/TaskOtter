package sqlfluff_test

import (
	"testing"

	"github.com/mostafakhairy0305-dot/TaskOtter/internal/tasktest"
)

var publicTasks = []string{
	"config:init",
	"fix",
	"install",
	"install:undo",
	"lint",
	"parse",
	"upgrade",
	"version",
}

var publicVars = []string{
	"CONFIG",
	"DIALECT",
	"EXTRA_ARGS",
	"TARGETS",
	"UV_LOAD",
}

func TestTaskfileModuleContract(t *testing.T) {
	tasktest.AssertModule(t, "sqlfluff", publicTasks, publicVars)
}

func TestRepresentativeDryRuns(t *testing.T) {
	tasktest.AssertDryRunContains(t, "sqlfluff",
		[]string{"lint"},
		"sqlfluff",
		"lint",
		".",
	)

	tasktest.AssertDryRunContains(t, "sqlfluff",
		[]string{"lint", "DIALECT=postgres", "TARGETS=./migrations"},
		"sqlfluff",
		"--dialect",
		"postgres",
		"./migrations",
	)

	tasktest.AssertDryRunContains(t, "sqlfluff",
		[]string{"fix"},
		"sqlfluff",
		"fix",
	)

	tasktest.AssertDryRunContains(t, "sqlfluff",
		[]string{"version"},
		"sqlfluff",
		"--version",
	)
}
