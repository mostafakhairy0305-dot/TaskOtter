package dependency

import (
	"fmt"
	"sort"
)

type CycleError struct {
	Path []string
}

func (e *CycleError) Error() string {
	return fmt.Sprintf("dependency cycle detected: %s", stringsJoinArrow(e.Path))
}

type MissingDependencyError struct {
	Module     string
	Dependency string
}

func (e *MissingDependencyError) Error() string {
	return fmt.Sprintf("module %q depends on missing module %q", e.Module, e.Dependency)
}

func stringsJoinArrow(parts []string) string {
	if len(parts) == 0 {
		return ""
	}
	out := parts[0]
	for i := 1; i < len(parts); i++ {
		out += " -> " + parts[i]
	}
	return out
}

func ResolveTransitive(requested []string, deps map[string][]string) ([]string, error) {
	needed := make(map[string]struct{})
	var visit func(module string, stack []string) error
	visit = func(module string, stack []string) error {
		if _, ok := deps[module]; !ok {
			return fmt.Errorf("module %q is not defined in .deps.yml", module)
		}
		for i, s := range stack {
			if s == module {
				cycle := append(append([]string{}, stack[i:]...), module)
				return &CycleError{Path: cycle}
			}
		}
		if _, ok := needed[module]; ok {
			return nil
		}
		needed[module] = struct{}{}
		for _, dep := range deps[module] {
			if _, ok := deps[dep]; !ok {
				return &MissingDependencyError{Module: module, Dependency: dep}
			}
			if err := visit(dep, append(stack, module)); err != nil {
				return err
			}
		}
		return nil
	}

	for _, module := range requested {
		if err := visit(module, nil); err != nil {
			return nil, err
		}
	}

	requestedSet := make(map[string]struct{}, len(requested))
	for _, m := range requested {
		requestedSet[m] = struct{}{}
	}

	var dependencies []string
	for module := range needed {
		if _, ok := requestedSet[module]; ok {
			continue
		}
		dependencies = append(dependencies, module)
	}
	sort.Strings(dependencies)
	return dependencies, nil
}
