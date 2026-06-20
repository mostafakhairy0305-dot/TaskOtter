package repo

import (
	"fmt"
	"strings"
)

func Parse(full string) (owner, name string, err error) {
	parts := strings.Split(full, "/")
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return "", "", fmt.Errorf("invalid repository %q", full)
	}
	return parts[0], parts[1], nil
}
