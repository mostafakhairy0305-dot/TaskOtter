// Package taskfile rewrites root and module Taskfile YAML during sync.
package taskfile

import (
	"errors"
	"fmt"
	"path/filepath"
	"slices"
	"strings"

	"gopkg.in/yaml.v3"
)

const (
	rootTemplate            = "version: \"3\"\n"
	yamlMappingPairKeyValue = 2
)

var errNoModuleVars = errors.New("module Taskfile has no vars")

// NewRootTemplate returns the minimal root Taskfile used when none exists yet.
func NewRootTemplate() []byte {
	return []byte(rootTemplate)
}

// RewriteError reports Taskfile YAML rewrite failures.
type RewriteError struct {
	Message string
}

func (e *RewriteError) Error() string {
	return e.Message
}

// RewriteIncludes updates include taskfile paths using sourceToDest mappings.
func RewriteIncludes(content []byte, sourceToDest map[string]string) ([]byte, error) {
	var node yaml.Node

	err := yaml.Unmarshal(content, &node)
	if err != nil {
		return nil, &RewriteError{Message: fmt.Sprintf("parse Taskfile YAML: %v", err)}
	}

	if len(node.Content) == 0 {
		return nil, &RewriteError{Message: "empty Taskfile YAML"}
	}

	root := node.Content[0]

	includesNode := findMappingValue(root, "includes")
	if includesNode != nil {
		rewriteIncludesNode(includesNode, sourceToDest)
	}

	out, err := yaml.Marshal(&node)
	if err != nil {
		return nil, &RewriteError{Message: fmt.Sprintf("marshal Taskfile YAML: %v", err)}
	}

	var validateNode yaml.Node

	err = yaml.Unmarshal(out, &validateNode)
	if err != nil {
		return nil, &RewriteError{Message: fmt.Sprintf("validate rewritten Taskfile YAML: %v", err)}
	}

	return out, nil
}

func rewriteIncludesNode(includes *yaml.Node, sourceToDest map[string]string) {
	if includes.Kind != yaml.MappingNode {
		return
	}

	for idx := 0; idx < len(includes.Content); idx += yamlMappingPairKeyValue {
		entry := includes.Content[idx+1]
		if entry.Kind != yaml.MappingNode {
			continue
		}

		taskfileNode := findMappingValue(entry, "taskfile")
		if taskfileNode == nil || taskfileNode.Kind != yaml.ScalarNode {
			continue
		}

		taskfileNode.Value = rewriteIncludePath(taskfileNode.Value, sourceToDest)
	}
}

func rewriteIncludePath(path string, sourceToDest map[string]string) string {
	normalized := filepath.ToSlash(path)
	if !strings.HasSuffix(normalized, "/Taskfile.yml") {
		return path
	}

	dir := strings.TrimSuffix(normalized, "/Taskfile.yml")
	dir = strings.TrimPrefix(dir, "./")

	dir = strings.TrimPrefix(dir, "../")
	if dir == "" || strings.Contains(dir, "/") {
		return path
	}

	dest, ok := sourceToDest[dir]
	if !ok {
		return path
	}

	prefix := ""
	if strings.HasPrefix(normalized, "../") {
		prefix = "../"
	} else if strings.HasPrefix(normalized, "./") {
		prefix = "./"
	}

	return prefix + dest + "/Taskfile.yml"
}

func findMappingValue(mapNode *yaml.Node, key string) *yaml.Node {
	if mapNode == nil || mapNode.Kind != yaml.MappingNode {
		return nil
	}

	for idx := 0; idx < len(mapNode.Content); idx += yamlMappingPairKeyValue {
		keyNode := mapNode.Content[idx]
		if keyNode.Kind == yaml.ScalarNode && keyNode.Value == key {
			return mapNode.Content[idx+1]
		}
	}

	return nil
}

