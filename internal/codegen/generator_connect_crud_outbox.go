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

//go:embed templates/connect-crud-outbox/*.tmpl
var connectCrudOutboxTemplateFS embed.FS

var connectCrudOutboxTemplates = template.Must(
	template.New("connect-crud-outbox").Funcs(template.FuncMap{
		"snakeCase": toSnakeCase,
	}).ParseFS(connectCrudOutboxTemplateFS, "templates/connect-crud-outbox/*.tmpl"),
)

type connectCrudOutboxGenerator struct {
	atlasSqlc *atlasSqlcGenerator
	goModule  string
}

type eventData struct {
	Model      string
	ModelSnake string
}

type outboxRpcData struct {
	Package      string
	DBPrefix     string
	Model        string
	ModelLower   string
	StoreImport  string
	ProtoImport  string
	ProtoAlias   string
	OutboxImport string
	Fields       []mapperField
}

// outboxOpTemplateMap maps mutating operations to their outbox RPC template names.
var outboxOpTemplateMap = map[string]string{
	"OPERATION_CREATE": "rpc_create.go.tmpl",
	"OPERATION_UPDATE": "rpc_update.go.tmpl",
	"OPERATION_DELETE": "rpc_delete.go.tmpl",
}

// outboxEventTemplateMap maps mutating operations to their event template names.
var outboxEventTemplateMap = map[string]string{
	"OPERATION_CREATE": "event_create.go.tmpl",
	"OPERATION_UPDATE": "event_update.go.tmpl",
	"OPERATION_DELETE": "event_delete.go.tmpl",
}

// readOnlyOpTemplateMap maps read operations to their connect-crud template names.
var readOnlyOpTemplateMap = map[string]string{
	"OPERATION_GET":  "rpc_get.go.tmpl",
	"OPERATION_LIST": "rpc_list.go.tmpl",
}

