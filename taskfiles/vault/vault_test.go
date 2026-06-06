package vault_test

import (
	"slices"
	"strings"
	"testing"

	"github.com/mostafakhairy0305-dot/TaskOtter/internal/tasktest"
	"gopkg.in/yaml.v3"
)

var publicTasks = []string{
	"health",
	"init",
	"install",
	"install:undo",
	"kv:get",
	"login",
	"login:approle",
	"login:root-token",
	"peers",
	"restore",
	"root-token",
	"seal",
	"snapshot",
	"status",
	"token:issue:approle",
	"token:revoke-self",
	"unseal",
	"upgrade",
	"verify",
	"version",
}

var publicVars = []string{
	"APPROLE_MOUNT",
	"EXTRA_ARGS",
	"FILE",
	"KEYS_FILE",
	"ROLE_ID",
	"ROOT_TOKEN",
	"SECRET_ID",
	"SHARES",
	"SNAPSHOT_FILE",
	"THRESHOLD",
	"VAULT_ADDR",
}

func TestTaskfileModuleContract(t *testing.T) {
	tasktest.AssertModule(t, "vault", publicTasks, publicVars)
}

func TestInputValidatedTasksDoNotInstallBeforePreconditions(t *testing.T) {
	tf := tasktest.LoadTaskfile(t, "vault")

	for _, name := range []string{"init", "login", "login:approle", "login:root-token", "restore", "unseal"} {
		task := tf.Tasks[name]
		if task.Deps != nil {
			t.Fatalf("%s should run install from cmds after local preconditions, got deps: %#v", name, task.Deps)
		}
	}
}

func TestVerifyDoesNotMaskStatusFailures(t *testing.T) {
	tf := tasktest.LoadTaskfile(t, "vault")
	cmds := taskFieldYAML(t, tf.Tasks["verify"].Cmds)

	if !strings.Contains(cmds, "vault status") {
		t.Fatalf("verify should run vault status\ncmds:\n%s", cmds)
	}
	if strings.Contains(cmds, "vault status || true") {
		t.Fatalf("verify should fail when vault status fails\ncmds:\n%s", cmds)
	}
}

func TestInitDoesNotOverwriteExistingKeysFile(t *testing.T) {
	tf := tasktest.LoadTaskfile(t, "vault")
	task := tf.Tasks["init"]
	preconditions := taskFieldYAML(t, task.Preconditions)
	cmds := taskFieldYAML(t, task.Cmds)

	for _, token := range []string{"test ! -e", "KEYS_FILE already exists"} {
		if !strings.Contains(preconditions, token) {
			t.Fatalf("init should refuse an existing KEYS_FILE with %q\npreconditions:\n%s", token, preconditions)
		}
	}
	for _, token := range []string{`TMP="${KF}.tmp.$$"`, `-format=json > "$TMP"`, `mv "$TMP" "$KF"`} {
		if !strings.Contains(cmds, token) {
			t.Fatalf("init should stage init output safely with %q\ncmds:\n%s", token, cmds)
		}
	}
	if strings.Contains(cmds, `-format=json > "$KF"`) {
		t.Fatalf("init should not redirect operator init directly to KEYS_FILE\ncmds:\n%s", cmds)
	}
}

func TestLoginDoesNotPassRootTokenAsCommandArgument(t *testing.T) {
	tf := tasktest.LoadTaskfile(t, "vault")
	cmds := taskFieldYAML(t, tf.Tasks["login"].Cmds)

	for _, token := range []string{`jq -r '.root_token' "$KF"`, `| vault login`, `-method=token`, `-no-print`} {
		if !strings.Contains(cmds, token) {
			t.Fatalf("login should pipe the root token to vault login stdin (missing %q)\ncmds:\n%s", token, cmds)
		}
	}
	for _, token := range []string{`vault login "$(jq`, `vault login token=`, `vault login "$TOKEN"`} {
		if strings.Contains(cmds, token) {
			t.Fatalf("login should not expose the root token as a command argument\ncmds:\n%s", cmds)
		}
	}
}

func TestLoginRootTokenPipesTokenViaStdin(t *testing.T) {
	tf := tasktest.LoadTaskfile(t, "vault")
	cmds := taskFieldYAML(t, tf.Tasks["login:root-token"].Cmds)

	for _, token := range []string{`printf '%s' "$VAULT_LOGIN_ROOT_TOKEN"`, `| vault login`, `-method=token`, `-no-print`} {
		if !strings.Contains(cmds, token) {
			t.Fatalf("login:root-token should pipe token to vault login stdin (missing %q)\ncmds:\n%s", token, cmds)
		}
	}
	for _, token := range []string{`vault login "$TOKEN"`, `vault login token=`} {
		if strings.Contains(cmds, token) {
			t.Fatalf("login:root-token should not expose token as command argument\ncmds:\n%s", cmds)
		}
	}
}

func TestLoginApproleRequiresBothCredentials(t *testing.T) {
	tf := tasktest.LoadTaskfile(t, "vault")
	task := tf.Tasks["login:approle"]
	preconditions := taskFieldYAML(t, task.Preconditions)
	cmds := taskFieldYAML(t, task.Cmds)

	for _, token := range []string{"ROLE_ID", "SECRET_ID"} {
		if !strings.Contains(preconditions, token) {
			t.Fatalf("login:approle should require %s in preconditions\npreconditions:\n%s", token, preconditions)
		}
	}
	for _, token := range []string{
		`printf '%s' "$VAULT_LOGIN_SECRET_ID"`,
		`| vault write`,
		`-field=token`,
		`"auth/${VAULT_LOGIN_APPROLE_MOUNT}/login"`,
		`"$VAULT_LOGIN_ROLE_ID"`,
		`secret_id=-`,
		`| vault login`,
		`-method=token`,
		`-no-print`,
	} {
		if !strings.Contains(cmds, token) {
			t.Fatalf("login:approle should exchange AppRole credentials without exposing the secret_id (missing %q)\ncmds:\n%s", token, cmds)
		}
	}
	for _, token := range []string{`secret_id="{{`, `secret_id="$VAULT_LOGIN_SECRET_ID"`} {
		if strings.Contains(cmds, token) {
			t.Fatalf("login:approle should not expose secret_id as a command argument\ncmds:\n%s", cmds)
		}
	}
}

func TestLinuxParentTasksGuardUnsupportedPackageManagers(t *testing.T) {
	tf := tasktest.LoadTaskfile(t, "vault")

	for _, name := range []string{"_install:linux", "_install:undo:linux", "_upgrade:linux"} {
		preconditions := taskFieldYAML(t, tf.Tasks[name].Preconditions)
		for _, token := range []string{"apt-get", "dnf"} {
			if !strings.Contains(preconditions, token) {
				t.Fatalf("%s should guard unsupported Linux package managers with %s\npreconditions:\n%s", name, token, preconditions)
			}
		}
	}
}

func TestStrictShellSetOnSensitiveTasks(t *testing.T) {
	tf := tasktest.LoadTaskfile(t, "vault")

	for _, name := range []string{"health", "init", "login", "login:approle", "login:root-token", "restore", "unseal"} {
		task := tf.Tasks[name]
		for _, option := range []string{"errexit", "nounset", "pipefail"} {
			if !slices.Contains(task.Set, option) {
				t.Fatalf("%s should set %s, got %#v", name, option, task.Set)
			}
		}
	}
}

func taskFieldYAML(t *testing.T, value any) string {
	t.Helper()

	content, err := yaml.Marshal(value)
	if err != nil {
		t.Fatalf("marshal task field: %v", err)
	}

	return string(content)
}
