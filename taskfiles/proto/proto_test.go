package proto_test

import (
	"runtime"
	"testing"

	"github.com/mostafakhairy0305-dot/TaskOtter/internal/tasktest"
)

var publicTasks = []string{
	"gen",
	"install",
	"install:undo",
	"upgrade",
	"ungen",
	"version",
}

var publicVars = []string{
	"GLOBAL_GO_BIN",
	"PROTO_PATH",
	"PROTO_PATTERN",
	"PROTOC_GEN_GO_GRPC_VERSION",
	"PROTOC_GEN_GO_VERSION",
	"PROTOC_VERSION",
}

func TestTaskfileModuleContract(t *testing.T) {
	tasktest.AssertModule(t, "proto", publicTasks, publicVars)
}

func TestRepresentativeDryRuns(t *testing.T) {
	tasktest.AssertDryRunContains(t, "proto",
		[]string{"version"},
		"protoc",
		"--version",
	)

	tasktest.AssertDryRunContains(t, "proto",
		[]string{"gen"},
		"protoc",
		"--go_out",
	)
}

func TestInstallDryRunUsesPlatformPackageManager(t *testing.T) {
	switch runtime.GOOS {
	case "darwin":
		tasktest.AssertDryRunContains(t, "proto",
			[]string{"install"},
			"brew",
			"protobuf",
		)
	case "linux":
		tasktest.AssertDryRunContains(t, "proto",
			[]string{"install"},
			"curl",
			"protoc-",
		)
	default:
		t.Skip("install dry-run is covered on macOS and Linux")
	}
}

func TestUpgradeDryRunUsesPlatformPackageManager(t *testing.T) {
	switch runtime.GOOS {
	case "darwin":
		tasktest.AssertDryRunContains(t, "proto",
			[]string{"upgrade"},
			"brew",
			"protobuf",
		)
	case "linux":
		tasktest.AssertDryRunContains(t, "proto",
			[]string{"upgrade"},
			"curl",
			"protoc-",
		)
	default:
		t.Skip("upgrade dry-run is covered on macOS and Linux")
	}
}

func TestUngenDryRun(t *testing.T) {
	switch runtime.GOOS {
	case "darwin", "linux":
		tasktest.AssertDryRunContains(t, "proto",
			[]string{"ungen"},
			"find",
			".pb.go",
		)
	default:
		t.Skip("ungen dry-run is covered on macOS and Linux")
	}
}