// RootUpdateInput carries data for updating the root Taskfile includes section.
type RootUpdateInput struct {
	Tasks           []string
	TargetFolder    string
	DestByTask      map[string]string
	ManagedTasks    []string
	ModuleTaskfiles map[string][]byte
}

// UpdateRootTaskfile merges managed module includes into the root Taskfile.
func UpdateRootTaskfile(content []byte, input RootUpdateInput) ([]byte, error) {
	var node yaml.Node

	err := yaml.Unmarshal(content, &node)
	if err != nil {
		return nil, &RewriteError{Message: fmt.Sprintf("parse root Taskfile YAML: %v", err)}
	}

	if len(node.Content) == 0 {
		return nil, &RewriteError{Message: "empty root Taskfile YAML"}
	}

	root := node.Content[0]

	includesNode, existing, err := prepareIncludesNode(root, input)
	if err != nil {
		return nil, err
	}

	err = upsertManagedIncludes(includesNode, input, existing)
	if err != nil {
		return nil, err
	}

	out, err := yaml.Marshal(&node)
	if err != nil {
		return nil, fmt.Errorf("marshal root Taskfile YAML: %w", err)
	}

	var validateNode yaml.Node

	err = yaml.Unmarshal(out, &validateNode)
	if err != nil {
		return nil, &RewriteError{Message: fmt.Sprintf("validate root Taskfile YAML: %v", err)}
	}

	return out, nil
}

func prepareIncludesNode(root *yaml.Node, input RootUpdateInput) (*yaml.Node, map[string]*yaml.Node, error) {
	managedSet := make(map[string]struct{}, len(input.Tasks))
	for _, task := range input.Tasks {
		managedSet[task] = struct{}{}
	}

	includesNode := findMappingValue(root, "includes")
	if includesNode == nil {
		includesNode = newYAMLMappingNode()
		appendMappingPair(root, yamlScalar("includes"), includesNode)
	}

	if includesNode.Kind != yaml.MappingNode {
		return nil, nil, &RewriteError{Message: "root Taskfile includes must be a mapping"}
	}

	existing := make(map[string]*yaml.Node)

	for idx := 0; idx < len(includesNode.Content); idx += yamlMappingPairKeyValue {
		keyNode := includesNode.Content[idx]
		existing[keyNode.Value] = includesNode.Content[idx+1]
	}

	pruneRemovedManagedIncludes(includesNode, existing, managedSet, input.ManagedTasks)

	return includesNode, existing, nil
}

func pruneRemovedManagedIncludes(
	includesNode *yaml.Node,
	existing map[string]*yaml.Node,
	managedSet map[string]struct{},
	managedTasks []string,
) {
	for alias := range existing {
		if _, managed := managedSet[alias]; managed {
			continue
		}

		if containsString(managedTasks, alias) {
			deleteMappingKey(includesNode, alias)
		}
	}
}

func upsertManagedIncludes(
	includesNode *yaml.Node,
	input RootUpdateInput,
	existing map[string]*yaml.Node,
) error {
	for _, task := range input.Tasks {
		dest, ok := input.DestByTask[task]
		if !ok {
			return &RewriteError{Message: fmt.Sprintf("missing destination for task %q", task)}
		}

		path := filepath.ToSlash(filepath.Join(input.TargetFolder, dest, "Taskfile.yml"))

		moduleVars, err := extractVarsNode(input.ModuleTaskfiles[task])
		if err != nil && !errors.Is(err, errNoModuleVars) {
			return err
		}

		if entry, ok := existing[task]; ok {
			if !isManagedInclude(entry, path, input.ManagedTasks, task) {
				return &RewriteError{
					Message: fmt.Sprintf(
						"include alias %q already exists and is not managed by TaskOtter",
						task,
					),
				}
			}

			setIncludePath(entry, path)
			mergeIncludeVars(entry, moduleVars)

			continue
		}

		entry := newIncludeEntry(path, moduleVars)
		appendMappingPair(includesNode, yamlScalar(task), entry)
	}

	return nil
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
		appendMappingPair(entry, yamlScalar("taskfile"), yamlScalar(path))

		return
	}

	taskfileNode.Value = path
}

