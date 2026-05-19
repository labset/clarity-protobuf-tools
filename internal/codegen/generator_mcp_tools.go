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

//go:embed templates/mcp-tools/*.tmpl
var mcpToolsTemplateFS embed.FS

var mcpToolsTemplates = template.Must(
	template.New("mcp-tools").ParseFS(mcpToolsTemplateFS, "templates/mcp-tools/*.tmpl"),
)

type mcpToolsGenerator struct {
	connectCrud *connectCrudGenerator
}

type mcpToolData struct {
	Package     string
	Model       string
	ModelSnake  string
	ProtoImport string
	ProtoAlias  string
}

type mcpOp struct {
	ToolName    string // e.g. "create_product"
	MethodName  string // e.g. "createProduct"
	Description string // e.g. "Create Product"
}

type mcpRegistrationData struct {
	Package       string
	Model         string
	ModelSnake    string
	ConnectImport string
	ConnectAlias  string
	Operations    []mcpOp
}

var mcpToolTemplateMap = map[string]string{
	"OPERATION_CREATE": "tool_create.go.tmpl",
	"OPERATION_GET":    "tool_get.go.tmpl",
	"OPERATION_LIST":   "tool_list.go.tmpl",
	"OPERATION_UPDATE": "tool_update.go.tmpl",
	"OPERATION_DELETE": "tool_delete.go.tmpl",
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
		outDir := fmt.Sprintf("%s/mcp", pe.meta.outputDir())
		if g.connectCrud.atlasSqlc.sqlc.outputDir != "" {
			outDir = fmt.Sprintf("%s/%s", g.connectCrud.atlasSqlc.sqlc.outputDir, outDir)
		}

		protoImport, protoAlias, connectImport, connectAlias := deriveGoImports(pe.goPackage)

		for _, msg := range pe.messages {
			ops := clarity.Operations(msg.Desc)
			if len(ops) == 0 {
				continue
			}

			modelName := string(msg.Desc.Name())
			modelSnake := toSnakeCase(modelName)

			toolData := mcpToolData{
				Package:     "mcp",
				Model:       modelName,
				ModelSnake:  modelSnake,
				ProtoImport: protoImport,
				ProtoAlias:  protoAlias,
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
				filePath := fmt.Sprintf("%s/tool_%s_%s.go", outDir, opName, modelSnake)
				if _, err := plugin.NewGeneratedFile(filePath, "").Write([]byte(content)); err != nil {
					return err
				}

				mcpOps = append(mcpOps, buildMcpOp(opName, modelName, modelSnake))
			}

			regData := mcpRegistrationData{
				Package:       "mcp",
				Model:         modelName,
				ModelSnake:    modelSnake,
				ConnectImport: connectImport,
				ConnectAlias:  connectAlias,
				Operations:    mcpOps,
			}
			regContent, err := renderMcpRegistration(regData)
			if err != nil {
				return err
			}
			regPath := fmt.Sprintf("%s/registry_%s.go", outDir, modelSnake)
			if _, err := plugin.NewGeneratedFile(regPath, "").Write([]byte(regContent)); err != nil {
				return err
			}
		}
	}

	return nil
}

func buildMcpOp(opName, modelName, modelSnake string) mcpOp {
	opPascal := toPascalCase(opName)
	toolName := opName + "_" + modelSnake
	methodName := opName + modelName
	if opName == "list" {
		toolName = opName + "_" + modelSnake + "s"
		methodName = opName + modelName + "s"
	}
	return mcpOp{
		ToolName:    toolName,
		MethodName:  methodName,
		Description: opPascal + " " + modelName,
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
	if err := mcpToolsTemplates.ExecuteTemplate(&buf, "registry.go.tmpl", data); err != nil {
		return "", fmt.Errorf("executing mcp registration template: %w", err)
	}
	return formatGo(buf.Bytes())
}
