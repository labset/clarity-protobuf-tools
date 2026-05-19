package codegen

import (
	"bytes"
	"embed"
	"fmt"
	"strings"
	"text/template"

	"github.com/labset/clarity-protobuf-tools/internal/clarity"
	"google.golang.org/protobuf/compiler/protogen"
	"google.golang.org/protobuf/reflect/protoreflect"
)

//go:embed templates/mcp-tools/*.tmpl
var mcpToolsTemplateFS embed.FS

var mcpToolsTemplates = template.Must(
	template.New("mcp-tools").ParseFS(mcpToolsTemplateFS, "templates/mcp-tools/*.tmpl"),
)

type mcpToolsGenerator struct {
	connectCrud *connectCrudGenerator
}

type mcpField struct {
	Name     string // PascalCase, e.g. "Name"
	JSONName string // snake_case, e.g. "name"
	GoType   string // Go type, e.g. "string", "int64"
	IsEnum   bool
	EnumType string
	IsRef    bool
	RefType  string
}

type mcpToolData struct {
	Package     string
	Model       string
	ModelLower  string
	ModelSnake  string
	ProtoImport string
	ProtoAlias  string
	Fields      []mcpField
}

type mcpOp struct {
	ToolName    string // e.g. "create_product"
	MethodName  string // e.g. "createProduct"
	Description string // e.g. "Create a new product"
}

type mcpRegistrationData struct {
	Package     string
	Model       string
	ModelLower  string
	ModelSnake  string
	StoreImport string
	Operations  []mcpOp
}

var mcpToolTemplateMap = map[string]string{
	"OPERATION_CREATE": "mcp_tool_create.go.tmpl",
	"OPERATION_GET":    "mcp_tool_get.go.tmpl",
	"OPERATION_LIST":   "mcp_tool_list.go.tmpl",
	"OPERATION_UPDATE": "mcp_tool_update.go.tmpl",
	"OPERATION_DELETE": "mcp_tool_delete.go.tmpl",
}

func (g *mcpToolsGenerator) Generate(plugin *protogen.Plugin) error {
	if err := g.connectCrud.Generate(plugin); err != nil {
		return err
	}

	packages, err := collectServiceEntities(plugin)
	if err != nil {
		return err
	}

	for _, pe := range packages {
		outDir := fmt.Sprintf("%s/api", pe.meta.outputDir())
		if g.connectCrud.atlasSqlc.sqlc.outputDir != "" {
			outDir = fmt.Sprintf("%s/%s", g.connectCrud.atlasSqlc.sqlc.outputDir, outDir)
		}

		storeBase := fmt.Sprintf(
			"internal/%s/%s/%s/db",
			pe.meta.Provider, pe.meta.Domain, pe.meta.Version,
		)
		if g.connectCrud.atlasSqlc.sqlc.outputDir != "" {
			storeBase = fmt.Sprintf("%s/%s", g.connectCrud.atlasSqlc.sqlc.outputDir, storeBase)
		}
		storeImport := fmt.Sprintf("%s/%s", g.connectCrud.goModule, storeBase)

		protoImport, protoAlias, _, _ := deriveGoImports(pe.goPackage)

		for _, msg := range pe.messages {
			ops := clarity.Operations(msg.Desc)
			if len(ops) == 0 {
				continue
			}

			modelName := string(msg.Desc.Name())
			modelSnake := toSnakeCase(modelName)
			modelLower := strings.ToLower(modelName[:1]) + modelName[1:]

			fields := extractMcpFields(msg)

			toolData := mcpToolData{
				Package:     "api",
				Model:       modelName,
				ModelLower:  modelLower,
				ModelSnake:  modelSnake,
				ProtoImport: protoImport,
				ProtoAlias:  protoAlias,
				Fields:      fields,
			}

			var mcpOps []mcpOp
			for _, op := range ops {
				tmplName, ok := mcpToolTemplateMap[op.String()]
				if !ok {
					continue
				}

				opName := strings.ToLower(strings.TrimPrefix(op.String(), "OPERATION_"))
				content, err := renderMcpTool(tmplName, toolData)
				if err != nil {
					return err
				}
				filePath := fmt.Sprintf("%s/mcp_tool_%s_%s.go", outDir, opName, modelSnake)
				if _, err := plugin.NewGeneratedFile(filePath, "").Write([]byte(content)); err != nil {
					return err
				}

				mcpOps = append(mcpOps, buildMcpOp(opName, modelName, modelSnake))
			}

			regData := mcpRegistrationData{
				Package:     "api",
				Model:       modelName,
				ModelLower:  modelLower,
				ModelSnake:  modelSnake,
				StoreImport: storeImport,
				Operations:  mcpOps,
			}
			regContent, err := renderMcpRegistration(regData)
			if err != nil {
				return err
			}
			regPath := fmt.Sprintf("%s/mcp_tools_%s.go", outDir, modelSnake)
			if _, err := plugin.NewGeneratedFile(regPath, "").Write([]byte(regContent)); err != nil {
				return err
			}
		}
	}

	return nil
}

