package archive_test

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/mostafakhairy0305-dot/TaskOtter/internal/archive"
)

func buildTarGz(t *testing.T, entries map[string][]byte) []byte {
	t.Helper()
	var buf bytes.Buffer
	gz := gzip.NewWriter(&buf)
	tw := tar.NewWriter(gz)
	for name, content := range entries {
		hdr := &tar.Header{
			Name: name,
			Mode: 0o644,
			Size: int64(len(content)),
		}
		if err := tw.WriteHeader(hdr); err != nil {
			t.Fatal(err)
		}
		if _, err := tw.Write(content); err != nil {
			t.Fatal(err)
		}
	}
	if err := tw.Close(); err != nil {
		t.Fatal(err)
	}
	if err := gz.Close(); err != nil {
		t.Fatal(err)
	}
	return buf.Bytes()
}

func TestExtractValidArchive(t *testing.T) {
	data := buildTarGz(t, map[string][]byte{
		"repo-main/taskfiles/go/Taskfile.yml": []byte("version: \"3\"\n"),
	})
	dest := t.TempDir()
	_, err := archive.ExtractTarGz(bytes.NewReader(data), dest)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(dest, "taskfiles/go/Taskfile.yml")); err != nil {
		t.Fatal(err)
	}
}

func TestRejectTraversal(t *testing.T) {
	data := buildTarGz(t, map[string][]byte{
		"repo-main/../../etc/passwd": []byte("nope"),
	})
	dest := t.TempDir()
	_, err := archive.ExtractTarGz(bytes.NewReader(data), dest)
	if err == nil {
		t.Fatal("expected traversal rejection")
	}
	if !strings.Contains(err.Error(), "unsafe") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestRejectSymlink(t *testing.T) {
	var buf bytes.Buffer
	gz := gzip.NewWriter(&buf)
	tw := tar.NewWriter(gz)
	hdr := &tar.Header{Name: "repo-main/link", Typeflag: tar.TypeSymlink, Linkname: "/etc/passwd"}
	if err := tw.WriteHeader(hdr); err != nil {
		t.Fatal(err)
	}
	_ = tw.Close()
	_ = gz.Close()
	dest := t.TempDir()
	_, err := archive.ExtractTarGz(bytes.NewReader(buf.Bytes()), dest)
	if err == nil {
		t.Fatal("expected symlink rejection")
	}
}

func TestSkipPAXGlobalHeader(t *testing.T) {
	data, err := os.ReadFile("testdata/github-pax.tar.gz")
	if err != nil {
		t.Fatal(err)
	}
	dest := t.TempDir()
	_, err = archive.ExtractTarGz(bytes.NewReader(data), dest)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(dest, "taskfiles/go/Taskfile.yml")); err != nil {
		t.Fatal(err)
	}
}
