package codegen

import (
	"os"
	"testing"

	pluginV1 "github.com/labset/clarity-protobuf-tools/api/clarity/plugin/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/compiler/protogen"
	"google.golang.org/protobuf/types/pluginpb"
)

func loadMcpToolsGolden(t *testing.T, name string) string {
	t.Helper()
	data, err := os.ReadFile("testdata/golden/mcp-tools/" + name)
	require.NoError(t, err)
	return string(data)
}

func newTestMcpToolsGenerator(outputDir string) *mcpToolsGenerator {
	return &mcpToolsGenerator{
		connectCrud: newTestConnectCrudGenerator(outputDir),
	}
}

func TestMcpToolsGenerator_AllOperations(t *testing.T) {
	deps := collectFileDescriptors(t,
		"clarity/plugin/v1/options.proto",
		"clarity/plugin/v1/entity.proto",
	)

	allOps := []pluginV1.Operation{
		pluginV1.Operation_OPERATION_CREATE,
		pluginV1.Operation_OPERATION_GET,
		pluginV1.Operation_OPERATION_LIST,
		pluginV1.Operation_OPERATION_UPDATE,
		pluginV1.Operation_OPERATION_DELETE,
	}
	modelsFile := testConnectCrudProtoFile(t, allOps...)

	req := &pluginpb.CodeGeneratorRequest{
		FileToGenerate: []string{"acme/inventory/v1/models.proto"},
		ProtoFile:      append(deps, modelsFile),
	}

	plugin, err := protogen.Options{}.New(req)
	require.NoError(t, err)

	gen := newTestMcpToolsGenerator("")
	err = gen.Generate(plugin)
	require.NoError(t, err)

	resp := plugin.Response()
	require.NotNil(t, resp)

	files := make(map[string]string)
	for _, f := range resp.GetFile() {
		files[f.GetName()] = f.GetContent()
	}

	// atlas-sqlc: 5 + connect-crud: 7 + mcp-tools: 6 (5 tool files + 1 registry) = 18
	assert.Len(t, files, 18)

	// Verify connect-crud files still present
	assert.Equal(
		t,
		loadConnectCrudGolden(t, "handler_product.go"),
		files["internal/acme/inventory/v1/api/handler_product.go"],
	)

	// Verify MCP registry
	assert.Equal(
		t,
		loadMcpToolsGolden(t, "registry_product.go"),
		files["internal/acme/inventory/v1/mcp/registry_product.go"],
	)

	// Verify all MCP tool files
	assert.Equal(
		t,
		loadMcpToolsGolden(t, "tool_create_product.go"),
		files["internal/acme/inventory/v1/mcp/tool_create_product.go"],
	)
	assert.Equal(
		t,
		loadMcpToolsGolden(t, "tool_get_product.go"),
		files["internal/acme/inventory/v1/mcp/tool_get_product.go"],
	)
	assert.Equal(
		t,
		loadMcpToolsGolden(t, "tool_list_product.go"),
		files["internal/acme/inventory/v1/mcp/tool_list_product.go"],
	)
	assert.Equal(
		t,
		loadMcpToolsGolden(t, "tool_update_product.go"),
		files["internal/acme/inventory/v1/mcp/tool_update_product.go"],
	)
	assert.Equal(
		t,
		loadMcpToolsGolden(t, "tool_delete_product.go"),
		files["internal/acme/inventory/v1/mcp/tool_delete_product.go"],
	)
}

func TestMcpToolsGenerator_NoOperations(t *testing.T) {
	deps := collectFileDescriptors(t,
		"clarity/plugin/v1/options.proto",
		"clarity/plugin/v1/entity.proto",
	)

	modelsFile := testConnectCrudProtoFile(t)

	req := &pluginpb.CodeGeneratorRequest{
		FileToGenerate: []string{"acme/inventory/v1/models.proto"},
		ProtoFile:      append(deps, modelsFile),
	}

	plugin, err := protogen.Options{}.New(req)
	require.NoError(t, err)

	gen := newTestMcpToolsGenerator("")
	err = gen.Generate(plugin)
	require.NoError(t, err)

	resp := plugin.Response()
	require.NotNil(t, resp)

	files := make(map[string]string)
	for _, f := range resp.GetFile() {
		files[f.GetName()] = f.GetContent()
	}

	assert.Contains(t, files, "internal/acme/inventory/v1/sql/schema.sql")
	assert.NotContains(t, files, "internal/acme/inventory/v1/mcp/registry_product.go")
}

func TestMcpToolsGenerator_SingleOperation(t *testing.T) {
	deps := collectFileDescriptors(t,
		"clarity/plugin/v1/options.proto",
		"clarity/plugin/v1/entity.proto",
	)

	modelsFile := testConnectCrudProtoFile(t, pluginV1.Operation_OPERATION_GET)

	req := &pluginpb.CodeGeneratorRequest{
		FileToGenerate: []string{"acme/inventory/v1/models.proto"},
		ProtoFile:      append(deps, modelsFile),
	}

	plugin, err := protogen.Options{}.New(req)
	require.NoError(t, err)

	gen := newTestMcpToolsGenerator("")
	err = gen.Generate(plugin)
	require.NoError(t, err)

	resp := plugin.Response()
	require.NotNil(t, resp)

	files := make(map[string]string)
	for _, f := range resp.GetFile() {
		files[f.GetName()] = f.GetContent()
	}

	// atlas-sqlc: 5 + connect-crud: 3 + mcp-tools: 2 = 10
	assert.Len(t, files, 10)
	assert.Contains(t, files, "internal/acme/inventory/v1/mcp/tool_get_product.go")
	assert.Contains(t, files, "internal/acme/inventory/v1/mcp/registry_product.go")
	assert.NotContains(t, files, "internal/acme/inventory/v1/mcp/tool_create_product.go")
}

func TestMcpToolsGenerator_ModeRegistration(t *testing.T) {
	gen, err := GeneratorForMode("mode=mcp-tools,go_module=github.com/acme/app")
	require.NoError(t, err)
	assert.IsType(t, &mcpToolsGenerator{}, gen)
}

func TestMcpToolsGenerator_ModeRequiresGoModule(t *testing.T) {
	_, err := GeneratorForMode("mode=mcp-tools")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "go_module")
}
