package codegen

import (
	"bytes"
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

func TestParsePackage(t *testing.T) {
	meta, err := parsePackage("acme.inventory.v1")
	require.NoError(t, err)
	assert.Equal(t, "acme", meta.Provider)
	assert.Equal(t, "inventory", meta.Domain)
	assert.Equal(t, "v1", meta.Version)
	assert.Equal(t, "acme_inventory", meta.Schema)
}

func TestParsePackage_Invalid(t *testing.T) {
	_, err := parsePackage("invalid")
	assert.Error(t, err)
}

func TestParsePackage_OutputDir(t *testing.T) {
	meta, err := parsePackage("acme.inventory.v1")
	require.NoError(t, err)
	assert.Equal(t, "internal/acme/inventory/v1", meta.outputDir())
}

func TestToSnakeCase(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"Product", "product"},
		{"MyProduct", "my_product"},
		{"simple", "simple"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			assert.Equal(t, tt.want, toSnakeCase(tt.input))
		})
	}
}

func loadGolden(t *testing.T, name string) string {
	t.Helper()
	data, err := os.ReadFile("testdata/golden/sqlc/" + name)
	require.NoError(t, err)
	return string(data)
}

func TestRenderSchema(t *testing.T) {
	data := schemaData{
		Schema: "acme_inventory",
		Tables: []tableData{
			{
				Name: "product",
				Columns: []string{
					"id UUID PRIMARY KEY NOT NULL",
					"created_at TIMESTAMPTZ NOT NULL",
					"updated_at TIMESTAMPTZ NOT NULL",
					"name TEXT NOT NULL",
					"price BIGINT NOT NULL",
				},
			},
		},
	}

	var buf bytes.Buffer
	err := sqlcTemplates.ExecuteTemplate(&buf, "schema.sql.tmpl", data)
	require.NoError(t, err)

	assert.Equal(t, loadGolden(t, "schema.sql"), buf.String())
}

func TestRenderQueries(t *testing.T) {
	data := queryData{
		Schema:          "acme_inventory",
		Table:           "product",
		MessageName:     "Product",
		AllColumns:      "id, created_at, updated_at, name, price",
		AllPlaceholders: "$1, $2, $3, $4, $5",
		UpdateSetClause: "created_at = $2, updated_at = $3, name = $4, price = $5",
	}

	var buf bytes.Buffer
	err := sqlcTemplates.ExecuteTemplate(&buf, "queries.sql.tmpl", data)
	require.NoError(t, err)

	assert.Equal(t, loadGolden(t, "queries_product.sql"), buf.String())
}

func TestRenderConfig(t *testing.T) {
	data := configData{Package: "v1"}

	var buf bytes.Buffer
	err := sqlcTemplates.ExecuteTemplate(&buf, "sqlc.yaml.tmpl", data)
	require.NoError(t, err)

	assert.Equal(t, loadGolden(t, "sqlc.yaml"), buf.String())
}

// collectFileDescriptors walks the global proto registry and collects
// the file descriptor proto and all its transitive dependencies in topological order.
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

		// Visit dependencies first.
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

func TestSqlcGenerator_Generate(t *testing.T) {
	// Collect all dependency file descriptors from the global registry.
	deps := collectFileDescriptors(t,
		"clarity/plugin/v1/options.proto",
		"clarity/plugin/v1/entity.proto",
	)

	// Build a sample proto with a ROLE_ENTITY message.
	productFile := &descriptorpb.FileDescriptorProto{
		Name:    proto.String("acme/inventory/v1/product.proto"),
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
		},
	}

	req := &pluginpb.CodeGeneratorRequest{
		FileToGenerate: []string{"acme/inventory/v1/product.proto"},
		ProtoFile:      append(deps, productFile),
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

	assert.Equal(t, loadGolden(t, "schema.sql"), files["internal/acme/inventory/v1/sql/schema.sql"])
	assert.Equal(t, loadGolden(t, "queries_product.sql"), files["internal/acme/inventory/v1/sql/queries/product.sql"])
	assert.Equal(t, loadGolden(t, "sqlc.yaml"), files["internal/acme/inventory/v1/sqlc.yaml"])
}
