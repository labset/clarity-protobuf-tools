package codegen

import (
	"bytes"
	"embed"
	"fmt"
	"strings"
	"text/template"

	optionsV1 "github.com/labset/clarity-protobuf-tools/api/labset/options/v1"
	"github.com/labset/clarity-protobuf-tools/internal/protoutil"
	"google.golang.org/protobuf/compiler/protogen"
	"google.golang.org/protobuf/types/descriptorpb"
)

//go:embed templates/service/*.tmpl
var serviceTemplateFS embed.FS

var serviceTemplates = template.Must(
	template.New("service").Funcs(template.FuncMap{
		"lower":      strings.ToLower,
		"snakeCase":  toSnakeCase,
		"trimPrefix": strings.TrimPrefix,
	}).ParseFS(serviceTemplateFS, "templates/service/*.tmpl"),
)

type serviceGenerator struct {
	outputDir string
}

type serviceEntityData struct {
	Syntax       string
	Package      string
	GoPackage    string
	Model        string
	ModelFile    string
	ImportPrefix string
	Ops          []optionsV1.Operation
}

type rpcFileData struct {
	Syntax    string
	Package   string
	GoPackage string
	Model     string
	ModelFile string
	Op        optionsV1.Operation
}

// servicePackageEntities extends packageEntities with service-specific metadata.
type servicePackageEntities struct {
	*packageEntities
	goPackage string
}

func collectServiceEntities(plugin *protogen.Plugin) ([]*servicePackageEntities, error) {
	packages, err := collectPackageEntities(plugin)
	if err != nil {
		return nil, err
	}

	var result []*servicePackageEntities
	for _, pe := range packages {
		goPackage := ""
		expectedPkg := pe.meta.Provider + "." + pe.meta.Domain + "." + pe.meta.Version
		for _, file := range plugin.Files {
			if !file.Generate {
				continue
			}
			if string(file.Desc.Package()) != expectedPkg {
				continue
			}
			if fileOpts, ok := file.Desc.Options().(*descriptorpb.FileOptions); ok &&
				fileOpts != nil {
				goPackage = fileOpts.GetGoPackage()
			}
			break
		}
		result = append(result, &servicePackageEntities{
			packageEntities: pe,
			goPackage:       goPackage,
		})
	}
	return result, nil
}

func (g *serviceGenerator) Generate(plugin *protogen.Plugin) error {
	packages, err := collectServiceEntities(plugin)
	if err != nil {
		return err
	}

	for _, pe := range packages {
		outDir := fmt.Sprintf("%s/%s/%s", pe.meta.Provider, pe.meta.Domain, pe.meta.Version)
		if g.outputDir != "" {
			outDir = fmt.Sprintf("%s/%s", g.outputDir, outDir)
		}

		protoPackage := pe.meta.Provider + "." + pe.meta.Domain + "." + pe.meta.Version
		importPrefix := fmt.Sprintf(
			"%s/%s/%s", pe.meta.Provider, pe.meta.Domain, pe.meta.Version,
		)

		for _, msg := range pe.messages {
			ops := protoutil.Operations(msg.Desc)
			if len(ops) == 0 {
				continue
			}

			modelName := string(msg.Desc.Name())
			modelSnake := toSnakeCase(modelName)
			modelImport := fmt.Sprintf("%s/models.proto", importPrefix)

			entityData := serviceEntityData{
				Syntax:       "proto3",
				Package:      protoPackage,
				GoPackage:    pe.goPackage,
				Model:        modelName,
				ModelFile:    modelImport,
				ImportPrefix: importPrefix,
				Ops:          ops,
			}

			serviceContent, err := renderServiceProto(entityData)
			if err != nil {
				return err
			}
			servicePath := fmt.Sprintf("%s/service_%s.proto", outDir, modelSnake)
			if _, err := plugin.NewGeneratedFile(servicePath, "").
				Write([]byte(serviceContent)); err != nil {
				return err
			}

			for _, op := range ops {
				rpcData := rpcFileData{
					Syntax:    "proto3",
					Package:   protoPackage,
					GoPackage: pe.goPackage,
					Model:     modelName,
					ModelFile: modelImport,
					Op:        op,
				}

				rpcContent, err := renderRPCProto(rpcData)
				if err != nil {
					return err
				}
				opName := strings.ToLower(strings.TrimPrefix(op.String(), "OPERATION_"))
				rpcPath := fmt.Sprintf("%s/rpc_%s_%s.proto", outDir, opName, modelSnake)
				if _, err := plugin.NewGeneratedFile(rpcPath, "").
					Write([]byte(rpcContent)); err != nil {
					return err
				}
			}
		}
	}

	return nil
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
