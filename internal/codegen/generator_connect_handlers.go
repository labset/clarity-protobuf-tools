package codegen

import (
	"bytes"
	"embed"
	"fmt"
	"strings"
	"text/template"

	"google.golang.org/protobuf/compiler/protogen"
	"google.golang.org/protobuf/types/descriptorpb"
)

//go:embed templates/connect-handlers/*.tmpl
var connectHandlersTemplateFS embed.FS

var connectHandlersTemplates = template.Must(
	template.New("connect-handlers").
		ParseFS(connectHandlersTemplateFS, "templates/connect-handlers/*.tmpl"),
)

type connectHandlersGenerator struct {
	outputDir string
}

type connectHandlerData struct {
	Package       string
	Service       string
	ServiceLower  string
	ConnectImport string
	ConnectAlias  string
}

type connectRpcStubData struct {
	Package      string
	Service      string
	ServiceLower string
	Method       string
	RequestType  string
	ResponseType string
	ProtoImport  string
	ProtoAlias   string
}

// serviceInfo groups service-level data extracted from a proto file.
type serviceInfo struct {
	goPackage   string
	packageMeta packageMeta
	service     *protogen.Service
}

// collectServices scans plugin files for proto services grouped by package.
func collectServices(plugin *protogen.Plugin) ([]serviceInfo, error) {
	var result []serviceInfo

	for _, file := range plugin.Files {
		if !file.Generate {
			continue
		}
		if len(file.Services) == 0 {
			continue
		}

		pkg := string(file.Desc.Package())
		meta, err := parsePackage(pkg)
		if err != nil {
			return nil, err
		}

		goPackage := ""
		if fileOpts, ok := file.Desc.Options().(*descriptorpb.FileOptions); ok && fileOpts != nil {
			goPackage = fileOpts.GetGoPackage()
		}

		for _, svc := range file.Services {
			result = append(result, serviceInfo{
				goPackage:   goPackage,
				packageMeta: meta,
				service:     svc,
			})
		}
	}

	return result, nil
}

func (g *connectHandlersGenerator) Generate(plugin *protogen.Plugin) error {
	services, err := collectServices(plugin)
	if err != nil {
		return err
	}

	for _, si := range services {
		outDir := fmt.Sprintf("%s/api", si.packageMeta.outputDir())
		if g.outputDir != "" {
			outDir = fmt.Sprintf("%s/%s", g.outputDir, outDir)
		}

		svcName := string(si.service.Desc.Name())
		svcSnake := toSnakeCase(svcName)
		svcLower := strings.ToLower(svcName[:1]) + svcName[1:]

		protoImport, protoAlias, connectImport, connectAlias := deriveGoImports(si.goPackage)

		// Generate handler file
		handlerPath := fmt.Sprintf("%s/handler_%s.go", outDir, svcSnake)
		if !fileExists(handlerPath) {
			hData := connectHandlerData{
				Package:       "api",
				Service:       svcName,
				ServiceLower:  svcLower,
				ConnectImport: connectImport,
				ConnectAlias:  connectAlias,
			}
			content, err := renderConnectHandler(hData)
			if err != nil {
				return err
			}
			if _, err := plugin.NewGeneratedFile(handlerPath, "").
				Write([]byte(content)); err != nil {
				return err
			}
		}

		// Generate RPC stub files
		for _, method := range si.service.Methods {
			methodName := string(method.Desc.Name())
			methodSnake := toSnakeCase(methodName)

			rpcPath := fmt.Sprintf("%s/rpc_%s.go", outDir, methodSnake)
			if fileExists(rpcPath) {
				continue
			}

			rData := connectRpcStubData{
				Package:      "api",
				Service:      svcName,
				ServiceLower: svcLower,
				Method:       methodName,
				RequestType:  string(method.Input.Desc.Name()),
				ResponseType: string(method.Output.Desc.Name()),
				ProtoImport:  protoImport,
				ProtoAlias:   protoAlias,
			}
			content, err := renderConnectRpcStub(rData)
			if err != nil {
				return err
			}
			if _, err := plugin.NewGeneratedFile(rpcPath, "").Write([]byte(content)); err != nil {
				return err
			}
		}
	}

	return nil
}

func renderConnectHandler(data connectHandlerData) (string, error) {
	var buf bytes.Buffer
	if err := connectHandlersTemplates.ExecuteTemplate(&buf, "handler.go.tmpl", data); err != nil {
		return "", fmt.Errorf("executing connect handler template: %w", err)
	}
	return formatGo(buf.Bytes())
}

func renderConnectRpcStub(data connectRpcStubData) (string, error) {
	var buf bytes.Buffer
	if err := connectHandlersTemplates.ExecuteTemplate(&buf, "rpc.go.tmpl", data); err != nil {
		return "", fmt.Errorf("executing connect rpc stub template: %w", err)
	}
	return formatGo(buf.Bytes())
}
