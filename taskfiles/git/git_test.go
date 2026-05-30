package git_test

import (
	"testing"

	"github.com/mostafakhairy0305-dot/TaskOtter/internal/tasktest"
)

var publicTasks = []string{
	"add",
	"add:all",
	"auth:setup",
	"branch:create",
	"branch:delete",
	"branch:list",
	"branch:rename",
	"branch:switch",
	"clean",
	"clone",
	"commit",
	"commit:amend",
	"config:list",
	"config:user",
	"diff",
	"diff:staged",
	"fetch",
	"help",
	"init",
	"log",
	"log:graph",
	"pr:create",
	"pr:open",
	"pull",
	"push",
	"push:force",
	"release:create",
	"remote:add",
	"remote:list",
	"remote:remove",
	"remote:set-url",
	"reset:hard",
	"reset:soft",
	"stash",
	"stash:drop",
	"stash:list",
	"stash:pop",
	"status",
	"sync",
	"tag:create",
	"tag:delete",
	"tag:list",
	"tag:push",
	"version",
}

var publicVars = []string{
	"BASE",
	"BODY",
	"BRANCH",
	"CLONE_DIR",
	"COMMIT",
	"COMMIT_MSG",
	"EMAIL",
	"EXTRA_ARGS",
	"FILES",
	"MERGE_METHOD",
	"MESSAGE",
	"NAME",
	"NOTES",
	"OWNER",
	"REMOTE",
	"REPO",
	"STASH_INDEX",
	"TAG",
	"TITLE",
	"URL",
}

func TestTaskfileModuleContract(t *testing.T) {
	tasktest.AssertModule(t, "git", publicTasks, publicVars)
}

func TestRepresentativeDryRuns(t *testing.T) {
	tasktest.AssertDryRunContains(t, "git",
		[]string{"commit", "COMMIT_MSG=feat: add login page"},
		"git commit -m",
		"feat: add login page",
	)

	tasktest.AssertDryRunContains(t, "git",
		[]string{"clone", "OWNER=github", "REPO=cli"},
		"gh repo clone",
		"github/cli",
	)

	tasktest.AssertDryRunContains(t, "git",
		[]string{"pr:create", "TITLE=feat: login page", "BASE=develop"},
		"git push origin HEAD",
		"--title",
		"feat: login page",
		"--base",
		"develop",
	)

	tasktest.AssertDryRunContains(t, "git",
		[]string{"tag:delete", "TAG=v0.1.0"},
		"git tag -d v0.1.0",
		"git push origin --delete v0.1.0",
	)
}
