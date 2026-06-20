package dependency_test

import (
	"testing"

	"github.com/mostafakhairy0305-dot/TaskOtter/internal/dependency"
)

func deps() map[string][]string {
	return map[string][]string{
		"eslint-pnpm-fnm": {"pnpm-fnm"},
		"pnpm-fnm":        {"corepack-fnm", "fnm"},
		"corepack-fnm":    {"fnm"},
		"fnm":             {},
		"go":              {},
	}
}

func TestResolveTransitive(t *testing.T) {
	got, err := dependency.ResolveTransitive([]string{"eslint-pnpm-fnm"}, deps())
	if err != nil {
		t.Fatal(err)
	}
	want := []string{"corepack-fnm", "fnm", "pnpm-fnm"}
	if len(got) != len(want) {
		t.Fatalf("got %#v, want %#v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("got %#v, want %#v", got, want)
		}
	}
}

func TestDuplicateDependencyDeduped(t *testing.T) {
	got, err := dependency.ResolveTransitive([]string{"eslint-pnpm-fnm", "pnpm-fnm"}, deps())
	if err != nil {
		t.Fatal(err)
	}
	for _, dep := range got {
		if dep == "pnpm-fnm" {
			t.Fatal("requested module should not appear in dependencies")
		}
	}
}

func TestMissingDependency(t *testing.T) {
	d := deps()
	d["eslint-pnpm-fnm"] = []string{"missing-mod"}
	_, err := dependency.ResolveTransitive([]string{"eslint-pnpm-fnm"}, d)
	if err == nil {
		t.Fatal("expected missing dependency error")
	}
}

func TestDependencyCycle(t *testing.T) {
	d := map[string][]string{
		"a": {"b"},
		"b": {"c"},
		"c": {"a"},
	}
	_, err := dependency.ResolveTransitive([]string{"a"}, d)
	if err == nil {
		t.Fatal("expected cycle error")
	}
}
