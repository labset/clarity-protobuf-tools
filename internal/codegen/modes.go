package codegen

import (
	"fmt"
	"strings"
)

func GeneratorForMode(raw string) (Generator, error) {
	p := parseParams(raw)
	switch p.mode {
	case "sqlc":
		return &sqlcGenerator{outputDir: p.outputDir}, nil
	case "atlas-sqlc":
		return &atlasSqlcGenerator{sqlc: &sqlcGenerator{outputDir: p.outputDir}}, nil
	case "service":
		return &serviceGenerator{outputDir: p.outputDir}, nil
	case "connect-crud":
		if p.goModule == "" {
			return nil, fmt.Errorf("connect-crud mode requires go_module parameter")
		}
		return &connectCrudGenerator{
			atlasSqlc: &atlasSqlcGenerator{sqlc: &sqlcGenerator{outputDir: p.outputDir}},
			goModule:  p.goModule,
		}, nil
	default:
		return nil, fmt.Errorf("unknown mode %q", p.mode)
	}
}

// params holds parsed plugin parameters.
type params struct {
	mode      string
	outputDir string
	goModule  string
}

func parseParams(raw string) params {
	p := params{}

	for _, param := range strings.Split(raw, ",") {
		key, value, ok := strings.Cut(param, "=")
		if !ok {
			continue
		}

		switch strings.TrimSpace(key) {
		case "mode":
			p.mode = strings.TrimSpace(value)
		case "output_dir":
			p.outputDir = strings.TrimSpace(value)
		case "go_module":
			p.goModule = strings.TrimSpace(value)
		}
	}

	return p
}
