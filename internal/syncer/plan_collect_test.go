package syncer_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/mostafakhairy0305-dot/TaskOtter/internal/syncer"
)

func TestCollectModuleFilesSkipsTestsAndDocs(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	write := func(rel, content string) {
		t.Helper()

		path := filepath.Join(dir, rel)

		err := os.MkdirAll(filepath.Dir(path), 0o755)
		if err != nil {
			t.Fatal(err)
		}

		err = os.WriteFile(path, []byte(content), 0o644)
		if err != nil {
			t.Fatal(err)
		}
	}
	write("Taskfile.yml", "version: \"3\"\n")
	write("README.md", "docs\n")
	write("docs/guide.md", "guide\n")
	write("go_test.go", "package go_test\n")

	withDocs, err := syncer.CollectModuleFiles(dir, true, nil)
	if err != nil {
		t.Fatal(err)
	}

	for _, path := range []string{"Taskfile.yml", "README.md", "docs/guide.md"} {
		if _, ok := withDocs[path]; !ok {
			t.Fatalf("expected %q in sync output", path)
		}
	}

	if _, ok := withDocs["go_test.go"]; ok {
		t.Fatal("test files should never be synced")
	}

	withoutDocs, err := syncer.CollectModuleFiles(dir, false, nil)
	if err != nil {
		t.Fatal(err)
	}

	if _, ok := withoutDocs["README.md"]; ok {
		t.Fatal("README should be excluded when includes-doc=false")
	}

	if _, ok := withoutDocs["docs/guide.md"]; ok {
		t.Fatal("docs/ should be excluded when includes-doc=false")
	}
}
