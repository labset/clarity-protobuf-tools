package codegen

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/compiler/protogen"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/descriptorpb"
	"google.golang.org/protobuf/types/pluginpb"
)

func loadConnectHandlersGolden(t *testing.T, name string) string {
	t.Helper()
	data, err := os.ReadFile("testdata/golden/connect-handlers/" + name)
	require.NoError(t, err)
	return string(data)
}

func testConnectHandlersProtoFile(t *testing.T) *descriptorpb.FileDescriptorProto {
	t.Helper()
	return &descriptorpb.FileDescriptorProto{
		Name:    proto.String("acme/inventory/v1/service_product.proto"),
		Package: proto.String("acme.inventory.v1"),
		Syntax:  proto.String("proto3"),
		Options: &descriptorpb.FileOptions{
			GoPackage: proto.String("github.com/acme/inventory/v1;inventoryv1"),
		},
		MessageType: []*descriptorpb.DescriptorProto{
			{Name: proto.String("CreateProductRequest")},
			{Name: proto.String("CreateProductResponse")},
			{Name: proto.String("GetProductRequest")},
			{Name: proto.String("GetProductResponse")},
			{Name: proto.String("ListProductsRequest")},
			{Name: proto.String("ListProductsResponse")},
		},
		Service: []*descriptorpb.ServiceDescriptorProto{
			{
				Name: proto.String("ProductService"),
				Method: []*descriptorpb.MethodDescriptorProto{
					{
						Name:       proto.String("CreateProduct"),
						InputType:  proto.String(".acme.inventory.v1.CreateProductRequest"),
						OutputType: proto.String(".acme.inventory.v1.CreateProductResponse"),
					},
					{
						Name:       proto.String("GetProduct"),
						InputType:  proto.String(".acme.inventory.v1.GetProductRequest"),
						OutputType: proto.String(".acme.inventory.v1.GetProductResponse"),
					},
					{
						Name:       proto.String("ListProducts"),
						InputType:  proto.String(".acme.inventory.v1.ListProductsRequest"),
						OutputType: proto.String(".acme.inventory.v1.ListProductsResponse"),
					},
				},
			},
		},
	}
}

func TestConnectHandlersGenerator_AllMethods(t *testing.T) {
	serviceFile := testConnectHandlersProtoFile(t)

	req := &pluginpb.CodeGeneratorRequest{
		FileToGenerate: []string{"acme/inventory/v1/service_product.proto"},
		ProtoFile:      []*descriptorpb.FileDescriptorProto{serviceFile},
	}

	plugin, err := protogen.Options{}.New(req)
	require.NoError(t, err)

	gen := &connectHandlersGenerator{}
	err = gen.Generate(plugin)
	require.NoError(t, err)

	resp := plugin.Response()
	require.NotNil(t, resp)

	files := make(map[string]string)
	for _, f := range resp.GetFile() {
		files[f.GetName()] = f.GetContent()
	}

	// 1 handler + 3 RPC stubs = 4 files
	assert.Len(t, files, 4)

	assert.Equal(
		t,
		loadConnectHandlersGolden(t, "handler_product_service.go"),
		files["internal/acme/inventory/v1/api/handler_product_service.go"],
	)
	assert.Equal(
		t,
		loadConnectHandlersGolden(t, "rpc_create_product.go"),
		files["internal/acme/inventory/v1/api/rpc_create_product.go"],
	)
	assert.Equal(
		t,
		loadConnectHandlersGolden(t, "rpc_get_product.go"),
		files["internal/acme/inventory/v1/api/rpc_get_product.go"],
	)
	assert.Equal(
		t,
		loadConnectHandlersGolden(t, "rpc_list_products.go"),
		files["internal/acme/inventory/v1/api/rpc_list_products.go"],
	)
}

func TestConnectHandlersGenerator_NoServices(t *testing.T) {
	noServiceFile := &descriptorpb.FileDescriptorProto{
		Name:    proto.String("acme/inventory/v1/models.proto"),
		Package: proto.String("acme.inventory.v1"),
		Syntax:  proto.String("proto3"),
		Options: &descriptorpb.FileOptions{
			GoPackage: proto.String("github.com/acme/inventory/v1;inventoryv1"),
		},
		MessageType: []*descriptorpb.DescriptorProto{
			{Name: proto.String("Product")},
		},
	}

	req := &pluginpb.CodeGeneratorRequest{
		FileToGenerate: []string{"acme/inventory/v1/models.proto"},
		ProtoFile:      []*descriptorpb.FileDescriptorProto{noServiceFile},
	}

	plugin, err := protogen.Options{}.New(req)
	require.NoError(t, err)

	gen := &connectHandlersGenerator{}
	err = gen.Generate(plugin)
	require.NoError(t, err)

	resp := plugin.Response()
	require.NotNil(t, resp)
	assert.Empty(t, resp.GetFile())
}

func TestConnectHandlersGenerator_OutputDir(t *testing.T) {
	serviceFile := testConnectHandlersProtoFile(t)

	req := &pluginpb.CodeGeneratorRequest{
		FileToGenerate: []string{"acme/inventory/v1/service_product.proto"},
		ProtoFile:      []*descriptorpb.FileDescriptorProto{serviceFile},
	}

	plugin, err := protogen.Options{}.New(req)
	require.NoError(t, err)

	gen := &connectHandlersGenerator{outputDir: "custom/out"}
	err = gen.Generate(plugin)
	require.NoError(t, err)

	resp := plugin.Response()
	require.NotNil(t, resp)

	files := make(map[string]string)
	for _, f := range resp.GetFile() {
		files[f.GetName()] = f.GetContent()
	}

	assert.Contains(t, files, "custom/out/internal/acme/inventory/v1/api/handler_product_service.go")
	assert.Contains(t, files, "custom/out/internal/acme/inventory/v1/api/rpc_create_product.go")
}

func TestConnectHandlersGenerator_ModeRegistration(t *testing.T) {
	gen, err := GeneratorForMode("mode=connect-handlers")
	require.NoError(t, err)
	assert.IsType(t, &connectHandlersGenerator{}, gen)
}
