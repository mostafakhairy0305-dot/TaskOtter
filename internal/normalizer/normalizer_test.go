package normalizer_test

import (
	"strings"
	"testing"

	"github.com/mostafakhairy0305-dot/taskotter-sync-action/internal/normalizer"
)

func TestNormalizeExamples(t *testing.T) {
	cases := map[string]string{
		"eslint-pnpm-fnm": "eslint",
		"eslint-pnpm-nvm": "eslint",
		"eslint-npm-fnm":  "eslint",
		"eslint-npm-nvm":  "eslint",
		"eslint-yarn-fnm": "eslint",
		"eslint-yarn-nvm": "eslint",
		"eslint-bun":      "eslint",
		"pnpm-fnm":        "pnpm",
		"npm-nvm":         "npm",
		"yarn-fnm":        "yarn",
		"corepack-fnm":    "corepack",
		"corepack-nvm":    "corepack",
		"fnm":             "fnm",
		"nvm":             "nvm",
		"bun":             "bun",
		"go":              "go",
	}
	for source, want := range cases {
		got, err := normalizer.Normalize(source)
		if err != nil {
			t.Fatalf("Normalize(%q) error = %v", source, err)
		}
		if got != want {
			t.Fatalf("Normalize(%q) = %q, want %q", source, got, want)
		}
	}
}

func TestLongestSuffixFirst(t *testing.T) {
	got, err := normalizer.Normalize("eslint-pnpm-fnm")
	if err != nil {
		t.Fatal(err)
	}
	if got != "eslint" {
		t.Fatalf("got %q", got)
	}
}

func TestDestinationCollision(t *testing.T) {
	_, err := normalizer.BuildDestinationMap([]string{"eslint-pnpm-fnm", "eslint-bun"})
	if err == nil {
		t.Fatal("expected collision error")
	}
	if !strings.Contains(err.Error(), "Destination collision") {
		t.Fatalf("unexpected error: %v", err)
	}
}
