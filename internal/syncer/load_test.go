package syncer_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/mostafakhairy0305-dot/TaskOtter/internal/syncer"
)

func TestLoadMetadataCorruptFails(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	rel := "metadata.yml"

	err := os.WriteFile(filepath.Join(root, rel), []byte("{{bad"), 0o644)
	if err != nil {
		t.Fatal(err)
	}

	_, err = syncer.LoadMetadata(root, rel)
	if err == nil {
		t.Fatal("expected corrupt metadata error")
	}
}

func TestLoadLockCorruptFails(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	rel := "lock.yml"

	err := os.WriteFile(filepath.Join(root, rel), []byte("{{bad"), 0o644)
	if err != nil {
		t.Fatal(err)
	}

	_, err = syncer.LoadLock(root, rel)
	if err == nil {
		t.Fatal("expected corrupt lock error")
	}
}
