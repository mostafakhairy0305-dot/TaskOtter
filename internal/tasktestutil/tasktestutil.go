package tasktestutil

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
	"testing"
	"time"

	"gopkg.in/yaml.v3"
)

// PublicTaskSpec describes expectations for a single public task.
type PublicTaskSpec struct {
	Name                  string
	Args                  map[string]string
	MustDryRunWithArgs    bool
	MustDryRunWithoutArgs bool
	ExpectedDefaultTokens []string
	RequiresGroupOutput   bool
	RequiresPrompt        bool
	RequiresSummary       bool
}

// TaskNode wraps a YAML node with its task name for error messages.
type TaskNode struct {
	Name string
	Node *yaml.Node
}

// LoadedTaskfile holds the parsed content of a Taskfile.
type LoadedTaskfile struct {
	Path  string
	Root  TaskNode
	Tasks map[string]TaskNode
}

// CommandResult holds the output of a task invocation.
type CommandResult struct {
	Stdout string
	Stderr string
	Err    error
	Args   []string
}

func (r CommandResult) Combined() string { return r.Stdout + "\n" + r.Stderr }

// SimpleTaskResult holds the output of a simple (non-isolated) task run.
type SimpleTaskResult struct {
	Output string
	Err    error
}

// Field returns the YAML child node for the given mapping key.
func (n TaskNode) Field(name string) *yaml.Node {
	if n.Node == nil || n.Node.Kind != yaml.MappingNode {
		return nil
	}
	for i := 0; i < len(n.Node.Content); i += 2 {
		if n.Node.Content[i].Value == name {
			return n.Node.Content[i+1]
		}
	}
	return nil
}

// StringField returns the text value of a scalar mapping field.
func (n TaskNode) StringField(name string) string { return NodeText(n.Field(name)) }

// BoolField returns true when the mapping field is the string "true" (case-insensitive).
func (n TaskNode) BoolField(name string) bool {
	f := n.Field(name)
	if f == nil {
		return false
	}
	return strings.EqualFold(f.Value, "true")
}

// ModuleRoot walks up from the working directory to find the nearest ancestor
// that contains a Taskfile.yml or Taskfile.yaml. When tests run from a module
// directory (e.g. taskfiles/bun/) this returns that directory, not the repo root.
func ModuleRoot(t *testing.T) string {
	t.Helper()

	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get working directory: %v", err)
	}

	current := wd
	for {
		if FileExists(filepath.Join(current, "Taskfile.yml")) ||
			FileExists(filepath.Join(current, "Taskfile.yaml")) {
			return current
		}
		parent := filepath.Dir(current)
		if parent == current {
			t.Fatal("could not find Taskfile.yml or Taskfile.yaml")
		}
		current = parent
	}
}

// ModuleTaskfilePath returns the path of the Taskfile.yml found by ModuleRoot.
func ModuleTaskfilePath(t *testing.T) string {
	t.Helper()

	root := ModuleRoot(t)
	for _, name := range []string{"Taskfile.yml", "Taskfile.yaml"} {
		if p := filepath.Join(root, name); FileExists(p) {
			return p
		}
	}

	t.Fatal("could not find Taskfile.yml or Taskfile.yaml")
	return ""
}

// LoadTaskfile parses the Taskfile in the module root and returns a LoadedTaskfile.
func LoadTaskfile(t *testing.T) LoadedTaskfile {
	t.Helper()

	path := ModuleTaskfilePath(t)
	content := ReadFile(t, path)

	var doc yaml.Node
	if err := yaml.Unmarshal([]byte(content), &doc); err != nil {
		t.Fatalf("failed to parse Taskfile: %v", err)
	}

	root := DocumentRoot(t, &doc)
	tasksNode := MappingField(root, "tasks")
	if tasksNode == nil {
		t.Fatal("Taskfile has no tasks map")
	}

	tasks := map[string]TaskNode{}
	for i := 0; i < len(tasksNode.Content); i += 2 {
		key := tasksNode.Content[i]
		tasks[key.Value] = TaskNode{Name: key.Value, Node: tasksNode.Content[i+1]}
	}

	return LoadedTaskfile{
		Path:  path,
		Root:  TaskNode{Name: "root", Node: root},
		Tasks: tasks,
	}
}

