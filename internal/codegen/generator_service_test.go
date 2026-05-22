package codegen

import (
	"os"
	"testing"

	optionsV1 "github.com/labset/clarity-protobuf-tools/api/labset/options/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/compiler/protogen"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/descriptorpb"
	"google.golang.org/protobuf/types/pluginpb"
)

func loadServiceGolden(t *testing.T, name string) string {
	t.Helper()
	data, err := os.ReadFile("testdata/golden/service/" + name)
	require.NoError(t, err)
	return string(data)
}

func entityMessageOptionsWithOps(
	t *testing.T,
	ops ...optionsV1.Operation,
) *descriptorpb.MessageOptions {
	t.Helper()
	mopts := &descriptorpb.MessageOptions{}
	proto.SetExtension(mopts, optionsV1.E_Message, &optionsV1.LabsetMessageOptions{
		Role:       optionsV1.Role_ROLE_ENTITY,
		Operations: ops,
	})
	return mopts
}

func testServiceProtoFile(
	t *testing.T,
	ops ...optionsV1.Operation,
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
			"labset/options/v1/options.proto",
			"labset/data/v1/entity.proto",
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

func TestServiceGenerator_AllOperations(t *testing.T) {
	deps := collectFileDescriptors(t,
		"labset/options/v1/options.proto",
		"labset/data/v1/entity.proto",
	)

	allOps := []optionsV1.Operation{
		optionsV1.Operation_OPERATION_CREATE,
		optionsV1.Operation_OPERATION_GET,
		optionsV1.Operation_OPERATION_LIST,
		optionsV1.Operation_OPERATION_UPDATE,
		optionsV1.Operation_OPERATION_DELETE,
	}
	modelsFile := testServiceProtoFile(t, allOps...)

	req := &pluginpb.CodeGeneratorRequest{
		FileToGenerate: []string{"acme/inventory/v1/models.proto"},
		ProtoFile:      append(deps, modelsFile),
	}

	plugin, err := protogen.Options{}.New(req)
	require.NoError(t, err)

	gen := &serviceGenerator{}
	err = gen.Generate(plugin)
	require.NoError(t, err)

	resp := plugin.Response()
	require.NotNil(t, resp)

	files := make(map[string]string)
	for _, f := range resp.GetFile() {
		files[f.GetName()] = f.GetContent()
	}

	// 1 service file + 5 rpc files = 6 total
	assert.Len(t, files, 6)

	assert.Equal(
		t,
		loadServiceGolden(t, "service_product.proto"),
		files["acme/inventory/v1/service_product.proto"],
	)
	assert.Equal(
		t,
		loadServiceGolden(t, "rpc_create_product.proto"),
		files["acme/inventory/v1/rpc_create_product.proto"],
	)
	assert.Equal(
		t,
		loadServiceGolden(t, "rpc_get_product.proto"),
		files["acme/inventory/v1/rpc_get_product.proto"],
	)
	assert.Equal(
		t,
		loadServiceGolden(t, "rpc_list_product.proto"),
		files["acme/inventory/v1/rpc_list_product.proto"],
	)
	assert.Equal(
		t,
		loadServiceGolden(t, "rpc_update_product.proto"),
		files["acme/inventory/v1/rpc_update_product.proto"],
	)
	assert.Equal(
		t,
		loadServiceGolden(t, "rpc_delete_product.proto"),
		files["acme/inventory/v1/rpc_delete_product.proto"],
	)
}

func TestServiceGenerator_SingleOperation(t *testing.T) {
	deps := collectFileDescriptors(t,
		"labset/options/v1/options.proto",
		"labset/data/v1/entity.proto",
	)

	modelsFile := testServiceProtoFile(t, optionsV1.Operation_OPERATION_GET)

	req := &pluginpb.CodeGeneratorRequest{
		FileToGenerate: []string{"acme/inventory/v1/models.proto"},
		ProtoFile:      append(deps, modelsFile),
	}

	plugin, err := protogen.Options{}.New(req)
	require.NoError(t, err)

	gen := &serviceGenerator{}
	err = gen.Generate(plugin)
	require.NoError(t, err)

	resp := plugin.Response()
	require.NotNil(t, resp)

	files := make(map[string]string)
	for _, f := range resp.GetFile() {
		files[f.GetName()] = f.GetContent()
	}

	// 1 service file + 1 rpc file
	assert.Len(t, files, 2)
	assert.Contains(t, files, "acme/inventory/v1/service_product.proto")
	assert.Contains(t, files, "acme/inventory/v1/rpc_get_product.proto")
}

func TestServiceGenerator_NoOperations(t *testing.T) {
	deps := collectFileDescriptors(t,
		"labset/options/v1/options.proto",
		"labset/data/v1/entity.proto",
	)

	// Entity with no operations — should produce no output
	modelsFile := testServiceProtoFile(t)

	req := &pluginpb.CodeGeneratorRequest{
		FileToGenerate: []string{"acme/inventory/v1/models.proto"},
		ProtoFile:      append(deps, modelsFile),
	}

	plugin, err := protogen.Options{}.New(req)
	require.NoError(t, err)

	gen := &serviceGenerator{}
	err = gen.Generate(plugin)
	require.NoError(t, err)

	resp := plugin.Response()
	require.NotNil(t, resp)
	assert.Empty(t, resp.GetFile())
}

func TestServiceGenerator_MultipleEntities(t *testing.T) {
	deps := collectFileDescriptors(t,
		"labset/options/v1/options.proto",
		"labset/data/v1/entity.proto",
	)

	modelsFile := &descriptorpb.FileDescriptorProto{
		Name:    proto.String("acme/inventory/v1/models.proto"),
		Package: proto.String("acme.inventory.v1"),
		Syntax:  proto.String("proto3"),
		Options: &descriptorpb.FileOptions{
			GoPackage: proto.String(
				"github.com/acme/inventory/v1;inventoryv1",
			),
		},
		Dependency: []string{
			"labset/options/v1/options.proto",
			"labset/data/v1/entity.proto",
		},
		MessageType: []*descriptorpb.DescriptorProto{
			{
				Name: proto.String("Product"),
				Options: entityMessageOptionsWithOps(
					t,
					optionsV1.Operation_OPERATION_CREATE,
					optionsV1.Operation_OPERATION_GET,
				),
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
				},
			},
			{
				Name: proto.String("Order"),
				Options: entityMessageOptionsWithOps(
					t,
					optionsV1.Operation_OPERATION_GET,
					optionsV1.Operation_OPERATION_LIST,
					optionsV1.Operation_OPERATION_DELETE,
				),
				Field: []*descriptorpb.FieldDescriptorProto{
					{
						Name:     proto.String("entity"),
						Number:   proto.Int32(1),
						Type:     descriptorpb.FieldDescriptorProto_TYPE_MESSAGE.Enum(),
						TypeName: proto.String(".clarity.plugin.v1.Entity"),
					},
					{
						Name:   proto.String("quantity"),
						Number: proto.Int32(2),
						Type:   descriptorpb.FieldDescriptorProto_TYPE_INT32.Enum(),
					},
				},
			},
		},
	}

	req := &pluginpb.CodeGeneratorRequest{
		FileToGenerate: []string{"acme/inventory/v1/models.proto"},
		ProtoFile:      append(deps, modelsFile),
	}

	plugin, err := protogen.Options{}.New(req)
	require.NoError(t, err)

	gen := &serviceGenerator{}
	err = gen.Generate(plugin)
	require.NoError(t, err)

	resp := plugin.Response()
	require.NotNil(t, resp)

	files := make(map[string]string)
	for _, f := range resp.GetFile() {
		files[f.GetName()] = f.GetContent()
	}

	// Product: 1 service + 2 rpcs = 3
	// Order: 1 service + 3 rpcs = 4
	// Total = 7
	assert.Len(t, files, 7)

	// Product files
	assert.Contains(t, files, "acme/inventory/v1/service_product.proto")
	assert.Contains(t, files, "acme/inventory/v1/rpc_create_product.proto")
	assert.Contains(t, files, "acme/inventory/v1/rpc_get_product.proto")

	// Order files
	assert.Contains(t, files, "acme/inventory/v1/service_order.proto")
	assert.Contains(t, files, "acme/inventory/v1/rpc_get_order.proto")
	assert.Contains(t, files, "acme/inventory/v1/rpc_list_order.proto")
	assert.Contains(t, files, "acme/inventory/v1/rpc_delete_order.proto")
}

func TestServiceGenerator_OutputDir(t *testing.T) {
	deps := collectFileDescriptors(t,
		"labset/options/v1/options.proto",
		"labset/data/v1/entity.proto",
	)

	modelsFile := testServiceProtoFile(t, optionsV1.Operation_OPERATION_GET)

	req := &pluginpb.CodeGeneratorRequest{
		FileToGenerate: []string{"acme/inventory/v1/models.proto"},
		ProtoFile:      append(deps, modelsFile),
	}

	plugin, err := protogen.Options{}.New(req)
	require.NoError(t, err)

	gen := &serviceGenerator{outputDir: "custom/out"}
	err = gen.Generate(plugin)
	require.NoError(t, err)

	resp := plugin.Response()
	require.NotNil(t, resp)

	for _, f := range resp.GetFile() {
		assert.Contains(t, f.GetName(), "custom/out/acme/inventory/v1/")
	}
}
