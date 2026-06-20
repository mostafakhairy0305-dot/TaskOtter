package resolver_test

import (
	"errors"
	"strings"
	"testing"

	"github.com/mostafakhairy0305-dot/TaskOtter/internal/config"
	"github.com/mostafakhairy0305-dot/TaskOtter/internal/resolver"
)

func catalog(names ...string) map[string]struct{} {
	cat := make(map[string]struct{}, len(names))
	for _, name := range names {
		cat[name] = struct{}{}
	}

	return cat
}

func TestResolveNonNodeTask(t *testing.T) {
	t.Parallel()

	res, err := resolver.Resolve("go", catalog("go"), "", "")
	if err != nil {
		t.Fatal(err)
	}

	if res.SourceModule != "go" {
		t.Fatalf("got %q", res.SourceModule)
	}
}

func TestResolveNodeVariants(t *testing.T) {
	t.Parallel()

	cat := catalog(
		"eslint-npm-fnm", "eslint-npm-nvm", "eslint-yarn-fnm", "eslint-yarn-nvm",
		"eslint-pnpm-fnm", "eslint-pnpm-nvm", "eslint-bun",
	)

	cases := []struct {
		pm   config.PackageManager
		vm   config.VersionManager
		want string
	}{
		{config.PMNPM, config.VMFnm, "eslint-npm-fnm"},
		{config.PMNPM, config.VMNvm, "eslint-npm-nvm"},
		{config.PMYarn, config.VMFnm, "eslint-yarn-fnm"},
		{config.PMYarn, config.VMNvm, "eslint-yarn-nvm"},
		{config.PMPnpm, config.VMFnm, "eslint-pnpm-fnm"},
		{config.PMPnpm, config.VMNvm, "eslint-pnpm-nvm"},
		{config.PMBun, "", "eslint-bun"},
	}
	for _, testCase := range cases {
		res, err := resolver.Resolve("eslint", cat, testCase.pm, testCase.vm)
		if err != nil {
			t.Fatalf("%+v: %v", testCase, err)
		}

		if res.SourceModule != testCase.want {
			t.Fatalf("%+v: got %q", testCase, res.SourceModule)
		}
	}
}

func TestNodeTaskRequiresPackageManager(t *testing.T) {
	t.Parallel()

	cat := catalog("eslint-bun")

	_, err := resolver.Resolve("eslint", cat, "", "")
	if err == nil {
		t.Fatal("expected error")
	}

	if !strings.Contains(err.Error(), "requires js configuration") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestNpmRequiresVersionManager(t *testing.T) {
	t.Parallel()

	cat := catalog("eslint-npm-fnm")

	_, err := resolver.Resolve("eslint", cat, config.PMNPM, "")
	if err == nil {
		t.Fatal("expected error")
	}

	if !strings.Contains(err.Error(), "js.version-manager required") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestMissingTaskCloseMatches(t *testing.T) {
	t.Parallel()

	cat := catalog("eslint-bun", "eslint-npm-fnm")

	_, err := resolver.Resolve("eslit", cat, config.PMBun, "")
	if err == nil {
		t.Fatal("expected error")
	}

	resolveErr := &resolver.ResolveError{
		LogicalTask:  "",
		Attempted:    "",
		Message:      "",
		CloseMatches: nil,
	}

	ok := errors.As(err, &resolveErr)
	if !ok {
		t.Fatalf("unexpected error type: %T", err)
	}

	if len(resolveErr.CloseMatches) == 0 {
		t.Fatal("expected close matches")
	}
}