// MustTask returns the named task or fails the test if it is missing.
func MustTask(t *testing.T, tf LoadedTaskfile, name string) TaskNode {
	t.Helper()

	task, ok := tf.Tasks[name]
	if !ok {
		t.Fatalf("expected public task %q is missing", name)
	}
	return task
}

// HasAlias reports whether the task declares the given alias.
func HasAlias(task TaskNode, alias string) bool {
	aliases := task.Field("aliases")
	if aliases == nil || aliases.Kind != yaml.SequenceNode {
		return false
	}
	for _, item := range aliases.Content {
		if item.Value == alias {
			return true
		}
	}
	return false
}

// RunTask runs the task binary with the given args and returns the result.
func RunTask(t *testing.T, root string, env []string, args ...string) CommandResult {
	t.Helper()
	return RunTaskTimeout(t, root, env, 2*time.Minute, args...)
}

// RunTaskTimeout runs the task binary with a custom timeout.
func RunTaskTimeout(t *testing.T, root string, env []string, timeout time.Duration, args ...string) CommandResult {
	t.Helper()

	taskBin := os.Getenv("TASK_BIN")
	if taskBin == "" {
		taskBin = "task"
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, taskBin, args...)
	cmd.Dir = root

	if env != nil {
		cmd.Env = env
	}

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	cmd.WaitDelay = 5 * time.Second

	err := cmd.Run()

	return CommandResult{
		Stdout: stdout.String(),
		Stderr: stderr.String(),
		Err:    err,
		Args:   args,
	}
}

// RunSimpleTask runs task in the given directory and returns combined output.
// Use this for the simple pnpm/yarn-style tests that don't need separate stdout/stderr.
func RunSimpleTask(t *testing.T, dir string, env []string, args ...string) SimpleTaskResult {
	t.Helper()

	cmd := exec.Command("task", args...)
	cmd.Dir = dir
	cmd.Env = env
	out, err := cmd.CombinedOutput()

	return SimpleTaskResult{Output: string(out), Err: err}
}

// IsolatedEnv returns a clean environment with a temporary HOME for tests that
// must not interact with the real user's shell profile or tool installations.
func IsolatedEnv(t *testing.T) []string {
	t.Helper()

	home := t.TempDir()
	profile := filepath.Join(home, ".bashrc")

	if err := os.WriteFile(profile, []byte(""), 0644); err != nil {
		t.Fatalf("failed to create fake shell profile: %v", err)
	}

	env := os.Environ()
	env = SetEnv(env, "HOME", home)
	env = SetEnv(env, "PROFILE", profile)
	env = SetEnv(env, "ZDOTDIR", home)
	env = SetEnv(env, "CI", "true")
	env = SetEnv(env, "TASK_COLOR", "0")
	env = SetEnv(env, "NO_COLOR", "1")
	env = SetEnv(env, "TASK_ASSUME_YES", "true")

	return env
}

// SetEnv sets or replaces a key=value pair in an env slice.
func SetEnv(env []string, key, value string) []string {
	prefix := key + "="
	for i, item := range env {
		if strings.HasPrefix(item, prefix) {
			env[i] = prefix + value
			return env
		}
	}
	return append(env, prefix+value)
}

// EnvValue returns the value for the given key from an env slice.
func EnvValue(env []string, key string) string {
	prefix := key + "="
	for _, item := range env {
		if strings.HasPrefix(item, prefix) {
			return strings.TrimPrefix(item, prefix)
		}
	}
	return ""
}

// ExpectedPublicTaskNames returns sorted task names from a PublicTaskSpec slice.
func ExpectedPublicTaskNames(specs []PublicTaskSpec) []string {
	names := make([]string, 0, len(specs))
	for _, s := range specs {
		names = append(names, s.Name)
	}
	slices.Sort(names)
	return names
}

// PublicTaskNamesFromTaskfile returns sorted names of public tasks in the Taskfile.
func PublicTaskNamesFromTaskfile(t *testing.T, tf LoadedTaskfile) []string {
	t.Helper()

	var names []string
	for name, task := range tf.Tasks {
		if name == "default" || strings.HasPrefix(name, "_") || task.BoolField("internal") {
			continue
		}
		if task.StringField("desc") != "" {
			names = append(names, name)
		}
	}
	slices.Sort(names)
	return names
}

