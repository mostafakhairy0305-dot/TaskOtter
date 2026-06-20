package resolver

import (
	"fmt"
	"sort"
	"strings"

	"github.com/mostafakhairy0305-dot/taskotter-sync-action/internal/config"
	"github.com/mostafakhairy0305-dot/taskotter-sync-action/internal/variants"
)

type Resolution struct {
	LogicalTask  string
	SourceModule string
}

type ResolveError struct {
	LogicalTask  string
	Attempted    string
	Message      string
	CloseMatches []string
}

func (e *ResolveError) Error() string {
	msg := fmt.Sprintf(`task %q`, e.LogicalTask)
	if e.Attempted != "" {
		msg += fmt.Sprintf(" (attempted source module %q)", e.Attempted)
	}
	msg += ": " + e.Message
	if len(e.CloseMatches) > 0 {
		msg += fmt.Sprintf("; close matches: %s", strings.Join(e.CloseMatches, ", "))
	}
	return msg
}

func ResolveAll(tasks []string, catalog map[string]struct{}, pm config.PackageManager, vm config.VersionManager) ([]Resolution, error) {
	var out []Resolution
	for _, task := range tasks {
		res, err := Resolve(task, catalog, pm, vm)
		if err != nil {
			return nil, err
		}
		out = append(out, res)
	}
	return out, nil
}

func Resolve(task string, catalog map[string]struct{}, pm config.PackageManager, vm config.VersionManager) (Resolution, error) {
	if _, ok := catalog[task]; ok {
		return Resolution{LogicalTask: task, SourceModule: task}, nil
	}

	nodeVariants := findVariants(task, catalog)
	if len(nodeVariants) == 0 {
		return Resolution{}, &ResolveError{
			LogicalTask:  task,
			Message:      "task not found in store",
			CloseMatches: closeMatches(task, catalogKeys(catalog), 5),
		}
	}

	if pm == "" {
		return Resolution{}, &ResolveError{
			LogicalTask: task,
			Message:     fmt.Sprintf(`Task %q requires node-package-manager. Select npm, yarn, pnpm, or bun.`, task),
		}
	}

	attempted, err := variants.BuildSourceModule(task, pm, vm)
	if err != nil {
		return Resolution{}, &ResolveError{LogicalTask: task, Message: err.Error()}
	}

	if _, ok := catalog[attempted]; !ok {
		return Resolution{}, &ResolveError{
			LogicalTask:  task,
			Attempted:    attempted,
			Message:      "source module not found in store",
			CloseMatches: closeMatches(attempted, catalogKeys(catalog), 5),
		}
	}

	return Resolution{LogicalTask: task, SourceModule: attempted}, nil
}

func findVariants(task string, catalog map[string]struct{}) []string {
	var out []string
	for name := range catalog {
		if variants.IsNodeToolVariant(name, task) {
			out = append(out, name)
		}
	}
	sort.Strings(out)
	return out
}

func catalogKeys(catalog map[string]struct{}) []string {
	keys := make([]string, 0, len(catalog))
	for k := range catalog {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

func closeMatches(query string, candidates []string, limit int) []string {
	type scored struct {
		name  string
		score int
	}
	var scores []scored
	for _, c := range candidates {
		score := similarity(query, c)
		if score > 0 {
			scores = append(scores, scored{name: c, score: score})
		}
	}
	sort.Slice(scores, func(i, j int) bool {
		if scores[i].score == scores[j].score {
			return scores[i].name < scores[j].name
		}
		return scores[i].score > scores[j].score
	})
	var out []string
	for i := 0; i < len(scores) && i < limit; i++ {
		out = append(out, scores[i].name)
	}
	return out
}

func similarity(a, b string) int {
	if a == b {
		return 1000
	}
	if strings.HasPrefix(b, a) || strings.HasPrefix(a, b) {
		return 500 + minInt(len(a), len(b))
	}
	return levenshtein(a, b)
}

func levenshtein(a, b string) int {
	if a == b {
		return 100
	}
	la, lb := len(a), len(b)
	if la == 0 || lb == 0 {
		return 0
	}
	prev := make([]int, lb+1)
	curr := make([]int, lb+1)
	for j := 0; j <= lb; j++ {
		prev[j] = j
	}
	for i := 1; i <= la; i++ {
		curr[0] = i
		for j := 1; j <= lb; j++ {
			cost := 1
			if a[i-1] == b[j-1] {
				cost = 0
			}
			curr[j] = min(curr[j-1]+1, prev[j]+1, prev[j-1]+cost)
		}
		prev, curr = curr, prev
	}
	dist := prev[lb]
	maxLen := max(la, lb)
	return max(0, 100-(dist*100/maxLen))
}

func min(a, b, c int) int {
	m := a
	if b < m {
		m = b
	}
	if c < m {
		m = c
	}
	return m
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
