package codegen

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/compiler/protogen"
	"google.golang.org/protobuf/types/pluginpb"
)

func loadAtlasSqlcGolden(t *testing.T, name string) string {
	t.Helper()
	data, err := os.ReadFile("testdata/golden/atlas-sqlc/" + name)
	require.NoError(t, err)
	return string(data)
}

func TestAtlasSqlcGenerator_Generate(t *testing.T) {
	deps := collectFileDescriptors(t,
		"labset/options/v1/options.proto",
		"labset/data/v1/entity.proto",
	)

	modelsFile := testProtoFile(t)

	req := &pluginpb.CodeGeneratorRequest{
		FileToGenerate: []string{"acme/inventory/v1/models.proto"},
		ProtoFile:      append(deps, modelsFile),
	}

	plugin, err := protogen.Options{}.New(req)
	require.NoError(t, err)

	gen := &atlasSqlcGenerator{sqlc: &sqlcGenerator{}}
	err = gen.Generate(plugin)
	require.NoError(t, err)

	resp := plugin.Response()
	require.NotNil(t, resp)

	files := make(map[string]string)
	for _, f := range resp.GetFile() {
		files[f.GetName()] = f.GetContent()
	}

	// Verify sqlc files are still generated.
	assert.Contains(t, files, "internal/acme/inventory/v1/sql/schema.sql")
	assert.Contains(t, files, "internal/acme/inventory/v1/sqlc.yaml")
	assert.Contains(t, files, "internal/acme/inventory/v1/sql/queries/product.sql")
	assert.Contains(t, files, "internal/acme/inventory/v1/sql/queries/order.sql")

	// Verify atlas-sqlc specific files.
	assert.Equal(
		t,
		loadAtlasSqlcGolden(t, "atlas.hcl"),
		files["internal/acme/inventory/v1/atlas.hcl"],
	)
	assert.Equal(
		t,
		loadAtlasSqlcGolden(t, "baseline.sql"),
		files["internal/acme/inventory/v1/sql/baseline.sql"],
	)
}

func TestAtlasSqlcGenerator_ExistingSqlcModeUnchanged(t *testing.T) {
	deps := collectFileDescriptors(t,
		"labset/options/v1/options.proto",
		"labset/data/v1/entity.proto",
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
	for _, f := range resp.GetFile() {
		files[f.GetName()] = f.GetContent()
	}

	// sqlc mode should NOT produce atlas files.
	assert.NotContains(t, files, "internal/acme/inventory/v1/atlas.hcl")
	assert.NotContains(t, files, "internal/acme/inventory/v1/sql/baseline.sql")
}
