package normalizer

import (
	"fmt"
	"sort"

	"github.com/mostafakhairy0305-dot/taskotter-sync-action/internal/variants"
)

type CollisionError struct {
	SourceA     string
	SourceB     string
	Destination string
}

func (e *CollisionError) Error() string {
	return fmt.Sprintf(
		`Destination collision: %q and %q both normalize to destination module %q.`,
		e.SourceA, e.SourceB, e.Destination,
	)
}

func Normalize(source string) (string, error) {
	current := source
	for {
		next, changed := variants.StripOneSuffix(current)
		if !changed {
			break
		}
		current = next
	}
	if current == "" {
		return "", fmt.Errorf("normalized name for %q is empty", source)
	}
	return current, nil
}

type Mapping struct {
	Source      string
	Destination string
}

func BuildDestinationMap(sources []string) (map[string]string, error) {
	destToSource := make(map[string]string, len(sources))
	result := make(map[string]string, len(sources))

	for _, source := range sources {
		dest, err := Normalize(source)
		if err != nil {
			return nil, err
		}
		if existing, ok := destToSource[dest]; ok && existing != source {
			return nil, &CollisionError{SourceA: existing, SourceB: source, Destination: dest}
		}
		destToSource[dest] = source
		result[source] = dest
	}
	return result, nil
}

func SortedSources(m map[string]string) []string {
	out := make([]string, 0, len(m))
	for source := range m {
		out = append(out, source)
	}
	sort.Strings(out)
	return out
}
