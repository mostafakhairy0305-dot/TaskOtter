// Package variants resolves Node task source module names from runtime configuration.
package variants

import (
	"errors"
	"fmt"
	"slices"

	"github.com/mostafakhairy0305-dot/TaskOtter/internal/config"
)

var (
	errVersionManagerRequired = errors.New("js.version-manager required for package manager")
	errInvalidPackageManager  = errors.New("invalid package manager")
)

func nodeToolSuffixes() []string {
	return []string{
		"npm-fnm", "npm-nvm",
		"yarn-fnm", "yarn-nvm",
		"pnpm-fnm", "pnpm-nvm",
		"bun",
	}
}

func stripSuffixes() []string {
	return []string{
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
}

// IsNodeToolVariant reports whether moduleName is a Node variant of logicalTask.
func IsNodeToolVariant(moduleName, logicalTask string) bool {
	prefix := logicalTask + "-"
	if len(moduleName) <= len(prefix) {
		return false
	}

	if moduleName[:len(prefix)] != prefix {
		return false
	}

	suffix := moduleName[len(prefix):]

	return slices.Contains(nodeToolSuffixes(), suffix)
}

// BuildSourceModule constructs the store module name for a logical task and JS configuration.
func BuildSourceModule(
	task string,
	packageManager config.PackageManager,
	versionManager config.VersionManager,
) (string, error) {
	switch packageManager {
	case config.PMBun:
		return task + "-bun", nil
	case config.PMNPM, config.PMYarn, config.PMPnpm:
		if versionManager == "" {
			return "", fmt.Errorf("%w %q", errVersionManagerRequired, packageManager)
		}

		return fmt.Sprintf("%s-%s-%s", task, packageManager, versionManager), nil
	default:
		return "", fmt.Errorf("%w %q", errInvalidPackageManager, packageManager)
	}
}

// StripOneSuffix removes one known suffix from the end of name.
func StripOneSuffix(name string) (string, bool) {
	for _, suffix := range stripSuffixes() {
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
