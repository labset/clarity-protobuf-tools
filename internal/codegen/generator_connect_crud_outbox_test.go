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

func TestConnectCrudOutboxGenerator_AllMutatingOps(t *testing.T) {
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

	// atlas-sqlc files: 5
	// outbox event files: 3 (create, update, delete — not get/list)
	// total = 8
	assert.Len(t, files, 8)

	// Verify atlas-sqlc files are present
	assert.Contains(t, files, "internal/acme/inventory/v1/sql/schema.sql")

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

	// atlas-sqlc files still generated, but no outbox files
	assert.Contains(t, files, "internal/acme/inventory/v1/sql/schema.sql")
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

	// atlas-sqlc: 5 + outbox: 1 (create only) = 6
	assert.Len(t, files, 6)
	assert.Contains(t, files, "internal/acme/inventory/v1/outbox/event_create_product.go")
	assert.NotContains(t, files, "internal/acme/inventory/v1/outbox/event_update_product.go")
	assert.NotContains(t, files, "internal/acme/inventory/v1/outbox/event_delete_product.go")
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

	assert.Contains(t, files, "custom/out/internal/acme/inventory/v1/outbox/event_create_product.go")
	assert.Contains(t, files, "custom/out/internal/acme/inventory/v1/sql/schema.sql")
}
