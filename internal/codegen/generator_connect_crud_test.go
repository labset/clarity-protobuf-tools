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

	gen := &connectCrudGenerator{goModule: "github.com/acme/app"}
	err = gen.Generate(plugin)
	require.NoError(t, err)

	resp := plugin.Response()
	require.NotNil(t, resp)

	files := make(map[string]string)
	for _, f := range resp.GetFile() {
		files[f.GetName()] = f.GetContent()
	}

	// 1 handler + 1 mapper + 5 rpc files = 7 total
	assert.Len(t, files, 7)

	assert.Equal(
		t,
		loadConnectCrudGolden(t, "handler_product.go"),
		files["acme/inventory/v1/api/handler_product.go"],
	)
	assert.Equal(
		t,
		loadConnectCrudGolden(t, "mapper_product.go"),
		files["acme/inventory/v1/api/mapper_product.go"],
	)
	assert.Equal(
		t,
		loadConnectCrudGolden(t, "rpc_create_product.go"),
		files["acme/inventory/v1/api/rpc_create_product.go"],
	)
	assert.Equal(
		t,
		loadConnectCrudGolden(t, "rpc_get_product.go"),
		files["acme/inventory/v1/api/rpc_get_product.go"],
	)
	assert.Equal(
		t,
		loadConnectCrudGolden(t, "rpc_list_product.go"),
		files["acme/inventory/v1/api/rpc_list_product.go"],
	)
	assert.Equal(
		t,
		loadConnectCrudGolden(t, "rpc_update_product.go"),
		files["acme/inventory/v1/api/rpc_update_product.go"],
	)
	assert.Equal(
		t,
		loadConnectCrudGolden(t, "rpc_delete_product.go"),
		files["acme/inventory/v1/api/rpc_delete_product.go"],
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

	gen := &connectCrudGenerator{goModule: "github.com/acme/app"}
	err = gen.Generate(plugin)
	require.NoError(t, err)

	resp := plugin.Response()
	require.NotNil(t, resp)
	assert.Empty(t, resp.GetFile())
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

	gen := &connectCrudGenerator{goModule: "github.com/acme/app"}
	err = gen.Generate(plugin)
	require.NoError(t, err)

	resp := plugin.Response()
	require.NotNil(t, resp)

	files := make(map[string]string)
	for _, f := range resp.GetFile() {
		files[f.GetName()] = f.GetContent()
	}

	// 1 handler + 1 mapper + 1 rpc file = 3
	assert.Len(t, files, 3)
	assert.Contains(t, files, "acme/inventory/v1/api/handler_product.go")
	assert.Contains(t, files, "acme/inventory/v1/api/mapper_product.go")
	assert.Contains(t, files, "acme/inventory/v1/api/rpc_get_product.go")
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

	gen := &connectCrudGenerator{outputDir: "custom/out", goModule: "github.com/acme/app"}
	err = gen.Generate(plugin)
	require.NoError(t, err)

	resp := plugin.Response()
	require.NotNil(t, resp)

	for _, f := range resp.GetFile() {
		assert.Contains(t, f.GetName(), "custom/out/acme/inventory/v1/api/")
	}
}
