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

func loadConnectCrudOutboxGolden(t *testing.T, name string) string {
	t.Helper()
	data, err := os.ReadFile("testdata/golden/connect-crud-outbox/" + name)
	require.NoError(t, err)
	return string(data)
}

func newTestConnectCrudOutboxGenerator(outputDir string) *connectCrudOutboxGenerator {
	return &connectCrudOutboxGenerator{
		atlasSqlc: &atlasSqlcGenerator{sqlc: &sqlcGenerator{outputDir: outputDir}},
		goModule:  "github.com/acme/app",
	}
}

func TestConnectCrudOutboxGenerator_AllOperations(t *testing.T) {
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

	gen := newTestConnectCrudOutboxGenerator("")
	err = gen.Generate(plugin)
	require.NoError(t, err)

	resp := plugin.Response()
	require.NotNil(t, resp)

	files := make(map[string]string)
	for _, f := range resp.GetFile() {
		files[f.GetName()] = f.GetContent()
	}

	// atlas-sqlc: 5
	// handler + mapper: 2
	// RPCs: 5 (create, get, list, update, delete)
	// outbox events: 3 (create, update, delete)
	// total = 15
	assert.Len(t, files, 15)

	// Verify atlas-sqlc files are present
	assert.Contains(t, files, "internal/acme/inventory/v1/sql/schema.sql")
	assert.Contains(t, files, "internal/acme/inventory/v1/sql/queries/product.sql")
	assert.Contains(t, files, "internal/acme/inventory/v1/sqlc.yaml")
	assert.Contains(t, files, "internal/acme/inventory/v1/atlas.hcl")
	assert.Contains(t, files, "internal/acme/inventory/v1/sql/baseline.sql")

	// Verify handler
	assert.Equal(
		t,
		loadConnectCrudOutboxGolden(t, "handler_product.go"),
		files["internal/acme/inventory/v1/api/handler_product.go"],
	)

	// Verify mapper (reused from connect-crud)
	assert.Equal(
		t,
		loadConnectCrudOutboxGolden(t, "mapper_product.go"),
		files["internal/acme/inventory/v1/api/mapper_product.go"],
	)

	// Verify transactional RPCs
	assert.Equal(
		t,
		loadConnectCrudOutboxGolden(t, "rpc_create_product.go"),
		files["internal/acme/inventory/v1/api/rpc_create_product.go"],
	)
	assert.Equal(
		t,
		loadConnectCrudOutboxGolden(t, "rpc_update_product.go"),
		files["internal/acme/inventory/v1/api/rpc_update_product.go"],
	)
	assert.Equal(
		t,
		loadConnectCrudOutboxGolden(t, "rpc_delete_product.go"),
		files["internal/acme/inventory/v1/api/rpc_delete_product.go"],
	)

	// Verify read-only RPCs (unchanged from connect-crud)
	assert.Equal(
		t,
		loadConnectCrudOutboxGolden(t, "rpc_get_product.go"),
		files["internal/acme/inventory/v1/api/rpc_get_product.go"],
	)
	assert.Equal(
		t,
		loadConnectCrudOutboxGolden(t, "rpc_list_product.go"),
		files["internal/acme/inventory/v1/api/rpc_list_product.go"],
	)

	// Verify outbox event files
	assert.Equal(
		t,
		loadConnectCrudOutboxGolden(t, "event_create_product.go"),
		files["internal/acme/inventory/v1/outbox/event_create_product.go"],
	)
	assert.Equal(
		t,
		loadConnectCrudOutboxGolden(t, "event_update_product.go"),
		files["internal/acme/inventory/v1/outbox/event_update_product.go"],
	)
	assert.Equal(
		t,
		loadConnectCrudOutboxGolden(t, "event_delete_product.go"),
		files["internal/acme/inventory/v1/outbox/event_delete_product.go"],
	)

	// Verify no event files for get/list
	assert.NotContains(t, files, "internal/acme/inventory/v1/outbox/event_get_product.go")
	assert.NotContains(t, files, "internal/acme/inventory/v1/outbox/event_list_product.go")
}

func TestConnectCrudOutboxGenerator_NoOperations(t *testing.T) {
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

	gen := newTestConnectCrudOutboxGenerator("")
	err = gen.Generate(plugin)
	require.NoError(t, err)

	resp := plugin.Response()
	require.NotNil(t, resp)

	files := make(map[string]string)
	for _, f := range resp.GetFile() {
		files[f.GetName()] = f.GetContent()
	}

	// atlas-sqlc files still generated, but no handler/RPC/outbox files
	assert.Contains(t, files, "internal/acme/inventory/v1/sql/schema.sql")
	assert.NotContains(t, files, "internal/acme/inventory/v1/api/handler_product.go")
	assert.NotContains(t, files, "internal/acme/inventory/v1/outbox/event_create_product.go")
}

