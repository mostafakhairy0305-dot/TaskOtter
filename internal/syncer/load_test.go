package syncer_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/mostafakhairy0305-dot/taskotter-sync-action/internal/syncer"
)

func TestLoadMetadataCorruptFails(t *testing.T) {
	path := filepath.Join(t.TempDir(), "metadata.yml")
	if err := os.WriteFile(path, []byte("{{bad"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := syncer.LoadMetadata(path); err == nil {
		t.Fatal("expected corrupt metadata error")
	}
}

func TestLoadLockCorruptFails(t *testing.T) {
	path := filepath.Join(t.TempDir(), "lock.yml")
	if err := os.WriteFile(path, []byte("{{bad"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := syncer.LoadLock(path); err == nil {
		t.Fatal("expected corrupt lock error")
	}
}
