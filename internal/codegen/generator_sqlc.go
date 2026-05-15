package codegen

import (
	"bytes"
	"embed"
	"fmt"
	"strings"
	"text/template"
	"unicode"

	pluginV1 "github.com/labset/go-protoc-gen-plugin/api/clarity/plugin/v1"
	"github.com/labset/go-protoc-gen-plugin/internal/codegen/pgtype"
	"google.golang.org/protobuf/compiler/protogen"
	"google.golang.org/protobuf/proto"
)

//go:embed templates/sqlc/*.tmpl
var sqlcTemplateFS embed.FS

var sqlcTemplates = template.Must(template.ParseFS(sqlcTemplateFS, "templates/sqlc/*.tmpl"))

type sqlcGenerator struct{}

func (g *sqlcGenerator) Generate(plugin *protogen.Plugin) error {
	for _, file := range plugin.Files {
		if !file.Generate {
			continue
		}

		meta, err := parsePackage(string(file.Desc.Package()))
		if err != nil {
			return err
		}

		var entityMessages []*protogen.Message
		for _, msg := range file.Messages {
			if isEntityMessage(msg) {
				entityMessages = append(entityMessages, msg)
			}
		}

		if len(entityMessages) == 0 {
			continue
		}

		schemaContent, err := renderSchema(meta, entityMessages)
		if err != nil {
			return err
		}

		schemaFile := plugin.NewGeneratedFile(
			fmt.Sprintf("%s/sql/schema.sql", meta.outputDir()),
			"",
		)
		schemaFile.P(schemaContent)
	}

	return nil
}

// packageMeta holds derived metadata from a proto package name.
type packageMeta struct {
	Provider string
	Domain   string
	Version  string
	Schema   string // <provider>_<domain>
}

// parsePackage extracts provider, domain, version from a package like "acme.inventory.v1".
func parsePackage(pkg string) (packageMeta, error) {
	parts := strings.Split(pkg, ".")
	if len(parts) < 3 {
		return packageMeta{}, fmt.Errorf("expected package format <provider>.<domain>.<version>, got %q", pkg)
	}
	return packageMeta{
		Provider: parts[0],
		Domain:   parts[1],
		Version:  parts[2],
		Schema:   parts[0] + "_" + parts[1],
	}, nil
}

func (m packageMeta) outputDir() string {
	return fmt.Sprintf("internal/%s/%s/%s", m.Provider, m.Domain, m.Version)
}

// schemaData is the template data for schema.sql.tmpl.
type schemaData struct {
	Schema string
	Tables []tableData
}

type tableData struct {
	Name    string
	Columns []string
}

func renderSchema(meta packageMeta, messages []*protogen.Message) (string, error) {
	data := schemaData{Schema: meta.Schema}

	for _, msg := range messages {
		table := tableData{Name: toSnakeCase(string(msg.Desc.Name()))}

		// Inlined entity columns.
		for _, col := range entityColumns() {
			table.Columns = append(table.Columns, col.ColumnSQL())
		}

		// Remaining fields (skip entity).
		for _, field := range msg.Fields {
			if string(field.Desc.Name()) == "entity" {
				continue
			}
			col := pgtype.MapField(field.Desc)
			if field.Oneof != nil && !field.Desc.HasOptionalKeyword() {
				col.Nullable = true
			}
			table.Columns = append(table.Columns, col.ColumnSQL())
		}

		data.Tables = append(data.Tables, table)
	}

	var buf bytes.Buffer
	if err := sqlcTemplates.ExecuteTemplate(&buf, "schema.sql.tmpl", data); err != nil {
		return "", fmt.Errorf("executing schema template: %w", err)
	}
	return buf.String(), nil
}

func entityColumns() []pgtype.Column {
	return []pgtype.Column{
		{Name: "id", Type: "UUID PRIMARY KEY"},
		{Name: "created_at", Type: "TIMESTAMPTZ"},
		{Name: "updated_at", Type: "TIMESTAMPTZ"},
	}
}

func isEntityMessage(msg *protogen.Message) bool {
	opts := msg.Desc.Options()
	if opts == nil {
		return false
	}
	if !proto.HasExtension(opts, pluginV1.E_Message) {
		return false
	}
	ext := proto.GetExtension(opts, pluginV1.E_Message)
	clarityOpts, ok := ext.(*pluginV1.ClarityMessageOptions)
	if !ok || clarityOpts == nil {
		return false
	}
	return clarityOpts.GetRole() == pluginV1.Role_ROLE_ENTITY
}

func toSnakeCase(s string) string {
	var b strings.Builder
	for i, r := range s {
		if unicode.IsUpper(r) {
			if i > 0 {
				b.WriteRune('_')
			}
			b.WriteRune(unicode.ToLower(r))
		} else {
			b.WriteRune(r)
		}
	}
	return b.String()
}