// TaskArgs converts a map of task variable assignments to "KEY=VALUE" args.
func TaskArgs(args map[string]string) []string {
	if len(args) == 0 {
		return nil
	}
	keys := make([]string, 0, len(args))
	for k := range args {
		keys = append(keys, k)
	}
	slices.Sort(keys)
	out := make([]string, 0, len(keys))
	for _, k := range keys {
		out = append(out, fmt.Sprintf("%s=%s", k, args[k]))
	}
	return out
}

// FormatList formats a string slice as a bulleted list.
func FormatList(values []string) string { return "- " + strings.Join(values, "\n- ") }

// WriteStub writes a stub shell script to dir/name with the given body.
func WriteStub(t *testing.T, dir, name, body string) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(dir, name), []byte(body), 0755); err != nil {
		t.Fatalf("write %s stub: %v", name, err)
	}
}

// SimplePublicTaskNames extracts public task names from a decoded tasks map.
func SimplePublicTaskNames(tasks map[string]any) []string {
	var names []string
	for name, raw := range tasks {
		if name == "default" || strings.HasPrefix(name, "_") {
			continue
		}
		if task, ok := raw.(map[string]any); ok && task["internal"] == true {
			continue
		}
		names = append(names, name)
	}
	slices.Sort(names)
	return names
}

// ReadmePublicTaskNames parses a README and returns task names listed in the
// "## Public Tasks" table (backtick-quoted entries in the first column).
func ReadmePublicTaskNames(content string) []string {
	row := regexp.MustCompile(`^\|\s*` + "`" + `([^` + "`" + `]+)` + "`" + `\s*\|`)
	var names []string
	inTable := false

	for _, line := range strings.Split(content, "\n") {
		trimmed := strings.TrimSpace(line)
		if trimmed == "## Public Tasks" {
			inTable = true
			continue
		}
		if inTable && strings.HasPrefix(trimmed, "## ") {
			break
		}
		if inTable {
			if m := row.FindStringSubmatch(trimmed); len(m) == 2 {
				names = append(names, m[1])
			}
		}
	}

	slices.Sort(names)
	return names
}

// MustRead reads a file and fails the test on error.
func MustRead(t *testing.T, path string) string {
	t.Helper()
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	return string(content)
}

// --- YAML helpers ---

// DocumentRoot returns the root mapping node of a YAML document node.
func DocumentRoot(t *testing.T, doc *yaml.Node) *yaml.Node {
	t.Helper()

	if doc.Kind != yaml.DocumentNode || len(doc.Content) == 0 {
		t.Fatal("invalid YAML document")
	}
	root := doc.Content[0]
	if root.Kind != yaml.MappingNode {
		t.Fatal("Taskfile root must be a YAML mapping")
	}
	return root
}

// MappingField returns the mapping-typed child of root named name, or nil.
func MappingField(root *yaml.Node, name string) *yaml.Node {
	node := NodeMappingValue(root, name)
	if node == nil || node.Kind != yaml.MappingNode {
		return nil
	}
	return node
}

// ScalarField returns the text of the scalar child of root named name.
func ScalarField(root *yaml.Node, name string) string {
	return NodeText(NodeMappingValue(root, name))
}

// NodeMappingValue returns the value node for the given key in a mapping node.
func NodeMappingValue(mapping *yaml.Node, key string) *yaml.Node {
	if mapping == nil || mapping.Kind != yaml.MappingNode {
		return nil
	}
	for i := 0; i < len(mapping.Content); i += 2 {
		if mapping.Content[i].Value == key {
			return mapping.Content[i+1]
		}
	}
	return nil
}

// NodeText returns the trimmed text content of a YAML node.
func NodeText(node *yaml.Node) string {
	if node == nil {
		return ""
	}
	if node.Kind == yaml.ScalarNode {
		return strings.TrimSpace(node.Value)
	}
	var parts []string
	for _, child := range node.Content {
		if t := NodeText(child); t != "" {
			parts = append(parts, t)
		}
	}
	return strings.TrimSpace(strings.Join(parts, " "))
}

// IsEmptyNode reports whether a YAML node is nil or carries no content.
func IsEmptyNode(node *yaml.Node) bool {
	if node == nil {
		return true
	}
	if node.Kind == yaml.ScalarNode {
		return strings.TrimSpace(node.Value) == ""
	}
	return len(node.Content) == 0
}

// FileExists reports whether the given path exists.
func FileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

// ReadFile reads a file and fails the test on error.
func ReadFile(t *testing.T, path string) string {
	t.Helper()
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to read %s: %v", path, err)
	}
	return string(content)
}