func (g *connectCrudOutboxGenerator) Generate(plugin *protogen.Plugin) error {
	if err := g.atlasSqlc.Generate(plugin); err != nil {
		return err
	}

	packages, err := collectServiceEntities(plugin)
	if err != nil {
		return err
	}

	for _, pe := range packages {
		baseDir := pe.meta.outputDir()
		if g.atlasSqlc.sqlc.outputDir != "" {
			baseDir = fmt.Sprintf("%s/%s", g.atlasSqlc.sqlc.outputDir, baseDir)
		}
		apiDir := fmt.Sprintf("%s/api", baseDir)
		outboxDir := fmt.Sprintf("%s/outbox", baseDir)

		storeBase := fmt.Sprintf(
			"internal/%s/%s/%s/db",
			pe.meta.Provider, pe.meta.Domain, pe.meta.Version,
		)
		if g.atlasSqlc.sqlc.outputDir != "" {
			storeBase = fmt.Sprintf("%s/%s", g.atlasSqlc.sqlc.outputDir, storeBase)
		}
		storeImport := fmt.Sprintf("%s/%s", g.goModule, storeBase)

		outboxBase := fmt.Sprintf(
			"internal/%s/%s/%s/outbox",
			pe.meta.Provider, pe.meta.Domain, pe.meta.Version,
		)
		if g.atlasSqlc.sqlc.outputDir != "" {
			outboxBase = fmt.Sprintf("%s/%s", g.atlasSqlc.sqlc.outputDir, outboxBase)
		}
		outboxImport := fmt.Sprintf("%s/%s", g.goModule, outboxBase)

		protoImport, protoAlias, connectImport, connectAlias := deriveGoImports(pe.goPackage)

		for _, msg := range pe.messages {
			ops := clarity.Operations(msg.Desc)
			if len(ops) == 0 {
				continue
			}

			modelName := string(msg.Desc.Name())
			modelSnake := toSnakeCase(modelName)
			modelLower := strings.ToLower(modelName[:1]) + modelName[1:]

			// Generate handler
			hData := handlerData{
				Package:       "api",
				Model:         modelName,
				ModelLower:    modelLower,
				ModelSnake:    modelSnake,
				StoreImport:   storeImport,
				ConnectImport: connectImport,
				ConnectAlias:  connectAlias,
			}
			content, err := renderOutboxHandler(hData)
			if err != nil {
				return err
			}
			filePath := fmt.Sprintf("%s/handler_%s.go", apiDir, modelSnake)
			if _, err := plugin.NewGeneratedFile(filePath, "").Write([]byte(content)); err != nil {
				return err
			}

			// Generate mapper (reuse connect-crud template)
			fields := extractMapperFields(msg)
			dbPrefix := toPascalCase(pe.meta.Schema)
			mData := mapperData{
				Package:     "api",
				DBPrefix:    dbPrefix,
				Model:       modelName,
				ModelLower:  modelLower,
				StoreImport: storeImport,
				ProtoImport: protoImport,
				ProtoAlias:  protoAlias,
				Fields:      fields,
			}
			mapperContent, err := renderMapper(mData)
			if err != nil {
				return err
			}
			mapperPath := fmt.Sprintf("%s/mapper_%s.go", apiDir, modelSnake)
			if _, err := plugin.NewGeneratedFile(mapperPath, "").
				Write([]byte(mapperContent)); err != nil {
				return err
			}

			// Generate RPC and event files per operation
			oRpc := outboxRpcData{
				Package:      "api",
				DBPrefix:     dbPrefix,
				Model:        modelName,
				ModelLower:   modelLower,
				StoreImport:  storeImport,
				ProtoImport:  protoImport,
				ProtoAlias:   protoAlias,
				OutboxImport: outboxImport,
				Fields:       fields,
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
			evtData := eventData{
				Model:      modelName,
				ModelSnake: modelSnake,
			}

			for _, op := range ops {
				opName := strings.ToLower(strings.TrimPrefix(op.String(), "OPERATION_"))

				// Mutating operations: outbox RPC + event file
				if tmplName, ok := outboxOpTemplateMap[op.String()]; ok {
					rpcContent, err := renderOutboxRPC(tmplName, oRpc)
					if err != nil {
						return err
					}
					rpcPath := fmt.Sprintf("%s/rpc_%s_%s.go", apiDir, opName, modelSnake)
					if _, err := plugin.NewGeneratedFile(rpcPath, "").
						Write([]byte(rpcContent)); err != nil {
						return err
					}

					evtTmpl := outboxEventTemplateMap[op.String()]
					evtContent, err := renderEvent(evtTmpl, evtData)
					if err != nil {
						return err
					}
					evtPath := fmt.Sprintf("%s/event_%s_%s.go", outboxDir, opName, modelSnake)
					if _, err := plugin.NewGeneratedFile(evtPath, "").
						Write([]byte(evtContent)); err != nil {
						return err
					}
					continue
				}

				// Read-only operations: reuse connect-crud templates
				if tmplName, ok := readOnlyOpTemplateMap[op.String()]; ok {
					rpcContent, err := renderRPC(tmplName, rpc)
					if err != nil {
						return err
					}
					rpcPath := fmt.Sprintf("%s/rpc_%s_%s.go", apiDir, opName, modelSnake)
					if _, err := plugin.NewGeneratedFile(rpcPath, "").
						Write([]byte(rpcContent)); err != nil {
						return err
					}
				}
			}
		}
	}

	return nil
}

func renderOutboxHandler(data handlerData) (string, error) {
	var buf bytes.Buffer
	if err := connectCrudOutboxTemplates.ExecuteTemplate(&buf, "handler.go.tmpl", data); err != nil {
		return "", fmt.Errorf("executing outbox handler template: %w", err)
	}
	return formatGo(buf.Bytes())
}

func renderOutboxRPC(tmplName string, data outboxRpcData) (string, error) {
	var buf bytes.Buffer
	if err := connectCrudOutboxTemplates.ExecuteTemplate(&buf, tmplName, data); err != nil {
		return "", fmt.Errorf("executing outbox %s template: %w", tmplName, err)
	}
	return formatGo(buf.Bytes())
}

func renderEvent(tmplName string, data eventData) (string, error) {
	var buf bytes.Buffer
	if err := connectCrudOutboxTemplates.ExecuteTemplate(&buf, tmplName, data); err != nil {
		return "", fmt.Errorf("executing %s template: %w", tmplName, err)
	}
	return formatGo(buf.Bytes())
}
