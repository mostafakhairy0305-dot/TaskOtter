package taskotter_test

import (
	"context"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"
	"testing"
	"time"

	"gopkg.in/yaml.v3"
)

func TestEveryTaskfileDirectoryHasReadmeAndTest(t *testing.T) {
	err := filepath.WalkDir("taskfiles", func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() || entry.Name() != "Taskfile.yml" {
			return nil
		}

		dir := filepath.Dir(path)
		if !fileExists(filepath.Join(dir, "README.md")) {
			t.Fatalf("%s must include README.md", dir)
		}

		tests, err := filepath.Glob(filepath.Join(dir, "*_test.go"))
		if err != nil {
			return err
		}
		if len(tests) == 0 {
			t.Fatalf("%s must include a *_test.go file", dir)
		}

		return nil
	})
	if err != nil {
		t.Fatalf("walk taskfiles: %v", err)
	}
}

func TestRootIncludesEveryTaskfileModule(t *testing.T) {
	content, err := os.ReadFile("Taskfile.yml")
	if err != nil {
		t.Fatalf("read root Taskfile: %v", err)
	}

	var root struct {
		Includes map[string]any `yaml:"includes"`
	}
	if err := yaml.Unmarshal(content, &root); err != nil {
		t.Fatalf("parse root Taskfile: %v", err)
	}

	var expected []string
	err = filepath.WalkDir("taskfiles", func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() || entry.Name() != "Taskfile.yml" {
			return nil
		}
		expected = append(expected, filepath.Base(filepath.Dir(path)))
		return nil
	})
	if err != nil {
		t.Fatalf("walk taskfiles: %v", err)
	}

	var actual []string
	for name := range root.Includes {
		actual = append(actual, name)
	}

	slices.Sort(expected)
	slices.Sort(actual)
	if !slices.Equal(expected, actual) {
		t.Fatalf("root Taskfile include drift\nexpected: %v\nactual:   %v", expected, actual)
	}
}

func TestRootToolDefaultsStayNamespaced(t *testing.T) {
	output := rootDryRun(t, "prettier:check")

	for _, token := range []string{
		"prettier:js:npm:exec",
		"--check",
		".prettierignore",
	} {
		if !strings.Contains(output, token) {
			t.Fatalf("root prettier dry-run missing %q\noutput:\n%s", token, output)
		}
	}

	if strings.Contains(output, "**/*.{css,scss,sass,less,vue,svelte,astro} --check") {
		t.Fatalf("prettier picked up stylelint targets from another include\noutput:\n%s", output)
	}
}

func TestRootAggregatesForwardCommonOverrides(t *testing.T) {
	output := rootDryRun(t, "check", "PM=pnpm", "TARGETS=src")

	for _, token := range []string{
		"eslint:js:pnpm:exec",
		"prettier:js:pnpm:exec",
		"biome:js:pnpm:exec",
		"stylelint:js:pnpm:exec",
		"knip:js:pnpm:exec",
		"depcheck:js:pnpm:exec",
	} {
		if !strings.Contains(output, token) {
			t.Fatalf("root aggregate dry-run missing %q\noutput:\n%s", token, output)
		}
	}
}

func TestRootVaultSnapshotHonorsSnapshotFileOverride(t *testing.T) {
	output := rootDryRun(t, "vault:snapshot", "SNAPSHOT_FILE=custom.snap")

	for _, token := range []string{`vault operator raft snapshot save`, `"custom.snap"`} {
		if !strings.Contains(output, token) {
			t.Fatalf("root vault snapshot did not use SNAPSHOT_FILE override (missing %q)\noutput:\n%s", token, output)
		}
	}
	if strings.Contains(output, `"src/index.ts"`) {
		t.Fatalf("root vault snapshot picked up the TypeScript FILE default\noutput:\n%s", output)
	}
	if strings.Contains(output, `"Dockerfile"`) {
		t.Fatalf("root vault snapshot picked up the Docker FILE default\noutput:\n%s", output)
	}
}

