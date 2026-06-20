package app_test

import (
	"bytes"
	"io"
	"os"
	"strings"
	"testing"

	"github.com/mostafakhairy0305-dot/TaskOtter/internal/app"
)

func captureStderr(t *testing.T, fn func()) string {
	t.Helper()
	old := os.Stderr
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	os.Stderr = w
	fn()
	_ = w.Close()
	os.Stderr = old
	var buf bytes.Buffer
	if _, err := io.Copy(&buf, r); err != nil {
		t.Fatal(err)
	}
	_ = r.Close()
	return buf.String()
}

func TestReportSyncRequiredWithPullRequest(t *testing.T) {
	result := &app.Result{
		Changed:           true,
		PullRequestNumber: "42",
		PullRequestURL:    "https://example.com/pull/42",
	}
	out := captureStderr(t, func() {
		app.ReportSyncRequired(result)
	})
	if !strings.Contains(out, "::error title=TaskOtter sync required::") {
		t.Fatalf("missing error annotation: %s", out)
	}
	if !strings.Contains(out, "TaskOtter opened sync PR #42: https://example.com/pull/42") {
		t.Fatalf("missing PR summary: %s", out)
	}
	if !strings.Contains(out, "::notice title=What happened::") {
		t.Fatalf("missing notice annotation: %s", out)
	}
}

func TestReportSyncRequiredWithoutPullRequest(t *testing.T) {
	result := &app.Result{Changed: true}
	out := captureStderr(t, func() {
		app.ReportSyncRequired(result)
	})
	if !strings.Contains(out, "did not return a pull request URL") {
		t.Fatalf("missing fallback summary: %s", out)
	}
}

func TestSyncRequired(t *testing.T) {
	if !app.SyncRequired(&app.Result{Changed: true}) {
		t.Fatal("expected changed result to require sync")
	}
	if app.SyncRequired(&app.Result{Changed: false}) {
		t.Fatal("expected unchanged result not to require sync")
	}
}
