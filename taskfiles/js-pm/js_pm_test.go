package jspm_test

import (
	"testing"

	"github.com/mostafakhairy0305-dot/TaskOtter/taskfiles/internal/tasktest"
)

var internalTasks = []string{
	"exec",
	"exec:ignore",
	"install",
}

var publicVars = []string{
	"ARGS",
	"BINARY",
	"EXTRA_ARGS",
	"IGNORE_FLAG",
	"IGNORE_PATH",
	"PACKAGES",
	"PM",
}

func TestTaskfileModuleContract(t *testing.T) {
	tasktest.AssertModule(t, "js-pm", nil, publicVars)

	tf := tasktest.LoadTaskfile(t, "js-pm")
	for _, name := range internalTasks {
		task, ok := tf.Tasks[name]
		if !ok {
			t.Fatalf("internal task %q is missing", name)
		}
		if !task.Internal {
			t.Fatalf("task %q must remain internal", name)
		}
	}
}