func extractVarsNode(content []byte) (*yaml.Node, error) {
	if len(content) == 0 {
		return nil, errNoModuleVars
	}

	var node yaml.Node

	err := yaml.Unmarshal(content, &node)
	if err != nil {
		return nil, &RewriteError{Message: fmt.Sprintf("parse module Taskfile YAML: %v", err)}
	}

	if len(node.Content) == 0 {
		return nil, errNoModuleVars
	}

	varsNode := findMappingValue(node.Content[0], "vars")
	if varsNode == nil || varsNode.Kind != yaml.MappingNode || len(varsNode.Content) == 0 {
		return nil, errNoModuleVars
	}

	return cloneYAMLNode(varsNode), nil
}

func mergeIncludeVars(entry *yaml.Node, moduleVars *yaml.Node) {
	if moduleVars == nil || moduleVars.Kind != yaml.MappingNode {
		return
	}

	existingVars := findMappingValue(entry, "vars")
	if existingVars == nil {
		appendMappingPair(entry, yamlScalar("vars"), moduleVars)

		return
	}

	if existingVars.Kind != yaml.MappingNode {
		return
	}

	existingKeys := make(map[string]struct{}, len(existingVars.Content)/yamlMappingPairKeyValue)
	for idx := 0; idx < len(existingVars.Content); idx += yamlMappingPairKeyValue {
		existingKeys[existingVars.Content[idx].Value] = struct{}{}
	}

	for idx := 0; idx < len(moduleVars.Content); idx += yamlMappingPairKeyValue {
		key := moduleVars.Content[idx].Value
		if _, ok := existingKeys[key]; ok {
			continue
		}

		appendMappingPair(
			existingVars,
			cloneYAMLNode(moduleVars.Content[idx]),
			cloneYAMLNode(moduleVars.Content[idx+1]),
		)
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

	err = yaml.Unmarshal(data, &out)
	if err != nil || len(out.Content) == 0 {
		return nil
	}

	return out.Content[0]
}

func newIncludeEntry(path string, moduleVars *yaml.Node) *yaml.Node {
	entry := newYAMLMappingNode()
	appendMappingPair(entry, yamlScalar("taskfile"), yamlScalar(path))

	if moduleVars != nil {
		appendMappingPair(entry, yamlScalar("vars"), moduleVars)
	}

	return entry
}

func newYAMLMappingNode() *yaml.Node {
	return &yaml.Node{
		Kind:        yaml.MappingNode,
		Style:       0,
		Tag:         "",
		Value:       "",
		Anchor:      "",
		Alias:       nil,
		Content:     nil,
		HeadComment: "",
		LineComment: "",
		FootComment: "",
		Line:        0,
		Column:      0,
	}
}

func yamlScalar(value string) *yaml.Node {
	return &yaml.Node{
		Kind:        yaml.ScalarNode,
		Style:       0,
		Tag:         "",
		Value:       value,
		Anchor:      "",
		Alias:       nil,
		Content:     nil,
		HeadComment: "",
		LineComment: "",
		FootComment: "",
		Line:        0,
		Column:      0,
	}
}

func appendMappingPair(mapNode *yaml.Node, key, value *yaml.Node) {
	mapNode.Content = append(mapNode.Content, key, value)
}

func deleteMappingKey(mapNode *yaml.Node, key string) {
	for idx := 0; idx < len(mapNode.Content); idx += yamlMappingPairKeyValue {
		if mapNode.Content[idx].Value == key {
			mapNode.Content = append(mapNode.Content[:idx], mapNode.Content[idx+yamlMappingPairKeyValue:]...)

			return
		}
	}
}

func containsString(list []string, target string) bool {
	return slices.Contains(list, target)
}
