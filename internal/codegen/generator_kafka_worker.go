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

//go:embed templates/kafka-worker/*.tmpl
var kafkaWorkerTemplateFS embed.FS

var kafkaWorkerTemplates = template.Must(
	template.New("kafka-worker").ParseFS(kafkaWorkerTemplateFS, "templates/kafka-worker/*.tmpl"),
)

type kafkaWorkerGenerator struct {
	goModule  string
	outputDir string
}

type kafkaRegisterData struct {
	Package      string
	Model        string
	ModelLower   string
	ModelSnake   string
	Domain       string
	Version      string
	OutboxImport string
	Operations   []kafkaWorkerOp
}

type kafkaWorkerOp struct {
	Lower        string
	Title        string
	HasFieldMask bool
}

type kafkaWorkerOpData struct {
	Package      string
	Model        string
	ModelLower   string
	ModelSnake   string
	OutboxImport string
	OpLower      string
	OpTitle      string
	HasFieldMask bool
}

type kafkaConsumerSubData struct {
	Package       string
	Model         string
	ModelSnake    string
	Domain        string
	Version       string
	WorkersImport string
	SubLower      string
	SubTitle      string
}

func (g *kafkaWorkerGenerator) Generate(plugin *protogen.Plugin) error {
	packages, err := collectPackageEntities(plugin)
	if err != nil {
		return err
	}

	for _, pe := range packages {
		baseDir := pe.meta.outputDir()
		if g.outputDir != "" {
			baseDir = fmt.Sprintf("%s/%s", g.outputDir, baseDir)
		}
		workersDir := fmt.Sprintf("%s/workers", baseDir)
		consumersDir := fmt.Sprintf("%s/consumers", baseDir)

		outboxBase := fmt.Sprintf(
			"internal/%s/%s/%s/outbox",
			pe.meta.Provider, pe.meta.Domain, pe.meta.Version,
		)
		if g.outputDir != "" {
			outboxBase = fmt.Sprintf("%s/%s", g.outputDir, outboxBase)
		}
		outboxImport := fmt.Sprintf("%s/%s", g.goModule, outboxBase)

		workersBase := fmt.Sprintf(
			"internal/%s/%s/%s/workers",
			pe.meta.Provider, pe.meta.Domain, pe.meta.Version,
		)
		if g.outputDir != "" {
			workersBase = fmt.Sprintf("%s/%s", g.outputDir, workersBase)
		}
		workersImport := fmt.Sprintf("%s/%s", g.goModule, workersBase)

		hasWorkers := false
		for _, msg := range pe.messages {
			subscribers := clarity.Subscribers(msg.Desc)
			if len(subscribers) == 0 {
				continue
			}

			ops := clarity.Operations(msg.Desc)
			if len(ops) == 0 {
				continue
			}

			var workerOps []kafkaWorkerOp
			for _, op := range ops {
				opStr := op.String()
				if _, ok := outboxOpTemplateMap[opStr]; !ok {
					continue
				}
				lower := strings.ToLower(strings.TrimPrefix(opStr, "OPERATION_"))
				title := strings.ToUpper(lower[:1]) + lower[1:]
				workerOps = append(workerOps, kafkaWorkerOp{
					Lower:        lower,
					Title:        title,
					HasFieldMask: opStr == "OPERATION_UPDATE",
				})
			}

			if len(workerOps) == 0 {
				continue
			}

			modelName := string(msg.Desc.Name())
			modelSnake := toSnakeCase(modelName)
			modelLower := strings.ToLower(modelName[:1]) + modelName[1:]

			// Generate register file
			regData := kafkaRegisterData{
				Package:      "workers",
				Model:        modelName,
				ModelLower:   modelLower,
				ModelSnake:   modelSnake,
				Domain:       pe.meta.Domain,
				Version:      pe.meta.Version,
				OutboxImport: outboxImport,
				Operations:   workerOps,
			}
			if err := g.writeTemplate(plugin, "register.go.tmpl", regData,
				fmt.Sprintf("%s/register_%s.go", workersDir, modelSnake)); err != nil {
				return err
			}

			// Generate one worker file per operation
			for _, op := range workerOps {
				opData := kafkaWorkerOpData{
					Package:      "workers",
					Model:        modelName,
					ModelLower:   modelLower,
					ModelSnake:   modelSnake,
					OutboxImport: outboxImport,
					OpLower:      op.Lower,
					OpTitle:      op.Title,
					HasFieldMask: op.HasFieldMask,
				}
				if err := g.writeTemplate(plugin, "worker_op.go.tmpl", opData,
					fmt.Sprintf("%s/worker_%s_%s.go", workersDir, op.Lower, modelSnake)); err != nil {
					return err
				}
			}

			// Generate one consumer file per subscriber
			for _, s := range subscribers {
				if s == 0 {
					continue
				}
				lower := strings.ToLower(strings.TrimPrefix(s.String(), "SUBSCRIBER_"))
				title := strings.ToUpper(lower[:1]) + lower[1:]

				subData := kafkaConsumerSubData{
					Package:       "consumers",
					Model:         modelName,
					ModelSnake:    modelSnake,
					Domain:        pe.meta.Domain,
					Version:       pe.meta.Version,
					WorkersImport: workersImport,
					SubLower:      lower,
					SubTitle:      title,
				}
				if err := g.writeTemplate(plugin, "consumer_sub.go.tmpl", subData,
					fmt.Sprintf("%s/consumer_%s_%s.go", consumersDir, lower, modelSnake)); err != nil {
					return err
				}
			}

			hasWorkers = true
		}

		if hasWorkers {
			if err := g.writeTemplate(plugin, "envelope.go.tmpl", nil,
				fmt.Sprintf("%s/envelope.go", workersDir)); err != nil {
				return err
			}
		}
	}

	return nil
}

func (g *kafkaWorkerGenerator) writeTemplate(plugin *protogen.Plugin, tmpl string, data any, path string) error {
	var buf bytes.Buffer
	if err := kafkaWorkerTemplates.ExecuteTemplate(&buf, tmpl, data); err != nil {
		return fmt.Errorf("executing %s template: %w", tmpl, err)
	}
	content, err := formatGo(buf.Bytes())
	if err != nil {
		return err
	}
	_, err = plugin.NewGeneratedFile(path, "").Write([]byte(content))
	return err
}
