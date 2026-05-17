package codegen

import (
	"bytes"
	"embed"
	"fmt"
	"strings"
	"text/template"

	"unicode"

	"github.com/labset/clarity-protobuf-tools/internal/clarity"
	"google.golang.org/protobuf/compiler/protogen"
)

//go:embed templates/connect-crud/*.tmpl
var connectCrudTemplateFS embed.FS

var connectCrudTemplates = template.Must(
	template.New("connect-crud").Funcs(template.FuncMap{
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
	ModelLower    string
	ModelSnake    string
	StoreImport   string
	ConnectImport string
	ConnectAlias  string
}

type mapperData struct {
	Package     string
	Model       string
	ModelLower  string
	StoreImport string
	ProtoImport string
	ProtoAlias  string
	Fields      []mapperField
}

type mapperField struct {
	ProtoName string // PascalCase, e.g. "Name"
	SQLCName  string // PascalCase SQLC column name, e.g. "Name"
}

type rpcData struct {
	Package     string
	Model       string
	ModelLower  string
	StoreImport string
	ProtoImport string
	ProtoAlias  string
	Fields      []mapperField
}

var opTemplateMap = map[string]string{
	"OPERATION_CREATE": "rpc_create.go.tmpl",
	"OPERATION_GET":    "rpc_get.go.tmpl",
	"OPERATION_LIST":   "rpc_list.go.tmpl",
	"OPERATION_UPDATE": "rpc_update.go.tmpl",
	"OPERATION_DELETE": "rpc_delete.go.tmpl",
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
			modelLower := strings.ToLower(modelName[:1]) + modelName[1:]

			data := handlerData{
				Package:       "api",
				Model:         modelName,
				ModelLower:    modelLower,
				ModelSnake:    modelSnake,
				StoreImport:   storeImport,
				ConnectImport: connectImport,
				ConnectAlias:  connectAlias,
			}

			content, err := renderHandler(data)
			if err != nil {
				return err
			}
			filePath := fmt.Sprintf("%s/handler_%s.go", outDir, modelSnake)
			if _, err := plugin.NewGeneratedFile(filePath, "").Write([]byte(content)); err != nil {
				return err
			}

			fields := extractMapperFields(msg)
			mapData := mapperData{
				Package:     "api",
				Model:       modelName,
				ModelLower:  modelLower,
				StoreImport: storeImport,
				ProtoImport: protoImport,
				ProtoAlias:  protoAlias,
				Fields:      fields,
			}

			mapperContent, err := renderMapper(mapData)
			if err != nil {
				return err
			}
			mapperPath := fmt.Sprintf("%s/mapper_%s.go", outDir, modelSnake)
			if _, err := plugin.NewGeneratedFile(mapperPath, "").Write([]byte(mapperContent)); err != nil {
				return err
			}

			rpc := rpcData{
				Package:     "api",
				Model:       modelName,
				ModelLower:  modelLower,
				StoreImport: storeImport,
				ProtoImport: protoImport,
				ProtoAlias:  protoAlias,
				Fields:      fields,
			}

			for _, op := range ops {
				tmplName, ok := opTemplateMap[op.String()]
				if !ok {
					continue
				}
				rpcContent, err := renderRPC(tmplName, rpc)
				if err != nil {
					return err
				}
				opName := strings.ToLower(strings.TrimPrefix(op.String(), "OPERATION_"))
				rpcPath := fmt.Sprintf("%s/rpc_%s_%s.go", outDir, opName, modelSnake)
				if _, err := plugin.NewGeneratedFile(rpcPath, "").Write([]byte(rpcContent)); err != nil {
					return err
				}
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

func renderRPC(tmplName string, data rpcData) (string, error) {
	var buf bytes.Buffer
	if err := connectCrudTemplates.ExecuteTemplate(&buf, tmplName, data); err != nil {
		return "", fmt.Errorf("executing %s template: %w", tmplName, err)
	}
	return buf.String(), nil
}

func renderMapper(data mapperData) (string, error) {
	var buf bytes.Buffer
	if err := connectCrudTemplates.ExecuteTemplate(&buf, "mapper.go.tmpl", data); err != nil {
		return "", fmt.Errorf("executing mapper template: %w", err)
	}
	return buf.String(), nil
}

// extractMapperFields extracts non-entity fields from a proto message for mapper generation.
func extractMapperFields(msg *protogen.Message) []mapperField {
	var fields []mapperField
	for _, field := range msg.Fields {
		if string(field.Desc.Name()) == "entity" {
			continue
		}
		fields = append(fields, mapperField{
			ProtoName: toPascalCase(string(field.Desc.Name())),
			SQLCName:  toPascalCase(string(field.Desc.Name())),
		})
	}
	return fields
}

func toPascalCase(s string) string {
	var b strings.Builder
	upper := true
	for _, r := range s {
		if r == '_' {
			upper = true
			continue
		}
		if upper {
			b.WriteRune(unicode.ToUpper(r))
			upper = false
		} else {
			b.WriteRune(r)
		}
	}
	return b.String()
}
