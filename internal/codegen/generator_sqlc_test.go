package codegen

import (
	"bytes"
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

func TestRenderSchema(t *testing.T) {
	// Template rendering test with mock data (no protogen dependency).
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

	var buf = new(bytes.Buffer)
	err := sqlcTemplates.ExecuteTemplate(buf, "schema.sql.tmpl", data)
	require.NoError(t, err)

	result := buf.String()
	assert.Contains(t, result, "CREATE SCHEMA IF NOT EXISTS acme_inventory;")
	assert.Contains(t, result, "CREATE TABLE acme_inventory.product (")
	assert.Contains(t, result, "id UUID PRIMARY KEY NOT NULL")
	assert.Contains(t, result, "name TEXT NOT NULL")
	assert.Contains(t, result, "price BIGINT NOT NULL")
}
