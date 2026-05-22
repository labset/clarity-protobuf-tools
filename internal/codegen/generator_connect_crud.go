package codegen

import (
	"bytes"
	"embed"
	"fmt"
	"go/format"
	"strings"
	"text/template"
	"unicode"

	"github.com/labset/clarity-protobuf-tools/internal/protoutil"
	"google.golang.org/protobuf/compiler/protogen"
	"google.golang.org/protobuf/reflect/protoreflect"
)

//go:embed templates/connect-crud/*.tmpl
var connectCrudTemplateFS embed.FS

var connectCrudTemplates = template.Must(
	template.New("connect-crud").Funcs(template.FuncMap{
		"snakeCase": toSnakeCase,
	}).ParseFS(connectCrudTemplateFS, "templates/connect-crud/*.tmpl"),
)

type connectCrudGenerator struct {
	atlasSqlc *atlasSqlcGenerator
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
	DBPrefix    string
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
	IsEnum    bool   // true if the field is a proto enum
	EnumType  string // short enum type name, e.g. "Status"
	IsRef     bool   // true if the field is a reference type
	RefType   string // short ref type name, e.g. "CategoryRef"
}

type rpcData struct {
	Package     string
	DBPrefix    string
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
	if err := g.atlasSqlc.Generate(plugin); err != nil {
		return err
	}

	packages, err := collectServiceEntities(plugin)
	if err != nil {
		return err
	}

	for _, pe := range packages {
		outDir := fmt.Sprintf("%s/api", pe.meta.outputDir())
		if g.atlasSqlc.sqlc.outputDir != "" {
			outDir = fmt.Sprintf("%s/%s", g.atlasSqlc.sqlc.outputDir, outDir)
		}

		storeBase := fmt.Sprintf(
			"internal/%s/%s/%s/db",
			pe.meta.Provider,
			pe.meta.Domain,
			pe.meta.Version,
		)
		if g.atlasSqlc.sqlc.outputDir != "" {
			storeBase = fmt.Sprintf("%s/%s", g.atlasSqlc.sqlc.outputDir, storeBase)
		}
		storeImport := fmt.Sprintf("%s/%s", g.goModule, storeBase)

		protoImport, protoAlias, connectImport, connectAlias := deriveGoImports(pe.goPackage)

		for _, msg := range pe.messages {
			ops := protoutil.Operations(msg.Desc)
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
			dbPrefix := toPascalCase(pe.meta.Schema)
			mapData := mapperData{
				Package:     "api",
				DBPrefix:    dbPrefix,
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
			if _, err := plugin.NewGeneratedFile(mapperPath, "").
				Write([]byte(mapperContent)); err != nil {
				return err
			}

			rpc := rpcData{
				Package:     "api",
				DBPrefix:    dbPrefix,
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
				if _, err := plugin.NewGeneratedFile(rpcPath, "").
					Write([]byte(rpcContent)); err != nil {
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
func deriveGoImports(
	goPackage string,
) (protoImport, protoAlias, connectImport, connectAlias string) {
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
	return formatGo(buf.Bytes())
}

func renderRPC(tmplName string, data rpcData) (string, error) {
	var buf bytes.Buffer
	if err := connectCrudTemplates.ExecuteTemplate(&buf, tmplName, data); err != nil {
		return "", fmt.Errorf("executing %s template: %w", tmplName, err)
	}
	return formatGo(buf.Bytes())
}

func renderMapper(data mapperData) (string, error) {
	var buf bytes.Buffer
	if err := connectCrudTemplates.ExecuteTemplate(&buf, "mapper.go.tmpl", data); err != nil {
		return "", fmt.Errorf("executing mapper template: %w", err)
	}
	return formatGo(buf.Bytes())
}

func formatGo(src []byte) (string, error) {
	formatted, err := format.Source(src)
	if err != nil {
		return "", fmt.Errorf("formatting generated Go code: %w", err)
	}
	return string(formatted), nil
}

// extractMapperFields extracts non-entity fields from a proto message for mapper generation.
func extractMapperFields(msg *protogen.Message) []mapperField {
	var fields []mapperField
	for _, field := range msg.Fields {
		if string(field.Desc.Name()) == "entity" {
			continue
		}
		mf := mapperField{
			ProtoName: toPascalCase(string(field.Desc.Name())),
			SQLCName:  toSQLCName(string(field.Desc.Name())),
		}
		if protoutil.IsReferenceField(field.Desc) {
			mf.IsRef = true
			mf.RefType = string(field.Desc.Message().Name())
			mf.SQLCName = toSQLCName(string(field.Desc.Name()) + "_id")
		} else if field.Desc.Kind() == protoreflect.EnumKind {
			mf.IsEnum = true
			mf.EnumType = string(field.Desc.Enum().Name())
		}
		fields = append(fields, mf)
	}
	return fields
}

// toSQLCName converts a snake_case field name to the PascalCase form that sqlc
// generates, respecting Go initialisms (e.g., "category_id" → "CategoryID").
func toSQLCName(s string) string {
	pascal := toPascalCase(s)
	if strings.HasSuffix(pascal, "Id") {
		return pascal[:len(pascal)-2] + "ID"
	}
	return pascal
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
