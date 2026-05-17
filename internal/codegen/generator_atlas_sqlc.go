package codegen

import (
	"bytes"
	"embed"
	"fmt"
	"os"
	"text/template"

	"google.golang.org/protobuf/compiler/protogen"
)

//go:embed templates/atlas-sqlc/*.tmpl
var atlasSqlcTemplateFS embed.FS

var atlasSqlcTemplates = template.Must(
	template.ParseFS(atlasSqlcTemplateFS, "templates/atlas-sqlc/*.tmpl"),
)

type atlasSqlcGenerator struct {
	sqlc *sqlcGenerator
}

func (g *atlasSqlcGenerator) Generate(plugin *protogen.Plugin) error {
	if err := g.sqlc.Generate(plugin); err != nil {
		return err
	}

	packages, err := collectPackageEntities(plugin)
	if err != nil {
		return err
	}

	for _, pe := range packages {
		outDir := pe.meta.outputDir()
		if g.sqlc.outputDir != "" {
			outDir = fmt.Sprintf("%s/%s", g.sqlc.outputDir, outDir)
		}

		atlasPath := fmt.Sprintf("%s/atlas.hcl", outDir)
		if !fileExists(atlasPath) {
			atlasContent, err := renderAtlasConfig(pe.meta)
			if err != nil {
				return err
			}
			if _, err := plugin.NewGeneratedFile(atlasPath, "").
				Write([]byte(atlasContent)); err != nil {
				return err
			}
		}

		baselinePath := fmt.Sprintf("%s/sql/baseline.sql", outDir)
		if !fileExists(baselinePath) {
			baselineContent, err := renderBaseline()
			if err != nil {
				return err
			}
			if _, err := plugin.NewGeneratedFile(baselinePath, "").
				Write([]byte(baselineContent)); err != nil {
				return err
			}
		}
	}

	return nil
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func renderAtlasConfig(meta packageMeta) (string, error) {
	var buf bytes.Buffer
	if err := atlasSqlcTemplates.ExecuteTemplate(&buf, "atlas.hcl.tmpl", meta); err != nil {
		return "", fmt.Errorf("executing atlas.hcl template: %w", err)
	}
	return buf.String(), nil
}

func renderBaseline() (string, error) {
	var buf bytes.Buffer
	if err := atlasSqlcTemplates.ExecuteTemplate(&buf, "baseline.sql.tmpl", nil); err != nil {
		return "", fmt.Errorf("executing baseline.sql template: %w", err)
	}
	return buf.String(), nil
}
