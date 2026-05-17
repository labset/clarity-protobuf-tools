package codegen

import (
	"bytes"
	"embed"
	"fmt"
	"strings"
	"text/template"

	"github.com/labset/clarity-protobuf-tools/internal/clarity"
	"google.golang.org/protobuf/compiler/protogen"
)

//go:embed templates/connect-crud/*.tmpl
var connectCrudTemplateFS embed.FS

var connectCrudTemplates = template.Must(
	template.New("connect-crud").Funcs(template.FuncMap{
		"lower":     strings.ToLower,
		"snakeCase": toSnakeCase,
	}).ParseFS(connectCrudTemplateFS, "templates/connect-crud/*.tmpl"),
)

type connectCrudGenerator struct {
	outputDir string
	goModule  string
}

type handlerData struct {
	Package       string
	Model         string
	ModelSnake    string
	StoreImport   string
	ProtoImport   string
	ConnectImport string
	ConnectAlias  string
	ProtoAlias    string
}

func (g *connectCrudGenerator) Generate(plugin *protogen.Plugin) error {
	packages, err := collectServiceEntities(plugin)
	if err != nil {
		return err
	}

	for _, pe := range packages {
		outDir := fmt.Sprintf("%s/%s/%s/api", pe.meta.Provider, pe.meta.Domain, pe.meta.Version)
		if g.outputDir != "" {
			outDir = fmt.Sprintf("%s/%s", g.outputDir, outDir)
		}

		storeImport := fmt.Sprintf(
			"%s/internal/%s/%s/%s/db",
			g.goModule, pe.meta.Provider, pe.meta.Domain, pe.meta.Version,
		)

		protoImport, protoAlias, connectImport, connectAlias := deriveGoImports(pe.goPackage)

		for _, msg := range pe.messages {
			ops := clarity.Operations(msg.Desc)
			if len(ops) == 0 {
				continue
			}

			modelName := string(msg.Desc.Name())
			modelSnake := toSnakeCase(modelName)

			data := handlerData{
				Package:       "api",
				Model:         modelName,
				ModelSnake:    modelSnake,
				StoreImport:   storeImport,
				ProtoImport:   protoImport,
				ConnectImport: connectImport,
				ConnectAlias:  connectAlias,
				ProtoAlias:    protoAlias,
			}

			content, err := renderHandler(data)
			if err != nil {
				return err
			}
			filePath := fmt.Sprintf("%s/handler_%s.go", outDir, modelSnake)
			if _, err := plugin.NewGeneratedFile(filePath, "").Write([]byte(content)); err != nil {
				return err
			}
		}
	}

	return nil
}

// deriveGoImports extracts proto and connect import paths from a go_package option.
// Given "github.com/acme/inventory/v1;inventoryv1", it returns:
//   - protoImport:   "github.com/acme/inventory/v1"
//   - protoAlias:    "inventoryv1"
//   - connectImport: "github.com/acme/inventory/v1/inventoryv1connect"
//   - connectAlias:  "inventoryv1connect"
func deriveGoImports(goPackage string) (protoImport, protoAlias, connectImport, connectAlias string) {
	importPath, alias, hasSemicolon := strings.Cut(goPackage, ";")
	if hasSemicolon {
		protoImport = importPath
		protoAlias = alias
	} else {
		protoImport = importPath
		parts := strings.Split(importPath, "/")
		protoAlias = parts[len(parts)-1]
	}
	connectAlias = protoAlias + "connect"
	connectImport = protoImport + "/" + connectAlias
	return
}

func renderHandler(data handlerData) (string, error) {
	var buf bytes.Buffer
	if err := connectCrudTemplates.ExecuteTemplate(&buf, "handler.go.tmpl", data); err != nil {
		return "", fmt.Errorf("executing handler template: %w", err)
	}
	return buf.String(), nil
}
