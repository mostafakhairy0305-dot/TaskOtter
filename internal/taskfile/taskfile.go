package taskfile

import (
	"fmt"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

const rootTemplate = "version: \"3\"\n"

// NewRootTemplate returns the minimal root Taskfile used when none exists yet.
func NewRootTemplate() []byte {
	return []byte(rootTemplate)
}

type RewriteError struct {
	Message string
}

func (e *RewriteError) Error() string {
	return e.Message
}

func RewriteIncludes(content []byte, sourceToDest map[string]string) ([]byte, error) {
	var node yaml.Node
	if err := yaml.Unmarshal(content, &node); err != nil {
		return nil, &RewriteError{Message: fmt.Sprintf("parse Taskfile YAML: %v", err)}
	}
	if len(node.Content) == 0 {
		return nil, &RewriteError{Message: "empty Taskfile YAML"}
	}
	root := node.Content[0]
	includesNode := findMappingValue(root, "includes")
	if includesNode != nil {
		if err := rewriteIncludesNode(includesNode, sourceToDest); err != nil {
			return nil, err
		}
	}
	out, err := yaml.Marshal(&node)
	if err != nil {
		return nil, &RewriteError{Message: fmt.Sprintf("marshal Taskfile YAML: %v", err)}
	}
	if err := yaml.Unmarshal(out, &yaml.Node{}); err != nil {
		return nil, &RewriteError{Message: fmt.Sprintf("validate rewritten Taskfile YAML: %v", err)}
	}
	return out, nil
}

func rewriteIncludesNode(includes *yaml.Node, sourceToDest map[string]string) error {
	if includes.Kind != yaml.MappingNode {
		return nil
	}
	for i := 0; i < len(includes.Content); i += 2 {
		entry := includes.Content[i+1]
		if entry.Kind != yaml.MappingNode {
			continue
		}
		taskfileNode := findMappingValue(entry, "taskfile")
		if taskfileNode == nil || taskfileNode.Kind != yaml.ScalarNode {
			continue
		}
		rewritten, err := rewriteIncludePath(taskfileNode.Value, sourceToDest)
		if err != nil {
			return err
		}
		taskfileNode.Value = rewritten
	}
	return nil
}

func rewriteIncludePath(path string, sourceToDest map[string]string) (string, error) {
	normalized := filepath.ToSlash(path)
	if !strings.HasSuffix(normalized, "/Taskfile.yml") {
		return path, nil
	}
	dir := strings.TrimSuffix(normalized, "/Taskfile.yml")
	dir = strings.TrimPrefix(dir, "./")
	dir = strings.TrimPrefix(dir, "../")
	if dir == "" || strings.Contains(dir, "/") {
		return path, nil
	}
	dest, ok := sourceToDest[dir]
	if !ok {
		return path, nil
	}
	prefix := ""
	if strings.HasPrefix(normalized, "../") {
		prefix = "../"
	} else if strings.HasPrefix(normalized, "./") {
		prefix = "./"
	}
	return prefix + dest + "/Taskfile.yml", nil
}

func findMappingValue(mapNode *yaml.Node, key string) *yaml.Node {
	if mapNode == nil || mapNode.Kind != yaml.MappingNode {
		return nil
	}
	for i := 0; i < len(mapNode.Content); i += 2 {
		k := mapNode.Content[i]
		if k.Kind == yaml.ScalarNode && k.Value == key {
			return mapNode.Content[i+1]
		}
	}
	return nil
}

type RootUpdateInput struct {
	Tasks           []string
	TargetFolder    string
	DestByTask      map[string]string
	ManagedTasks    []string
	ModuleTaskfiles map[string][]byte
}

func UpdateRootTaskfile(content []byte, in RootUpdateInput) ([]byte, error) {
	var node yaml.Node
	if err := yaml.Unmarshal(content, &node); err != nil {
		return nil, &RewriteError{Message: fmt.Sprintf("parse root Taskfile YAML: %v", err)}
	}
	if len(node.Content) == 0 {
		return nil, &RewriteError{Message: "empty root Taskfile YAML"}
	}
	root := node.Content[0]

	managedSet := make(map[string]struct{}, len(in.Tasks))
	for _, task := range in.Tasks {
		managedSet[task] = struct{}{}
	}

	includesNode := findMappingValue(root, "includes")
	if includesNode == nil {
		includesNode = &yaml.Node{Kind: yaml.MappingNode}
		appendMappingPair(root, scalar("includes"), includesNode)
	}
	if includesNode.Kind != yaml.MappingNode {
		return nil, &RewriteError{Message: "root Taskfile includes must be a mapping"}
	}

	existing := make(map[string]*yaml.Node)
	for i := 0; i < len(includesNode.Content); i += 2 {
		keyNode := includesNode.Content[i]
		existing[keyNode.Value] = includesNode.Content[i+1]
	}

	for alias := range existing {
		if _, managed := managedSet[alias]; managed {
			continue
		}
		if containsString(in.ManagedTasks, alias) {
			deleteMappingKey(includesNode, alias)
		}
	}

	for _, task := range in.Tasks {
		dest, ok := in.DestByTask[task]
		if !ok {
			return nil, &RewriteError{Message: fmt.Sprintf("missing destination for task %q", task)}
		}
		path := filepath.ToSlash(filepath.Join(in.TargetFolder, dest, "Taskfile.yml"))
		moduleVars, err := extractVarsNode(in.ModuleTaskfiles[task])
		if err != nil {
			return nil, err
		}
		if entry, ok := existing[task]; ok {
			if !isManagedInclude(entry, path, in.ManagedTasks, task) {
				return nil, &RewriteError{Message: fmt.Sprintf("include alias %q already exists and is not managed by TaskOtter", task)}
			}
			setIncludePath(entry, path)
			mergeIncludeVars(entry, moduleVars)
			continue
		}
		entry := newIncludeEntry(path, moduleVars)
		appendMappingPair(includesNode, scalar(task), entry)
	}

	out, err := yaml.Marshal(&node)
	if err != nil {
		return nil, err
	}
	if err := yaml.Unmarshal(out, &yaml.Node{}); err != nil {
		return nil, &RewriteError{Message: fmt.Sprintf("validate root Taskfile YAML: %v", err)}
	}
	return out, nil
}

func isManagedInclude(entry *yaml.Node, expectedPath string, managedTasks []string, task string) bool {
	taskfileNode := findMappingValue(entry, "taskfile")
	if taskfileNode != nil {
		return taskfileNode.Value == expectedPath
	}
	if entry.Kind == yaml.ScalarNode {
		return entry.Value == expectedPath
	}
	return containsString(managedTasks, task)
}

func setIncludePath(entry *yaml.Node, path string) {
	taskfileNode := findMappingValue(entry, "taskfile")
	if taskfileNode == nil {
		appendMappingPair(entry, scalar("taskfile"), scalar(path))
		return
	}
	taskfileNode.Value = path
}

func extractVarsNode(content []byte) (*yaml.Node, error) {
	if len(content) == 0 {
		return nil, nil
	}
	var node yaml.Node
	if err := yaml.Unmarshal(content, &node); err != nil {
		return nil, &RewriteError{Message: fmt.Sprintf("parse module Taskfile YAML: %v", err)}
	}
	if len(node.Content) == 0 {
		return nil, nil
	}
	varsNode := findMappingValue(node.Content[0], "vars")
	if varsNode == nil || varsNode.Kind != yaml.MappingNode || len(varsNode.Content) == 0 {
		return nil, nil
	}
	return cloneYAMLNode(varsNode), nil
}

func mergeIncludeVars(entry *yaml.Node, moduleVars *yaml.Node) {
	if moduleVars == nil || moduleVars.Kind != yaml.MappingNode {
		return
	}
	existingVars := findMappingValue(entry, "vars")
	if existingVars == nil {
		appendMappingPair(entry, scalar("vars"), moduleVars)
		return
	}
	if existingVars.Kind != yaml.MappingNode {
		return
	}
	existingKeys := make(map[string]struct{}, len(existingVars.Content)/2)
	for i := 0; i < len(existingVars.Content); i += 2 {
		existingKeys[existingVars.Content[i].Value] = struct{}{}
	}
	for i := 0; i < len(moduleVars.Content); i += 2 {
		key := moduleVars.Content[i].Value
		if _, ok := existingKeys[key]; ok {
			continue
		}
		appendMappingPair(existingVars, cloneYAMLNode(moduleVars.Content[i]), cloneYAMLNode(moduleVars.Content[i+1]))
	}
}

func cloneYAMLNode(node *yaml.Node) *yaml.Node {
	if node == nil {
		return nil
	}
	data, err := yaml.Marshal(node)
	if err != nil {
		return nil
	}
	var out yaml.Node
	if err := yaml.Unmarshal(data, &out); err != nil || len(out.Content) == 0 {
		return nil
	}
	return out.Content[0]
}

func newIncludeEntry(path string, moduleVars *yaml.Node) *yaml.Node {
	entry := &yaml.Node{Kind: yaml.MappingNode}
	appendMappingPair(entry, scalar("taskfile"), scalar(path))
	if moduleVars != nil {
		appendMappingPair(entry, scalar("vars"), moduleVars)
	}
	return entry
}

func scalar(value string) *yaml.Node {
	return &yaml.Node{Kind: yaml.ScalarNode, Value: value}
}

func appendMappingPair(mapNode *yaml.Node, key, value *yaml.Node) {
	mapNode.Content = append(mapNode.Content, key, value)
}

func deleteMappingKey(mapNode *yaml.Node, key string) {
	for i := 0; i < len(mapNode.Content); i += 2 {
		if mapNode.Content[i].Value == key {
			mapNode.Content = append(mapNode.Content[:i], mapNode.Content[i+2:]...)
			return
		}
	}
}

func containsString(list []string, target string) bool {
	for _, v := range list {
		if v == target {
			return true
		}
	}
	return false
}
