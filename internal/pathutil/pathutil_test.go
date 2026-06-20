package pathutil_test

import (
	"testing"

	"github.com/mostafakhairy0305-dot/TaskOtter/internal/pathutil"
)

func TestHasFolderPrefix(t *testing.T) {
	cases := []struct {
		path   string
		folder string
		want   bool
	}{
		{"taskfiles/go", "taskfiles", true},
		{"taskfiles/go/Taskfile.yml", "taskfiles", true},
		{"task", "task", true},
		{"taskfiles-extra/foo", "task", false},
		{"task/extra", "taskfiles", false},
		{"", "taskfiles", false},
	}
	for _, tc := range cases {
		got := pathutil.HasFolderPrefix(tc.path, tc.folder)
		if got != tc.want {
			t.Fatalf("HasFolderPrefix(%q, %q) = %t, want %t", tc.path, tc.folder, got, tc.want)
		}
	}
}