func TestRootVaultSnapshotHonorsVaultFileOverride(t *testing.T) {
	output := rootDryRun(t, "vault:snapshot", "VAULT_FILE=my.snap")

	for _, token := range []string{`vault operator raft snapshot save`, `"my.snap"`} {
		if !strings.Contains(output, token) {
			t.Fatalf("root vault snapshot did not use VAULT_FILE override (missing %q)\noutput:\n%s", token, output)
		}
	}
}

func TestRootVaultLoginForwardsDirectToken(t *testing.T) {
	const secret = "review-root-token"
	output := rootDryRun(t, "vault:login:root-token", "ROOT_TOKEN="+secret)

	for _, token := range []string{`printf '%s' "$VAULT_LOGIN_ROOT_TOKEN"`, "-method=token", "-no-print"} {
		if !strings.Contains(output, token) {
			t.Fatalf("root vault token login missing %q\noutput:\n%s", token, output)
		}
	}
	if strings.Contains(output, secret) {
		t.Fatalf("root vault token login exposed ROOT_TOKEN in command output\noutput:\n%s", output)
	}
}

func TestRootVaultLoginForwardsAppRoleCredentials(t *testing.T) {
	const (
		roleID   = "review-role-id"
		secretID = "review-secret-id"
		mount    = "review-approle"
	)
	output := rootDryRun(
		t,
		"vault:login:approle",
		"ROLE_ID="+roleID,
		"SECRET_ID="+secretID,
		"APPROLE_MOUNT="+mount,
	)

	for _, token := range []string{
		`printf '%s' "$VAULT_LOGIN_SECRET_ID"`,
		`| vault write`,
		`-field=token`,
		`"auth/${VAULT_LOGIN_APPROLE_MOUNT}/login"`,
		`"$VAULT_LOGIN_ROLE_ID"`,
		"secret_id=-",
		`| vault login`,
		"-method=token",
		"-no-print",
	} {
		if !strings.Contains(output, token) {
			t.Fatalf("root vault AppRole login missing %q\noutput:\n%s", token, output)
		}
	}
	for _, secret := range []string{roleID, secretID, mount} {
		if strings.Contains(output, secret) {
			t.Fatalf("root vault AppRole login exposed credential or mount in command output\noutput:\n%s", output)
		}
	}
}

func TestRootVaultDescriptionsAdvertiseVaultFileOverride(t *testing.T) {
	rootContent, err := os.ReadFile("Taskfile.yml")
	if err != nil {
		t.Fatalf("read root Taskfile: %v", err)
	}

	var root struct {
		Includes map[string]struct {
			Vars map[string]string `yaml:"vars"`
		} `yaml:"includes"`
	}
	if err := yaml.Unmarshal(rootContent, &root); err != nil {
		t.Fatalf("parse root Taskfile: %v", err)
	}

	vaultInclude, ok := root.Includes["vault"]
	if !ok {
		t.Fatal("root Taskfile missing vault include")
	}
	if got := vaultInclude.Vars["VAULT_FILE_OVERRIDE"]; got != "{{.VAULT_FILE}}" {
		t.Fatalf("root vault include should forward VAULT_FILE, got %q", got)
	}

	vaultContent, err := os.ReadFile(filepath.Join("taskfiles", "vault", "Taskfile.yml"))
	if err != nil {
		t.Fatalf("read vault Taskfile: %v", err)
	}

	var vault struct {
		Tasks map[string]struct {
			Desc string `yaml:"desc"`
		} `yaml:"tasks"`
	}
	if err := yaml.Unmarshal(vaultContent, &vault); err != nil {
		t.Fatalf("parse vault Taskfile: %v", err)
	}

	for _, name := range []string{"snapshot", "restore"} {
		task, ok := vault.Tasks[name]
		if !ok {
			t.Fatalf("vault Taskfile missing %s task", name)
		}
		if !strings.Contains(task.Desc, "root VAULT_FILE=path") {
			t.Fatalf("vault %s desc should advertise root VAULT_FILE=path, got %q", name, task.Desc)
		}
	}
}

