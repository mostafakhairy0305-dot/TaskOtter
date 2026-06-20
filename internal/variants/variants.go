package variants

import (
	"fmt"

	"github.com/mostafakhairy0305-dot/taskotter-sync-action/internal/config"
)

// NodeToolSuffixes identifies node tool variant suffixes (without leading task name).
var NodeToolSuffixes = []string{
	"npm-fnm", "npm-nvm",
	"yarn-fnm", "yarn-nvm",
	"pnpm-fnm", "pnpm-nvm",
	"bun",
}

// StripSuffixes are suffixes removed from the end of source module names (longest first).
var StripSuffixes = []string{
	"-npm-fnm",
	"-npm-nvm",
	"-yarn-fnm",
	"-yarn-nvm",
	"-pnpm-fnm",
	"-pnpm-nvm",
	"-bun",
	"-fnm",
	"-nvm",
}

func IsNodeToolVariant(moduleName, logicalTask string) bool {
	prefix := logicalTask + "-"
	if len(moduleName) <= len(prefix) {
		return false
	}
	if moduleName[:len(prefix)] != prefix {
		return false
	}
	suffix := moduleName[len(prefix):]
	for _, vs := range NodeToolSuffixes {
		if suffix == vs {
			return true
		}
	}
	return false
}

func BuildSourceModule(task string, pm config.PackageManager, vm config.VersionManager) (string, error) {
	switch pm {
	case config.PMBun:
		return task + "-bun", nil
	case config.PMNPM, config.PMYarn, config.PMPnpm:
		if vm == "" {
			return "", fmt.Errorf("node-version-manager required for package manager %q", pm)
		}
		return fmt.Sprintf("%s-%s-%s", task, pm, vm), nil
	default:
		return "", fmt.Errorf("invalid package manager %q", pm)
	}
}

func StripOneSuffix(name string) (string, bool) {
	for _, suffix := range StripSuffixes {
		if len(name) > len(suffix) && name[len(name)-len(suffix):] == suffix {
			stripped := name[:len(name)-len(suffix)]
			if stripped == "" {
				continue
			}
			return stripped, true
		}
	}
	return name, false
}
