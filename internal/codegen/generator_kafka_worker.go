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

type kafkaWorkerData struct {
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

type kafkaConsumerData struct {
	Package     string
	Model       string
	ModelLower  string
	ModelSnake  string
	Domain      string
	Version     string
	Subscribers []kafkaSubscriber
}

type kafkaSubscriber struct {
	Lower string
	Title string
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

		outboxBase := fmt.Sprintf(
			"internal/%s/%s/%s/outbox",
			pe.meta.Provider, pe.meta.Domain, pe.meta.Version,
		)
		if g.outputDir != "" {
			outboxBase = fmt.Sprintf("%s/%s", g.outputDir, outboxBase)
		}
		outboxImport := fmt.Sprintf("%s/%s", g.goModule, outboxBase)

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

			data := kafkaWorkerData{
				Package:      "workers",
				Model:        modelName,
				ModelLower:   modelLower,
				ModelSnake:   modelSnake,
				Domain:       pe.meta.Domain,
				Version:      pe.meta.Version,
				OutboxImport: outboxImport,
				Operations:   workerOps,
			}

			content, err := renderKafkaWorker(data)
			if err != nil {
				return err
			}
			filePath := fmt.Sprintf("%s/worker_%s.go", workersDir, modelSnake)
			if _, err := plugin.NewGeneratedFile(filePath, "").Write([]byte(content)); err != nil {
				return err
			}

			var subs []kafkaSubscriber
			for _, s := range subscribers {
				if s == 0 {
					continue
				}
				lower := strings.ToLower(strings.TrimPrefix(s.String(), "SUBSCRIBER_"))
				title := strings.ToUpper(lower[:1]) + lower[1:]
				subs = append(subs, kafkaSubscriber{Lower: lower, Title: title})
			}

			consumerData := kafkaConsumerData{
				Package:     "workers",
				Model:       modelName,
				ModelLower:  modelLower,
				ModelSnake:  modelSnake,
				Domain:      pe.meta.Domain,
				Version:     pe.meta.Version,
				Subscribers: subs,
			}

			consumerContent, err := renderKafkaConsumer(consumerData)
			if err != nil {
				return err
			}
			consumerPath := fmt.Sprintf("%s/consumer_%s.go", workersDir, modelSnake)
			if _, err := plugin.NewGeneratedFile(consumerPath, "").Write([]byte(consumerContent)); err != nil {
				return err
			}

			hasWorkers = true
		}

		if hasWorkers {
			envelopeContent, err := renderKafkaEnvelope()
			if err != nil {
				return err
			}
			envelopePath := fmt.Sprintf("%s/envelope.go", workersDir)
			if _, err := plugin.NewGeneratedFile(envelopePath, "").Write([]byte(envelopeContent)); err != nil {
				return err
			}
		}
	}

	return nil
}

func renderKafkaWorker(data kafkaWorkerData) (string, error) {
	var buf bytes.Buffer
	if err := kafkaWorkerTemplates.ExecuteTemplate(&buf, "worker.go.tmpl", data); err != nil {
		return "", fmt.Errorf("executing kafka worker template: %w", err)
	}
	return formatGo(buf.Bytes())
}

func renderKafkaConsumer(data kafkaConsumerData) (string, error) {
	var buf bytes.Buffer
	if err := kafkaWorkerTemplates.ExecuteTemplate(&buf, "consumer.go.tmpl", data); err != nil {
		return "", fmt.Errorf("executing kafka consumer template: %w", err)
	}
	return formatGo(buf.Bytes())
}

func renderKafkaEnvelope() (string, error) {
	var buf bytes.Buffer
	if err := kafkaWorkerTemplates.ExecuteTemplate(&buf, "envelope.go.tmpl", nil); err != nil {
		return "", fmt.Errorf("executing kafka envelope template: %w", err)
	}
	return formatGo(buf.Bytes())
}
