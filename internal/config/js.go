package config

import (
	"fmt"
	"strings"

	"gopkg.in/yaml.v3"
)

type JSRuntime string

const (
	JSRuntimeBun    JSRuntime = "bun"
	JSRuntimeNodeJS JSRuntime = "nodejs"
)

type jsInput struct {
	Runtime        string `yaml:"runtime"`
	PackageManager string `yaml:"package-manager"`
	VersionManager string `yaml:"version-manager"`
}

type jsConfig struct {
	Runtime            JSRuntime
	NodePackageManager PackageManager
	NodeVersionManager VersionManager
}

func parseJS(raw string) (*jsConfig, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil, nil
	}

	var in jsInput
	if err := yaml.Unmarshal([]byte(raw), &in); err != nil {
		return nil, &ValidationError{
			Field:   "js",
			Message: fmt.Sprintf("invalid YAML: %v", err),
		}
	}

	runtime := strings.TrimSpace(in.Runtime)
	if runtime == "" {
		runtime = string(JSRuntimeNodeJS)
	}

	switch JSRuntime(runtime) {
	case JSRuntimeBun:
		if strings.TrimSpace(in.PackageManager) != "" {
			return nil, &ValidationError{
				Field:   "js.package-manager",
				Message: "is only valid when js.runtime is nodejs",
			}
		}
		if strings.TrimSpace(in.VersionManager) != "" {
			return nil, &ValidationError{
				Field:   "js.version-manager",
				Message: "is only valid when js.runtime is nodejs",
			}
		}
		return &jsConfig{
			Runtime:            JSRuntimeBun,
			NodePackageManager: PMBun,
		}, nil

	case JSRuntimeNodeJS:
		pmRaw := strings.TrimSpace(in.PackageManager)
		if pmRaw == "" {
			pmRaw = string(PMNPM)
		}
		vmRaw := strings.TrimSpace(in.VersionManager)
		if vmRaw == "" {
			vmRaw = string(VMFnm)
		}

		pm, err := parseNodePackageManager(pmRaw)
		if err != nil {
			return nil, err
		}
		if pm == PMBun {
			return nil, &ValidationError{
				Field:   "js.package-manager",
				Message: `use js.runtime "bun" instead of package-manager "bun"`,
			}
		}

		vm, err := parseNodeVersionManager(vmRaw)
		if err != nil {
			return nil, err
		}

		return &jsConfig{
			Runtime:            JSRuntimeNodeJS,
			NodePackageManager: pm,
			NodeVersionManager: vm,
		}, nil

	default:
		return nil, &ValidationError{
			Field:   "js.runtime",
			Message: fmt.Sprintf("invalid value %q: allowed values are bun or nodejs", runtime),
		}
	}
}

func parseNodePackageManager(raw string) (PackageManager, error) {
	switch raw {
	case "npm", "yarn", "pnpm":
		return PackageManager(raw), nil
	default:
		return "", &ValidationError{
			Field:   "js.package-manager",
			Message: fmt.Sprintf("invalid value %q: allowed values are npm, yarn, or pnpm", raw),
		}
	}
}

func parseNodeVersionManager(raw string) (VersionManager, error) {
	switch raw {
	case "fnm", "nvm":
		return VersionManager(raw), nil
	default:
		return "", &ValidationError{
			Field:   "js.version-manager",
			Message: fmt.Sprintf("invalid value %q: allowed values are fnm or nvm", raw),
		}
	}
}
