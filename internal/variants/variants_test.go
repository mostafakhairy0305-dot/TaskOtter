package variants_test

import (
	"testing"

	"github.com/mostafakhairy0305-dot/taskotter-sync-action/internal/config"
	"github.com/mostafakhairy0305-dot/taskotter-sync-action/internal/variants"
)

func TestBuildSourceModule(t *testing.T) {
	got, err := variants.BuildSourceModule("eslint", config.PMPnpm, config.VMFnm)
	if err != nil {
		t.Fatal(err)
	}
	if got != "eslint-pnpm-fnm" {
		t.Fatalf("got %q", got)
	}
	got, err = variants.BuildSourceModule("eslint", config.PMBun, "")
	if err != nil {
		t.Fatal(err)
	}
	if got != "eslint-bun" {
		t.Fatalf("got %q", got)
	}
}

func TestIsNodeToolVariant(t *testing.T) {
	if !variants.IsNodeToolVariant("eslint-pnpm-fnm", "eslint") {
		t.Fatal("expected variant")
	}
	if variants.IsNodeToolVariant("go", "go") {
		t.Fatal("go is not a node variant")
	}
}

func TestStripOneSuffix(t *testing.T) {
	got, ok := variants.StripOneSuffix("eslint-pnpm-fnm")
	if !ok || got != "eslint" {
		t.Fatalf("got %q ok=%t", got, ok)
	}
}
