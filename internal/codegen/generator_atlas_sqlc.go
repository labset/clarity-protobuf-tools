package codegen

import (
	"bytes"
	"embed"
	"fmt"
	"path"
	"text/template"

	"google.golang.org/protobuf/compiler/protogen"
)

//go:embed templates/atlas-sqlc/*.tmpl
var atlasSqlcTemplateFS embed.FS

var atlasSqlcTemplates = template.Must(template.ParseFS(atlasSqlcTemplateFS, "templates/atlas-sqlc/*.tmpl"))

type atlasSqlcGenerator struct {
	sqlc *sqlcGenerator
}

func (g *atlasSqlcGenerator) Generate(plugin *protogen.Plugin) error {
	if err := g.sqlc.Generate(plugin); err != nil {
		return err
	}

	// Collect packages that produced output.
	pkgMap := make(map[string]*packageEntities)
	var pkgOrder []string

	for _, file := range plugin.Files {
		if !file.Generate {
			continue
		}
		if path.Base(file.Desc.Path()) != "models.proto" {
			continue
		}

		pkg := string(file.Desc.Package())
		pe, ok := pkgMap[pkg]
		if !ok {
			meta, err := parsePackage(pkg)
			if err != nil {
				return err
			}
			pe = &packageEntities{meta: meta}
			pkgMap[pkg] = pe
			pkgOrder = append(pkgOrder, pkg)
		}

		for _, msg := range file.Messages {
			if isEntityMessage(msg) {
				pe.messages = append(pe.messages, msg)
			}
		}
	}

	for _, pkg := range pkgOrder {
		pe := pkgMap[pkg]
		if len(pe.messages) == 0 {
			continue
		}

		outDir := pe.meta.outputDir()
		if g.sqlc.outputDir != "" {
			outDir = fmt.Sprintf("%s/%s", g.sqlc.outputDir, outDir)
		}

		atlasContent, err := renderAtlasConfig(pe.meta)
		if err != nil {
			return err
		}
		if _, err := plugin.NewGeneratedFile(fmt.Sprintf("%s/atlas.hcl", outDir), "").Write([]byte(atlasContent)); err != nil {
			return err
		}

		baselineContent, err := renderBaseline()
		if err != nil {
			return err
		}
		if _, err := plugin.NewGeneratedFile(fmt.Sprintf("%s/sql/baseline.sql", outDir), "").Write([]byte(baselineContent)); err != nil {
			return err
		}
	}

	return nil
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
