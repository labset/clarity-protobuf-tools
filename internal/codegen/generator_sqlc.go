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

type sqlcGenerator struct {
	outputDir string
}

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

		outDir := meta.outputDir()
		if g.outputDir != "" {
			outDir = fmt.Sprintf("%s/%s", g.outputDir, outDir)
		}

		schemaContent, err := renderSchema(meta, entityMessages)
		if err != nil {
			return err
		}
		if _, err := plugin.NewGeneratedFile(fmt.Sprintf("%s/sql/schema.sql", outDir), "").Write([]byte(schemaContent)); err != nil {
			return err
		}

		for _, msg := range entityMessages {
			queryFileName := toSnakeCase(string(msg.Desc.Name()))
			path := fmt.Sprintf("%s/sql/queries/%s.sql", outDir, queryFileName)
			queryContent, err := renderQueries(meta, msg)
			if err != nil {
				return err
			}
			if _, err := plugin.NewGeneratedFile(path, "").Write([]byte(queryContent)); err != nil {
				return err
			}
		}

		configContent, err := renderConfig(meta)
		if err != nil {
			return err
		}
		if _, err := plugin.NewGeneratedFile(fmt.Sprintf("%s/sqlc.yaml", outDir), "").Write([]byte(configContent)); err != nil {
			return err
		}
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

// queryData is the template data for queries.sql.tmpl.
type queryData struct {
	Schema          string
	Table           string
	MessageName     string
	AllColumns      string
	AllPlaceholders string
	UpdateSetClause string
}

func renderQueries(meta packageMeta, msg *protogen.Message) (string, error) {
	tableName := toSnakeCase(string(msg.Desc.Name()))

	var columnNames []string
	for _, col := range entityColumns() {
		columnNames = append(columnNames, col.Name)
	}
	for _, field := range msg.Fields {
		if string(field.Desc.Name()) == "entity" {
			continue
		}
		columnNames = append(columnNames, string(field.Desc.Name()))
	}

	placeholders := make([]string, len(columnNames))
	for i := range columnNames {
		placeholders[i] = fmt.Sprintf("$%d", i+1)
	}

	// Update sets all columns except id, with id as $1.
	var setClauses []string
	paramIdx := 2 // $1 is id in WHERE clause
	for _, col := range columnNames {
		if col == "id" {
			continue
		}
		setClauses = append(setClauses, fmt.Sprintf("%s = $%d", col, paramIdx))
		paramIdx++
	}

	data := queryData{
		Schema:          meta.Schema,
		Table:           tableName,
		MessageName:     string(msg.Desc.Name()),
		AllColumns:      strings.Join(columnNames, ", "),
		AllPlaceholders: strings.Join(placeholders, ", "),
		UpdateSetClause: strings.Join(setClauses, ", "),
	}

	var buf bytes.Buffer
	if err := sqlcTemplates.ExecuteTemplate(&buf, "queries.sql.tmpl", data); err != nil {
		return "", fmt.Errorf("executing queries template: %w", err)
	}
	return buf.String(), nil
}

// configData is the template data for sqlc.yaml.tmpl.
type configData struct {
	Package string
}

func renderConfig(meta packageMeta) (string, error) {
	data := configData{Package: meta.Version}
	var buf bytes.Buffer
	if err := sqlcTemplates.ExecuteTemplate(&buf, "sqlc.yaml.tmpl", data); err != nil {
		return "", fmt.Errorf("executing sqlc config template: %w", err)
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
