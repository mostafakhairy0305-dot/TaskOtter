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
	"SQLFLUFF_VERSION",
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
		[]string{"lint", "DIALECT_OVERRIDE=postgres", "TARGETS_OVERRIDE=./migrations"},
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