// CollectCommandStrings extracts all command/sh scalar values from a task node.
func CollectCommandStrings(node *yaml.Node) []string {
	if node == nil {
		return nil
	}
	var out []string
	switch node.Kind {
	case yaml.ScalarNode:
		if strings.TrimSpace(node.Value) != "" {
			out = append(out, node.Value)
		}
	case yaml.SequenceNode:
		for _, child := range node.Content {
			out = append(out, CollectCommandStrings(child)...)
		}
	case yaml.MappingNode:
		for i := 0; i < len(node.Content); i += 2 {
			key := node.Content[i]
			value := node.Content[i+1]
			switch key.Value {
			case "cmd", "sh":
				if value.Kind == yaml.ScalarNode {
					out = append(out, value.Value)
				}
			case "cmds", "status", "preconditions":
				out = append(out, CollectCommandStrings(value)...)
			}
		}
	}
	return out
}

// ReferencedLocalShellScripts returns all ./path/to/script.sh references in a command string.
func ReferencedLocalShellScripts(command string) []string {
	re := regexp.MustCompile(`(?:^|\s)(\./[A-Za-z0-9_./-]+\.sh)(?:\s|$)`)
	matches := re.FindAllStringSubmatch(command, -1)
	var out []string
	for _, m := range matches {
		if len(m) > 1 {
			out = append(out, m[1])
		}
	}
	return out
}

// --- Assertion helpers ---

func AssertExitCode(t *testing.T, result CommandResult, expected int) {
	t.Helper()

	actual := 0
	if result.Err != nil {
		exitErr, ok := result.Err.(*exec.ExitError)
		if !ok {
			t.Fatalf(
				"command failed without exit code\nargs: %v\nerror: %v\nstdout:\n%s\nstderr:\n%s",
				result.Args, result.Err, result.Stdout, result.Stderr,
			)
		}
		actual = exitErr.ExitCode()
	}

	if actual != expected {
		t.Fatalf(
			"expected exit code %d, got %d\nargs: %v\nerror: %v\nstdout:\n%s\nstderr:\n%s",
			expected, actual, result.Args, result.Err, result.Stdout, result.Stderr,
		)
	}
}

func AssertContains(t *testing.T, value, expected string) {
	t.Helper()
	if !strings.Contains(value, expected) {
		t.Fatalf("expected output to contain %q\n\nOutput:\n%s", expected, value)
	}
}

func AssertNotContains(t *testing.T, value, unexpected string) {
	t.Helper()
	if strings.Contains(value, unexpected) {
		t.Fatalf("expected output not to contain %q\n\nOutput:\n%s", unexpected, value)
	}
}

func AssertNotEmpty(t *testing.T, value, message string) {
	t.Helper()
	if strings.TrimSpace(value) == "" {
		t.Fatal(message)
	}
}

func AssertFileExists(t *testing.T, path string) {
	t.Helper()
	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("expected file %s to exist: %v", path, err)
	}
	if info.IsDir() {
		t.Fatalf("expected file but found directory at %s", path)
	}
}

func AssertDirExists(t *testing.T, path string) {
	t.Helper()
	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("expected directory %s to exist: %v", path, err)
	}
	if !info.IsDir() {
		t.Fatalf("expected directory but found file at %s", path)
	}
}

func AssertDirNotExists(t *testing.T, path string) {
	t.Helper()
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Fatalf("expected %s to not exist, but it does", path)
	}
}

func AssertDirHasEntries(t *testing.T, path string) {
	t.Helper()
	entries, err := os.ReadDir(path)
	if err != nil {
		t.Fatalf("failed to read directory %s: %v", path, err)
	}
	if len(entries) == 0 {
		t.Fatalf("expected %s to contain at least one entry", path)
	}
}

