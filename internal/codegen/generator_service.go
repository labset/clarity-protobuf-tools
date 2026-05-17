package codegen

import (
	"bytes"
	"embed"
	"fmt"
	"strings"
	"text/template"
	"unicode"

	pluginV1 "github.com/labset/clarity-protobuf-tools/api/clarity/plugin/v1"
	"github.com/labset/clarity-protobuf-tools/internal/clarity"
	"google.golang.org/protobuf/compiler/protogen"
)

//go:embed templates/service/*.tmpl
var serviceTemplateFS embed.FS

var serviceTemplates = template.Must(
	template.New("service").Funcs(template.FuncMap{
		"lower":      strings.ToLower,
		"snakeCase":  toSnakeCase,
		"trimPrefix": strings.TrimPrefix,
		"add": func(a, b int) int {
			return a + b
		},
	}).ParseFS(serviceTemplateFS, "templates/service/*.tmpl"),
)

type serviceGenerator struct {
	outputDir string
}

type serviceEntityData struct {
	Syntax    string
	Package   string
	GoPackage string
	Model     string
	ModelFile string
	Ops       []pluginV1.Operation
}

type rpcFileData struct {
	Syntax    string
	Package   string
	GoPackage string
	Model     string
	ModelFile string
	Op        pluginV1.Operation
	Fields    []serviceField
}

type serviceField struct {
	Name     string
	Type     string
	Number   int
	Optional bool
}

func (g *serviceGenerator) Generate(plugin *protogen.Plugin) error {
	packages, err := collectPackageEntities(plugin)
	if err != nil {
		return err
	}

	for _, pe := range packages {
		outDir := fmt.Sprintf("%s/%s/%s", pe.meta.Provider, pe.meta.Domain, pe.meta.Version)
		if g.outputDir != "" {
			outDir = fmt.Sprintf("%s/%s", g.outputDir, outDir)
		}

		for _, msg := range pe.messages {
			ops := clarity.Operations(msg.Desc)
			if len(ops) == 0 {
				continue
			}

			modelName := string(msg.Desc.Name())
			modelSnake := toSnakeCase(modelName)
			modelFile := "models.proto"

			protoPackage := string(pe.meta.Provider) + "." + pe.meta.Domain + "." + pe.meta.Version
			goPackage := fmt.Sprintf("github.com/labset/clarity-protobuf-tools/api/%s/%s/%s;%s%s",
				pe.meta.Provider, pe.meta.Domain, pe.meta.Version,
				pe.meta.Domain, capitalize(pe.meta.Version))

			entityData := serviceEntityData{
				Syntax:    "proto3",
				Package:   protoPackage,
				GoPackage: goPackage,
				Model:     modelName,
				ModelFile: modelFile,
				Ops:       ops,
			}

			// Generate service_<model>.proto
			serviceContent, err := renderServiceProto(entityData)
			if err != nil {
				return err
			}
			servicePath := fmt.Sprintf("%s/service_%s.proto", outDir, modelSnake)
			if _, err := plugin.NewGeneratedFile(servicePath, "").Write([]byte(serviceContent)); err != nil {
				return err
			}

			// Generate rpc_<op>_<model>.proto per operation
			for _, op := range ops {
				rpcData := rpcFileData{
					Syntax:    "proto3",
					Package:   protoPackage,
					GoPackage: goPackage,
					Model:     modelName,
					ModelFile: modelFile,
					Op:        op,
					Fields:    extractMutableFields(msg),
				}

				rpcContent, err := renderRPCProto(rpcData)
				if err != nil {
					return err
				}
				opName := strings.ToLower(strings.TrimPrefix(op.String(), "OPERATION_"))
				rpcPath := fmt.Sprintf("%s/rpc_%s_%s.proto", outDir, opName, modelSnake)
				if _, err := plugin.NewGeneratedFile(rpcPath, "").Write([]byte(rpcContent)); err != nil {
					return err
				}
			}
		}
	}

	return nil
}

func extractMutableFields(msg *protogen.Message) []serviceField {
	var fields []serviceField
	number := 1
	for _, field := range msg.Fields {
		if string(field.Desc.Name()) == "entity" {
			continue
		}
		f := serviceField{
			Name:     string(field.Desc.Name()),
			Type:     protoFieldType(field),
			Number:   number,
			Optional: field.Desc.HasOptionalKeyword(),
		}
		fields = append(fields, f)
		number++
	}
	return fields
}

func protoFieldType(field *protogen.Field) string {
	if field.Message != nil {
		return string(field.Message.Desc.FullName())
	}
	return field.Desc.Kind().String()
}

func renderServiceProto(data serviceEntityData) (string, error) {
	var buf bytes.Buffer
	if err := serviceTemplates.ExecuteTemplate(&buf, "service.proto.tmpl", data); err != nil {
		return "", fmt.Errorf("executing service template: %w", err)
	}
	return buf.String(), nil
}

func renderRPCProto(data rpcFileData) (string, error) {
	var buf bytes.Buffer
	if err := serviceTemplates.ExecuteTemplate(&buf, "rpc.proto.tmpl", data); err != nil {
		return "", fmt.Errorf("executing rpc template: %w", err)
	}
	return buf.String(), nil
}

func capitalize(s string) string {
	if s == "" {
		return s
	}
	runes := []rune(s)
	runes[0] = unicode.ToUpper(runes[0])
	return string(runes)
}