func TestConnectCrudOutboxGenerator_CreateOnly(t *testing.T) {
	deps := collectFileDescriptors(t,
		"clarity/plugin/v1/options.proto",
		"clarity/plugin/v1/entity.proto",
	)

	modelsFile := testConnectCrudProtoFile(t, pluginV1.Operation_OPERATION_CREATE)

	req := &pluginpb.CodeGeneratorRequest{
		FileToGenerate: []string{"acme/inventory/v1/models.proto"},
		ProtoFile:      append(deps, modelsFile),
	}

	plugin, err := protogen.Options{}.New(req)
	require.NoError(t, err)

	gen := newTestConnectCrudOutboxGenerator("")
	err = gen.Generate(plugin)
	require.NoError(t, err)

	resp := plugin.Response()
	require.NotNil(t, resp)

	files := make(map[string]string)
	for _, f := range resp.GetFile() {
		files[f.GetName()] = f.GetContent()
	}

	// atlas-sqlc: 5 + handler + mapper: 2 + rpc: 1 + event: 1 = 9
	assert.Len(t, files, 9)
	assert.Contains(t, files, "internal/acme/inventory/v1/api/handler_product.go")
	assert.Contains(t, files, "internal/acme/inventory/v1/api/mapper_product.go")
	assert.Contains(t, files, "internal/acme/inventory/v1/api/rpc_create_product.go")
	assert.Contains(t, files, "internal/acme/inventory/v1/outbox/event_create_product.go")
	assert.NotContains(t, files, "internal/acme/inventory/v1/outbox/event_update_product.go")
	assert.NotContains(t, files, "internal/acme/inventory/v1/outbox/event_delete_product.go")
}

func TestConnectCrudOutboxGenerator_RefFields(t *testing.T) {
	deps := collectFileDescriptors(t,
		"clarity/plugin/v1/options.proto",
		"clarity/plugin/v1/entity.proto",
	)

	refFiles := testRefProtoFiles(t)
	refFiles[1].MessageType[0].Options = entityMessageOptionsWithOps(t,
		pluginV1.Operation_OPERATION_CREATE,
		pluginV1.Operation_OPERATION_GET,
	)

	req := &pluginpb.CodeGeneratorRequest{
		FileToGenerate: []string{"acme/inventory/v1/models.proto"},
		ProtoFile:      append(deps, refFiles...),
	}

	plugin, err := protogen.Options{}.New(req)
	require.NoError(t, err)

	gen := newTestConnectCrudOutboxGenerator("")
	err = gen.Generate(plugin)
	require.NoError(t, err)

	resp := plugin.Response()
	require.NotNil(t, resp)

	files := make(map[string]string)
	for _, f := range resp.GetFile() {
		files[f.GetName()] = f.GetContent()
	}

	// Verify ref mapper is generated (reuses connect-crud mapper template)
	assert.Equal(
		t,
		loadConnectCrudGolden(t, "ref_mapper_product.go"),
		files["internal/acme/inventory/v1/api/mapper_product.go"],
	)

	// Verify outbox event and transactional RPC are generated for create
	assert.Contains(t, files, "internal/acme/inventory/v1/outbox/event_create_product.go")
	assert.Contains(t, files, "internal/acme/inventory/v1/api/rpc_create_product.go")

	// Verify get RPC is generated (read-only, no event)
	assert.Contains(t, files, "internal/acme/inventory/v1/api/rpc_get_product.go")
	assert.NotContains(t, files, "internal/acme/inventory/v1/outbox/event_get_product.go")
}

func TestConnectCrudOutboxGenerator_OutputDir(t *testing.T) {
	deps := collectFileDescriptors(t,
		"clarity/plugin/v1/options.proto",
		"clarity/plugin/v1/entity.proto",
	)

	modelsFile := testConnectCrudProtoFile(t, pluginV1.Operation_OPERATION_CREATE)

	req := &pluginpb.CodeGeneratorRequest{
		FileToGenerate: []string{"acme/inventory/v1/models.proto"},
		ProtoFile:      append(deps, modelsFile),
	}

	plugin, err := protogen.Options{}.New(req)
	require.NoError(t, err)

	gen := newTestConnectCrudOutboxGenerator("custom/out")
	err = gen.Generate(plugin)
	require.NoError(t, err)

	resp := plugin.Response()
	require.NotNil(t, resp)

	files := make(map[string]string)
	for _, f := range resp.GetFile() {
		files[f.GetName()] = f.GetContent()
	}

	// Verify all paths include output_dir
	assert.Contains(t, files, "custom/out/internal/acme/inventory/v1/api/handler_product.go")
	assert.Contains(t, files, "custom/out/internal/acme/inventory/v1/api/rpc_create_product.go")
	assert.Contains(
		t,
		files,
		"custom/out/internal/acme/inventory/v1/outbox/event_create_product.go",
	)
	assert.Contains(t, files, "custom/out/internal/acme/inventory/v1/sql/schema.sql")

	// Verify store import includes output_dir
	mapperContent := files["custom/out/internal/acme/inventory/v1/api/mapper_product.go"]
	assert.Contains(
		t,
		mapperContent,
		"\"github.com/acme/app/custom/out/internal/acme/inventory/v1/db\"",
	)

	// Verify outbox import includes output_dir
	rpcContent := files["custom/out/internal/acme/inventory/v1/api/rpc_create_product.go"]
	assert.Contains(
		t,
		rpcContent,
		"\"github.com/acme/app/custom/out/internal/acme/inventory/v1/outbox\"",
	)
}