func buildMcpOp(opName, modelName, modelSnake string) mcpOp {
	switch opName {
	case "create":
		return mcpOp{
			ToolName:    "create_" + modelSnake,
			MethodName:  "create" + modelName,
			Description: "Create a new " + modelSnake,
		}
	case "get":
		return mcpOp{
			ToolName:    "get_" + modelSnake,
			MethodName:  "get" + modelName,
			Description: "Get a " + modelSnake + " by ID",
		}
	case "list":
		return mcpOp{
			ToolName:    "list_" + modelSnake + "s",
			MethodName:  "list" + modelName + "s",
			Description: "List " + modelSnake + "s",
		}
	case "update":
		return mcpOp{
			ToolName:    "update_" + modelSnake,
			MethodName:  "update" + modelName,
			Description: "Update a " + modelSnake,
		}
	case "delete":
		return mcpOp{
			ToolName:    "delete_" + modelSnake,
			MethodName:  "delete" + modelName,
			Description: "Delete a " + modelSnake,
		}
	default:
		return mcpOp{}
	}
}

func extractMcpFields(msg *protogen.Message) []mcpField {
	var fields []mcpField
	for _, field := range msg.Fields {
		if string(field.Desc.Name()) == "entity" {
			continue
		}
		f := mcpField{
			Name:     toPascalCase(string(field.Desc.Name())),
			JSONName: string(field.Desc.Name()),
		}
		if clarity.IsReferenceField(field.Desc) {
			f.IsRef = true
			f.RefType = string(field.Desc.Message().Name())
			f.GoType = "string"
		} else if field.Desc.Kind() == protoreflect.EnumKind {
			f.IsEnum = true
			f.EnumType = string(field.Desc.Enum().Name())
			f.GoType = "string"
		} else {
			f.GoType = protoKindToGoType(field.Desc.Kind())
		}
		fields = append(fields, f)
	}
	return fields
}

func protoKindToGoType(k protoreflect.Kind) string {
	switch k {
	case protoreflect.StringKind:
		return "string"
	case protoreflect.Int64Kind, protoreflect.Sint64Kind, protoreflect.Sfixed64Kind:
		return "int64"
	case protoreflect.Int32Kind, protoreflect.Sint32Kind, protoreflect.Sfixed32Kind:
		return "int32"
	case protoreflect.Uint64Kind, protoreflect.Fixed64Kind:
		return "uint64"
	case protoreflect.Uint32Kind, protoreflect.Fixed32Kind:
		return "uint32"
	case protoreflect.BoolKind:
		return "bool"
	case protoreflect.DoubleKind:
		return "float64"
	case protoreflect.FloatKind:
		return "float32"
	case protoreflect.BytesKind:
		return "[]byte"
	default:
		return "string"
	}
}

func renderMcpTool(tmplName string, data mcpToolData) (string, error) {
	var buf bytes.Buffer
	if err := mcpToolsTemplates.ExecuteTemplate(&buf, tmplName, data); err != nil {
		return "", fmt.Errorf("executing mcp %s template: %w", tmplName, err)
	}
	return formatGo(buf.Bytes())
}

func renderMcpRegistration(data mcpRegistrationData) (string, error) {
	var buf bytes.Buffer
	if err := mcpToolsTemplates.ExecuteTemplate(&buf, "mcp_tools.go.tmpl", data); err != nil {
		return "", fmt.Errorf("executing mcp registration template: %w", err)
	}
	return formatGo(buf.Bytes())
}
