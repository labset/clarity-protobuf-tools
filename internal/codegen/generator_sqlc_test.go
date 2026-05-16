package codegen

import (
	"os"
	"testing"

	pluginV1 "github.com/labset/go-protoc-gen-plugin/api/clarity/plugin/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/compiler/protogen"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protodesc"
	"google.golang.org/protobuf/reflect/protoregistry"
	"google.golang.org/protobuf/types/descriptorpb"
	"google.golang.org/protobuf/types/pluginpb"
)

func loadGolden(t *testing.T, name string) string {
	t.Helper()
	data, err := os.ReadFile("testdata/golden/sqlc/" + name)
	require.NoError(t, err)
	return string(data)
}

func collectFileDescriptors(t *testing.T, paths ...string) []*descriptorpb.FileDescriptorProto {
	t.Helper()
	seen := make(map[string]bool)
	var result []*descriptorpb.FileDescriptorProto

	var visit func(path string)
	visit = func(path string) {
		if seen[path] {
			return
		}
		seen[path] = true
		fd, err := protoregistry.GlobalFiles.FindFileByPath(path)
		require.NoError(t, err, "finding %s in global registry", path)

		imports := fd.Imports()
		for i := range imports.Len() {
			visit(string(imports.Get(i).Path()))
		}
		result = append(result, protodesc.ToFileDescriptorProto(fd))
	}

	for _, p := range paths {
		visit(p)
	}
	return result
}

func entityMessageOptions(t *testing.T) *descriptorpb.MessageOptions {
	t.Helper()
	opts := &descriptorpb.MessageOptions{}
	proto.SetExtension(opts, pluginV1.E_Message, &pluginV1.ClarityMessageOptions{
		Role: pluginV1.Role_ROLE_ENTITY,
	})
	return opts
}

func TestParsePackage_Invalid(t *testing.T) {
	_, err := parsePackage("invalid")
	assert.Error(t, err)
}

func testProtoFile(t *testing.T) *descriptorpb.FileDescriptorProto {
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
				Options: entityMessageOptions(t),
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
			{
				Name:    proto.String("Order"),
				Options: entityMessageOptions(t),
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
					{
						Name:       proto.String("billing_address"),
						Number:     proto.Int32(3),
						Type:       descriptorpb.FieldDescriptorProto_TYPE_STRING.Enum(),
						OneofIndex: proto.Int32(0),
					},
					{
						Name:       proto.String("shipping_address"),
						Number:     proto.Int32(4),
						Type:       descriptorpb.FieldDescriptorProto_TYPE_STRING.Enum(),
						OneofIndex: proto.Int32(0),
					},
				},
				OneofDecl: []*descriptorpb.OneofDescriptorProto{
					{Name: proto.String("address")},
				},
			},
		},
	}
}

func TestSqlcGenerator_Generate(t *testing.T) {
	deps := collectFileDescriptors(t,
		"clarity/plugin/v1/options.proto",
		"clarity/plugin/v1/entity.proto",
	)

	modelsFile := testProtoFile(t)

	req := &pluginpb.CodeGeneratorRequest{
		FileToGenerate: []string{"acme/inventory/v1/models.proto"},
		ProtoFile:      append(deps, modelsFile),
	}

	plugin, err := protogen.Options{}.New(req)
	require.NoError(t, err)

	gen := &sqlcGenerator{}
	err = gen.Generate(plugin)
	require.NoError(t, err)

	resp := plugin.Response()
	require.NotNil(t, resp)

	files := make(map[string]string)
	for _, f := range resp.File {
		files[f.GetName()] = f.GetContent()
	}

	// Verify exactly 4 files generated: schema + 2 queries + sqlc.yaml
	assert.Len(t, files, 4)

	// Schema contains both tables.
	assert.Equal(t, loadGolden(t, "schema.sql"), files["internal/acme/inventory/v1/sql/schema.sql"])

	// Per-entity query files.
	assert.Equal(t, loadGolden(t, "queries_product.sql"), files["internal/acme/inventory/v1/sql/queries/product.sql"])
	assert.Equal(t, loadGolden(t, "queries_order.sql"), files["internal/acme/inventory/v1/sql/queries/order.sql"])

	// sqlc config.
	assert.Equal(t, loadGolden(t, "sqlc.yaml"), files["internal/acme/inventory/v1/sqlc.yaml"])
}

func TestSqlcGenerator_Generate_OutputDir(t *testing.T) {
	deps := collectFileDescriptors(t,
		"clarity/plugin/v1/options.proto",
		"clarity/plugin/v1/entity.proto",
	)

	modelsFile := testProtoFile(t)

	req := &pluginpb.CodeGeneratorRequest{
		FileToGenerate: []string{"acme/inventory/v1/models.proto"},
		ProtoFile:      append(deps, modelsFile),
	}

	plugin, err := protogen.Options{}.New(req)
	require.NoError(t, err)

	gen := &sqlcGenerator{outputDir: "custom/out"}
	err = gen.Generate(plugin)
	require.NoError(t, err)

	resp := plugin.Response()
	require.NotNil(t, resp)

	for _, f := range resp.File {
		assert.Contains(t, f.GetName(), "custom/out/internal/acme/inventory/v1/")
	}
}