func AssertGithubGroupOutput(t *testing.T, taskName string, outputNode *yaml.Node) {
	t.Helper()

	if outputNode == nil {
		t.Fatalf("task %q requires output.group config but no output config was found", taskName)
	}
	if outputNode.Kind != yaml.MappingNode {
		t.Fatalf("task %q output must use advanced object format, not scalar format", taskName)
	}

	groupNode := NodeMappingValue(outputNode, "group")
	if groupNode == nil || groupNode.Kind != yaml.MappingNode {
		t.Fatalf("task %q output must include group config", taskName)
	}

	begin := NodeText(NodeMappingValue(groupNode, "begin"))
	end := NodeText(NodeMappingValue(groupNode, "end"))
	errorOnly := NodeMappingValue(groupNode, "error_only")

	if begin != "::group::{{.TASK}}" {
		t.Fatalf("task %q output.group.begin must be %q, got %q", taskName, "::group::{{.TASK}}", begin)
	}
	if end != "::endgroup::" {
		t.Fatalf("task %q output.group.end must be %q, got %q", taskName, "::endgroup::", end)
	}
	if errorOnly == nil {
		t.Fatalf("task %q output.group.error_only must be explicitly set to false", taskName)
	}
	if !strings.EqualFold(errorOnly.Value, "false") {
		t.Fatalf("task %q output.group.error_only must be false, got %q", taskName, errorOnly.Value)
	}
}

func AssertTextFileClean(t *testing.T, path, content string) {
	t.Helper()

	if content == "" {
		t.Fatalf("%s is empty", path)
	}
	if strings.Contains(content, "\r\n") {
		t.Fatalf("%s uses CRLF line endings; use LF only", path)
	}
	if strings.Contains(content, "\t") {
		t.Fatalf("%s contains tabs; use spaces in YAML", path)
	}
	if !strings.HasSuffix(content, "\n") {
		t.Fatalf("%s must end with a newline", path)
	}
	for i, line := range strings.Split(content, "\n") {
		if strings.TrimRight(line, " ") != line {
			t.Fatalf("%s has trailing whitespace at line %d", path, i+1)
		}
	}
}

func AssertNoDuplicateMappingKeys(t *testing.T, node *yaml.Node, path string) {
	t.Helper()

	if node == nil {
		return
	}
	if node.Kind == yaml.DocumentNode && len(node.Content) > 0 {
		AssertNoDuplicateMappingKeys(t, node.Content[0], path)
		return
	}
	if node.Kind == yaml.MappingNode {
		seen := map[string]bool{}
		for i := 0; i < len(node.Content); i += 2 {
			key := node.Content[i]
			value := node.Content[i+1]
			if seen[key.Value] {
				t.Fatalf("duplicate YAML key at %s.%s", path, key.Value)
			}
			seen[key.Value] = true
			AssertNoDuplicateMappingKeys(t, value, path+"."+key.Value)
		}
	}
	if node.Kind == yaml.SequenceNode {
		for i, child := range node.Content {
			AssertNoDuplicateMappingKeys(t, child, fmt.Sprintf("%s[%d]", path, i))
		}
	}
}

func AssertNoYamlAliases(t *testing.T, node *yaml.Node, path string) {
	t.Helper()

	if node == nil {
		return
	}
	if node.Kind == yaml.AliasNode {
		t.Fatalf("YAML aliases/anchors are not allowed for clean Taskfile config at %s", path)
	}
	for i, child := range node.Content {
		AssertNoYamlAliases(t, child, fmt.Sprintf("%s[%d]", path, i))
	}
}

func AssertNoPlaceholderText(t *testing.T, taskName, value string) {
	t.Helper()

	upper := strings.ToUpper(value)
	for _, placeholder := range []string{"TODO", "FIXME", "CHANGEME", "REPLACE_ME", "LOREM IPSUM"} {
		if strings.Contains(upper, placeholder) {
			t.Fatalf("task %q contains placeholder text %q", taskName, placeholder)
		}
	}
}

// ValidateJSON returns an error if s is not valid JSON.
func ValidateJSON(s string) error {
	var payload any
	return json.Unmarshal([]byte(s), &payload)
}

// DangerousCommandPatterns returns regexps that match unsafe shell command patterns.
func DangerousCommandPatterns() []*regexp.Regexp {
	return []*regexp.Regexp{
		regexp.MustCompile(`(?m)\brm\s+-[a-zA-Z]*r[a-zA-Z]*f[a-zA-Z]*\s+/(?:\s|$)`),
		regexp.MustCompile(`(?m)\bsudo\s+rm\s+-[a-zA-Z]*r[a-zA-Z]*f`),
		regexp.MustCompile(`(?m)\bchmod\s+-R\s+777\s+/`),
		regexp.MustCompile(`(?m)\bcurl\b.*\s-k(?:\s|$)`),
		regexp.MustCompile(`(?m)\bcurl\b.*--insecure`),
	}
}
