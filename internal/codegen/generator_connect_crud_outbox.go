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
	template.New("connect-crud-outbox").ParseFS(
		connectCrudOutboxTemplateFS, "templates/connect-crud-outbox/*.tmpl",
	),
)

type connectCrudOutboxGenerator struct {
	atlasSqlc *atlasSqlcGenerator
	goModule  string
}

type eventData struct {
	Model      string
	ModelSnake string
}

// outboxOpTemplateMap maps mutating operations to their event template names.
var outboxOpTemplateMap = map[string]string{
	"OPERATION_CREATE": "event_create.go.tmpl",
	"OPERATION_UPDATE": "event_update.go.tmpl",
	"OPERATION_DELETE": "event_delete.go.tmpl",
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
		outboxDir := fmt.Sprintf("%s/outbox", pe.meta.outputDir())
		if g.atlasSqlc.sqlc.outputDir != "" {
			outboxDir = fmt.Sprintf("%s/%s", g.atlasSqlc.sqlc.outputDir, outboxDir)
		}

		for _, msg := range pe.messages {
			ops := clarity.Operations(msg.Desc)
			if len(ops) == 0 {
				continue
			}

			modelName := string(msg.Desc.Name())
			modelSnake := toSnakeCase(modelName)

			data := eventData{
				Model:      modelName,
				ModelSnake: modelSnake,
			}

			for _, op := range ops {
				tmplName, ok := outboxOpTemplateMap[op.String()]
				if !ok {
					continue
				}
				content, err := renderEvent(tmplName, data)
				if err != nil {
					return err
				}
				opName := strings.ToLower(strings.TrimPrefix(op.String(), "OPERATION_"))
				filePath := fmt.Sprintf("%s/event_%s_%s.go", outboxDir, opName, modelSnake)
				if _, err := plugin.NewGeneratedFile(filePath, "").
					Write([]byte(content)); err != nil {
					return err
				}
			}
		}
	}

	return nil
}

func renderEvent(tmplName string, data eventData) (string, error) {
	var buf bytes.Buffer
	if err := connectCrudOutboxTemplates.ExecuteTemplate(&buf, tmplName, data); err != nil {
		return "", fmt.Errorf("executing %s template: %w", tmplName, err)
	}
	return formatGo(buf.Bytes())
}
