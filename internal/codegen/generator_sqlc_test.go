package codegen

import (
	"bytes"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
