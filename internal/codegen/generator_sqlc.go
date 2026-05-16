package codegen

import (
	"bytes"
	"embed"
	"fmt"
	"path"
	"strings"
	"text/template"
	"unicode"

	"github.com/labset/clarity-protobuf-tools/internal/clarity"
	"github.com/labset/clarity-protobuf-tools/internal/codegen/pgtype"
	"google.golang.org/protobuf/compiler/protogen"
)

//go:embed templates/sqlc/*.tmpl
var sqlcTemplateFS embed.FS

var sqlcTemplates = template.Must(template.ParseFS(sqlcTemplateFS, "templates/sqlc/*.tmpl"))

type sqlcGenerator struct {
	outputDir string
}

// packageEntities groups entity messages and metadata by package.
type packageEntities struct {
	meta     packageMeta
	messages []*protogen.Message
}

func (g *sqlcGenerator) Generate(plugin *protogen.Plugin) error {
	// Aggregate entity messages across files by package.
	pkgMap := make(map[string]*packageEntities)
	var pkgOrder []string

	for _, file := range plugin.Files {
		if !file.Generate {
			continue
		}
		if path.Base(file.Desc.Path()) != "models.proto" {
			continue
		}

		pkg := string(file.Desc.Package())
		pe, ok := pkgMap[pkg]
		if !ok {
			meta, err := parsePackage(pkg)
			if err != nil {
				return err
			}
			pe = &packageEntities{meta: meta}
			pkgMap[pkg] = pe
			pkgOrder = append(pkgOrder, pkg)
		}

		for _, msg := range file.Messages {
			if isEntityMessage(msg) {
				pe.messages = append(pe.messages, msg)
			}
		}
	}

	// Emit files per package.
	for _, pkg := range pkgOrder {
		pe := pkgMap[pkg]
		if len(pe.messages) == 0 {
			continue
		}

		outDir := pe.meta.outputDir()
		if g.outputDir != "" {
			outDir = fmt.Sprintf("%s/%s", g.outputDir, outDir)
		}

		schemaContent, err := renderSchema(pe.meta, pe.messages)
		if err != nil {
			return err
		}
		if _, err := plugin.NewGeneratedFile(fmt.Sprintf("%s/sql/schema.sql", outDir), "").Write([]byte(schemaContent)); err != nil {
			return err
		}

		for _, msg := range pe.messages {
			queryFileName := toSnakeCase(string(msg.Desc.Name()))
			p := fmt.Sprintf("%s/sql/queries/%s.sql", outDir, queryFileName)
			queryContent, err := renderQueries(pe.meta, msg)
			if err != nil {
				return err
			}
			if _, err := plugin.NewGeneratedFile(p, "").Write([]byte(queryContent)); err != nil {
				return err
			}
		}

		configContent, err := renderConfig(pe.meta)
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
		return packageMeta{}, fmt.Errorf(
			"expected package format <provider>.<domain>.<version>, got %q",
			pkg,
		)
	}
	return packageMeta{
		Provider: parts[0],
		Domain:   parts[1],
		Version:  parts[2],
		Schema:   parts[0] + "_" + parts[1] + "_" + parts[2],
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
	InsertColumns   string
	InsertNamedArgs string
	UpdateSetClause string
}

// insertColumns are columns excluded from INSERT (auto-managed).
var insertExcluded = map[string]bool{
	"deleted_at": true,
}

func renderQueries(meta packageMeta, msg *protogen.Message) (string, error) {
	tableName := toSnakeCase(string(msg.Desc.Name()))

	var allColumns []string
	for _, col := range entityColumns() {
		allColumns = append(allColumns, col.Name)
	}
	for _, field := range msg.Fields {
		if string(field.Desc.Name()) == "entity" {
			continue
		}
		allColumns = append(allColumns, string(field.Desc.Name()))
	}

	// Insert excludes auto-managed columns.
	var insertCols []string
	var insertArgs []string
	for _, col := range allColumns {
		if insertExcluded[col] {
			continue
		}
		insertCols = append(insertCols, col)
		insertArgs = append(insertArgs, fmt.Sprintf("@%s", col))
	}

	// Update sets user-provided columns (excludes managed columns).
	var setClauses []string
	for _, col := range allColumns {
		if managedColumns[col] {
			continue
		}
		setClauses = append(setClauses, fmt.Sprintf("%s = @%s", col, col))
	}
	setClauses = append(setClauses, "updated_at = NOW()")

	data := queryData{
		Schema:          meta.Schema,
		Table:           tableName,
		MessageName:     string(msg.Desc.Name()),
		AllColumns:      strings.Join(allColumns, ", "),
		InsertColumns:   strings.Join(insertCols, ", "),
		InsertNamedArgs: strings.Join(insertArgs, ", "),
		UpdateSetClause: strings.Join(setClauses, ", "),
	}

	var buf bytes.Buffer
	if err := sqlcTemplates.ExecuteTemplate(&buf, "queries.sql.tmpl", data); err != nil {
		return "", fmt.Errorf("executing queries template: %w", err)
	}
	return buf.String(), nil
}

func renderConfig(meta packageMeta) (string, error) {
	var buf bytes.Buffer
	if err := sqlcTemplates.ExecuteTemplate(&buf, "sqlc.yaml.tmpl", meta); err != nil {
		return "", fmt.Errorf("executing sqlc config template: %w", err)
	}
	return buf.String(), nil
}

func entityColumns() []pgtype.Column {
	return []pgtype.Column{
		{Name: "id", Type: "UUID PRIMARY KEY"},
		{Name: "created_at", Type: "TIMESTAMPTZ"},
		{Name: "updated_at", Type: "TIMESTAMPTZ"},
		{Name: "deleted_at", Type: "TIMESTAMPTZ", Nullable: true},
	}
}

// managedColumns are columns excluded from UPDATE SET clauses.
var managedColumns = map[string]bool{
	"id":         true,
	"created_at": true,
	"updated_at": true,
	"deleted_at": true,
}

func isEntityMessage(msg *protogen.Message) bool {
	return clarity.IsEntity(msg.Desc)
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