func rootDryRun(t *testing.T, args ...string) string {
	t.Helper()

	allArgs := append([]string{"--dry", "--yes", "--verbose"}, args...)
	return rootTaskOutput(t, allArgs...)
}

func rootTaskOutput(t *testing.T, args ...string) string {
	t.Helper()

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	// Resolve the root Taskfile path absolutely so we can run from an isolated
	// project directory (needed for package.json and fnm preconditions).
	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	rootTaskfile := filepath.Join(wd, "Taskfile.yml")

	projectDir, env := setupRootDryRunEnv(t)
	allArgs := append([]string{"--taskfile", rootTaskfile}, args...)
	cmd := exec.CommandContext(ctx, "task", allArgs...)
	cmd.Dir = projectDir
	cmd.Env = env

	output, err := cmd.CombinedOutput()
	if ctx.Err() == context.DeadlineExceeded {
		t.Fatalf("task command timed out: task %s", strings.Join(allArgs, " "))
	}
	if err != nil {
		t.Fatalf("task command failed: task %s\nerror: %v\noutput:\n%s", strings.Join(allArgs, " "), err, string(output))
	}

	return string(output)
}

func setupRootDryRunEnv(t *testing.T) (string, []string) {
	t.Helper()

	home := t.TempDir()
	projectDir := t.TempDir()
	binDir := filepath.Join(projectDir, ".stub-bin")

	for _, dir := range []string{
		binDir,
		filepath.Join(home, ".bun", "bin"),
		filepath.Join(home, ".local", "share", "fnm"),
	} {
		if err := os.MkdirAll(dir, 0755); err != nil {
			t.Fatalf("create stub dir %s: %v", dir, err)
		}
	}

	const stub = "#!/usr/bin/env bash\nexit 0\n"
	for _, name := range []string{
		"fnm", "node", "npm", "npx", "pnpm", "yarn", "bun", "corepack",
		"prettier", "eslint", "biome", "stylelint", "knip", "depcheck", "bru", "vault",
	} {
		if err := os.WriteFile(filepath.Join(binDir, name), []byte(stub), 0755); err != nil {
			t.Fatalf("write stub %s: %v", name, err)
		}
	}

	if err := os.WriteFile(filepath.Join(home, ".bun", "bin", "bun"), []byte(stub), 0755); err != nil {
		t.Fatalf("write bun file stub: %v", err)
	}
	if err := os.WriteFile(filepath.Join(home, ".local", "share", "fnm", "fnm"), []byte(stub), 0755); err != nil {
		t.Fatalf("write fnm file stub: %v", err)
	}
	if err := os.WriteFile(filepath.Join(projectDir, "package.json"), []byte("{}\n"), 0644); err != nil {
		t.Fatalf("write package.json: %v", err)
	}

	env := os.Environ()
	env = rootSetEnv(env, "HOME", home)
	env = rootSetEnv(env, "PATH", binDir+":"+rootGetEnv(env, "PATH"))
	env = rootSetEnv(env, "CI", "true")
	env = rootSetEnv(env, "NO_COLOR", "1")
	env = rootSetEnv(env, "TASK_ASSUME_YES", "true")

	return projectDir, env
}

func rootSetEnv(env []string, key, value string) []string {
	prefix := key + "="
	for i, item := range env {
		if strings.HasPrefix(item, prefix) {
			env[i] = prefix + value
			return env
		}
	}
	return append(env, prefix+value)
}

func rootGetEnv(env []string, key string) string {
	prefix := key + "="
	for _, item := range env {
		if strings.HasPrefix(item, prefix) {
			return strings.TrimPrefix(item, prefix)
		}
	}
	return ""
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}
