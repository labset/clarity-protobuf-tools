package codegen

import (
	"os"
	"testing"

	pluginV1 "github.com/labset/clarity-protobuf-tools/api/clarity/plugin/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/compiler/protogen"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/descriptorpb"
	"google.golang.org/protobuf/types/pluginpb"
)

func loadConnectCrudGolden(t *testing.T, name string) string {
	t.Helper()
	data, err := os.ReadFile("testdata/golden/connect-crud/" + name)
	require.NoError(t, err)
	return string(data)
}

func newTestConnectCrudGenerator(outputDir string) *connectCrudGenerator {
	return &connectCrudGenerator{
		atlasSqlc: &atlasSqlcGenerator{sqlc: &sqlcGenerator{outputDir: outputDir}},
		goModule:  "github.com/acme/app",
	}
}

func testConnectCrudProtoFile(
	t *testing.T,
	ops ...pluginV1.Operation,
) *descriptorpb.FileDescriptorProto {
	t.Helper()
	return &descriptorpb.FileDescriptorProto{
		Name:    proto.String("acme/inventory/v1/models.proto"),
		Package: proto.String("acme.inventory.v1"),
		Syntax:  proto.String("proto3"),
		Options: &descriptorpb.FileOptions{
			GoPackage: proto.String("github.com/acme/inventory/v1;inventoryv1"),
		},
		Dependency: []string{
			"clarity/plugin/v1/options.proto",
			"clarity/plugin/v1/entity.proto",
		},
		MessageType: []*descriptorpb.DescriptorProto{
			{
				Name:    proto.String("Product"),
				Options: entityMessageOptionsWithOps(t, ops...),
				Field: []*descriptorpb.FieldDescriptorProto{
					{
						Name:     proto.String("entity"),
						Number:   proto.Int32(1),
						Type:     descriptorpb.FieldDescriptorProto_TYPE_MESSAGE.Enum(),
						TypeName: proto.String(".clarity.plugin.v1.Entity"),
					},
					{
						Name:   proto.String("name"),
						Number: proto.Int32(2),
						Type:   descriptorpb.FieldDescriptorProto_TYPE_STRING.Enum(),
					},
					{
						Name:   proto.String("price"),
						Number: proto.Int32(3),
						Type:   descriptorpb.FieldDescriptorProto_TYPE_INT64.Enum(),
					},
				},
			},
		},
	}
}

func TestConnectCrudGenerator_AllOperations(t *testing.T) {
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

	gen := newTestConnectCrudGenerator("")
	err = gen.Generate(plugin)
	require.NoError(t, err)

	resp := plugin.Response()
	require.NotNil(t, resp)

	files := make(map[string]string)
	for _, f := range resp.GetFile() {
		files[f.GetName()] = f.GetContent()
	}

	// atlas-sqlc files: schema.sql, queries/product.sql, sqlc.yaml, atlas.hcl, baseline.sql = 5
	// connect-crud files: 1 handler + 1 mapper + 5 rpc files = 7
	// total = 12
	assert.Len(t, files, 12)

	// Verify atlas-sqlc files are present
	assert.Contains(t, files, "internal/acme/inventory/v1/sql/schema.sql")
	assert.Contains(t, files, "internal/acme/inventory/v1/sql/queries/product.sql")
	assert.Contains(t, files, "internal/acme/inventory/v1/sqlc.yaml")
	assert.Contains(t, files, "internal/acme/inventory/v1/atlas.hcl")
	assert.Contains(t, files, "internal/acme/inventory/v1/sql/baseline.sql")

	// Verify connect-crud files
	assert.Equal(
		t,
		loadConnectCrudGolden(t, "handler_product.go"),
		files["internal/acme/inventory/v1/api/handler_product.go"],
	)
	assert.Equal(
		t,
		loadConnectCrudGolden(t, "mapper_product.go"),
		files["internal/acme/inventory/v1/api/mapper_product.go"],
	)
	assert.Equal(
		t,
		loadConnectCrudGolden(t, "rpc_create_product.go"),
		files["internal/acme/inventory/v1/api/rpc_create_product.go"],
	)
	assert.Equal(
		t,
		loadConnectCrudGolden(t, "rpc_get_product.go"),
		files["internal/acme/inventory/v1/api/rpc_get_product.go"],
	)
	assert.Equal(
		t,
		loadConnectCrudGolden(t, "rpc_list_product.go"),
		files["internal/acme/inventory/v1/api/rpc_list_product.go"],
	)
	assert.Equal(
		t,
		loadConnectCrudGolden(t, "rpc_update_product.go"),
		files["internal/acme/inventory/v1/api/rpc_update_product.go"],
	)
	assert.Equal(
		t,
		loadConnectCrudGolden(t, "rpc_delete_product.go"),
		files["internal/acme/inventory/v1/api/rpc_delete_product.go"],
	)
}

func TestConnectCrudGenerator_NoOperations(t *testing.T) {
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

	gen := newTestConnectCrudGenerator("")
	err = gen.Generate(plugin)
	require.NoError(t, err)

	resp := plugin.Response()
	require.NotNil(t, resp)

	files := make(map[string]string)
	for _, f := range resp.GetFile() {
		files[f.GetName()] = f.GetContent()
	}

	// atlas-sqlc files still generated (entity exists), but no connect-crud files (no operations)
	assert.Contains(t, files, "internal/acme/inventory/v1/sql/schema.sql")
	assert.NotContains(t, files, "internal/acme/inventory/v1/api/handler_product.go")
}

func TestConnectCrudGenerator_SingleOperation(t *testing.T) {
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

	gen := newTestConnectCrudGenerator("")
	err = gen.Generate(plugin)
	require.NoError(t, err)

	resp := plugin.Response()
	require.NotNil(t, resp)

	files := make(map[string]string)
	for _, f := range resp.GetFile() {
		files[f.GetName()] = f.GetContent()
	}

	// atlas-sqlc: 5 files + connect-crud: 1 handler + 1 mapper + 1 rpc = 3 → total 8
	assert.Len(t, files, 8)
	assert.Contains(t, files, "internal/acme/inventory/v1/api/handler_product.go")
	assert.Contains(t, files, "internal/acme/inventory/v1/api/mapper_product.go")
	assert.Contains(t, files, "internal/acme/inventory/v1/api/rpc_get_product.go")
}

func TestConnectCrudGenerator_OutputDir(t *testing.T) {
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

	gen := newTestConnectCrudGenerator("custom/out")
	err = gen.Generate(plugin)
	require.NoError(t, err)

	resp := plugin.Response()
	require.NotNil(t, resp)

	files := make(map[string]string)
	for _, f := range resp.GetFile() {
		files[f.GetName()] = f.GetContent()
	}

	// Verify connect-crud files use output_dir
	assert.Contains(t, files, "custom/out/internal/acme/inventory/v1/api/handler_product.go")
	// Verify atlas-sqlc files also use output_dir
	assert.Contains(t, files, "custom/out/internal/acme/inventory/v1/sql/schema.sql")
}
