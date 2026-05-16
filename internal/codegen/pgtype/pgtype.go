package pgtype

import (
	"fmt"
	"strings"

	"google.golang.org/protobuf/reflect/protoreflect"
)

// Column represents a PostgreSQL column definition.
type Column struct {
	Name     string
	Type     string
	Nullable bool
	Check    string // optional CHECK constraint expression
}

// ColumnSQL returns the column definition as a SQL fragment (e.g. "name TEXT NOT NULL").
func (c Column) ColumnSQL() string {
	var b strings.Builder
	b.WriteString(c.Name)
	b.WriteString(" ")
	b.WriteString(c.Type)
	if !c.Nullable {
		b.WriteString(" NOT NULL")
	}
	if c.Check != "" {
		b.WriteString(" CHECK (")
		b.WriteString(c.Check)
		b.WriteString(")")
	}
	return b.String()
}

// MapField maps a proto field descriptor to a PostgreSQL Column.
// For oneof variant fields, set nullable to true at the call site.
func MapField(fd protoreflect.FieldDescriptor) Column {
	name := string(fd.Name())

	if fd.IsMap() {
		return Column{Name: name, Type: "JSONB"}
	}

	if fd.IsList() {
		base := scalarOrMessageType(fd)
		if fd.Kind() == protoreflect.MessageKind || fd.Kind() == protoreflect.GroupKind {
			return Column{Name: name, Type: "JSONB"}
		}
		return Column{Name: name, Type: base + "[]"}
	}

	if fd.Kind() == protoreflect.EnumKind {
		return enumColumn(fd)
	}

	if fd.Kind() == protoreflect.MessageKind {
		return messageColumn(fd)
	}

	return Column{Name: name, Type: scalarType(fd.Kind())}
}

func enumColumn(fd protoreflect.FieldDescriptor) Column {
	enumDesc := fd.Enum()
	values := enumDesc.Values()
	names := make([]string, values.Len())
	for i := range values.Len() {
		names[i] = fmt.Sprintf("'%s'", values.Get(i).Name())
	}
	return Column{
		Name:  string(fd.Name()),
		Type:  "TEXT",
		Check: fmt.Sprintf("%s IN (%s)", fd.Name(), strings.Join(names, ", ")),
	}
}

func messageColumn(fd protoreflect.FieldDescriptor) Column {
	name := string(fd.Name())
	switch string(fd.Message().FullName()) {
	case "google.protobuf.Timestamp":
		return Column{Name: name, Type: "TIMESTAMPTZ"}
	case "google.protobuf.Duration":
		return Column{Name: name, Type: "INTERVAL"}
	case "google.protobuf.Struct", "google.protobuf.Value":
		return Column{Name: name, Type: "JSONB"}
	case "google.protobuf.StringValue":
		return Column{Name: name, Type: "TEXT", Nullable: true}
	case "google.protobuf.BytesValue":
		return Column{Name: name, Type: "BYTEA", Nullable: true}
	case "google.protobuf.BoolValue":
		return Column{Name: name, Type: "BOOLEAN", Nullable: true}
	case "google.protobuf.Int32Value":
		return Column{Name: name, Type: "INTEGER", Nullable: true}
	case "google.protobuf.UInt32Value":
		return Column{Name: name, Type: "INTEGER", Nullable: true}
	case "google.protobuf.Int64Value":
		return Column{Name: name, Type: "BIGINT", Nullable: true}
	case "google.protobuf.UInt64Value":
		return Column{Name: name, Type: "BIGINT", Nullable: true}
	case "google.protobuf.FloatValue":
		return Column{Name: name, Type: "REAL", Nullable: true}
	case "google.protobuf.DoubleValue":
		return Column{Name: name, Type: "DOUBLE PRECISION", Nullable: true}
	default:
		return Column{Name: name, Type: "JSONB"}
	}
}

func scalarOrMessageType(fd protoreflect.FieldDescriptor) string {
	if fd.Kind() == protoreflect.EnumKind {
		return "TEXT"
	}
	if fd.Kind() == protoreflect.MessageKind || fd.Kind() == protoreflect.GroupKind {
		return "JSONB"
	}
	return scalarType(fd.Kind())
}

func scalarType(k protoreflect.Kind) string {
	switch k {
	case protoreflect.StringKind:
		return "TEXT"
	case protoreflect.BytesKind:
		return "BYTEA"
	case protoreflect.BoolKind:
		return "BOOLEAN"
	case protoreflect.Int32Kind, protoreflect.Sint32Kind, protoreflect.Sfixed32Kind,
		protoreflect.Uint32Kind, protoreflect.Fixed32Kind:
		return "INTEGER"
	case protoreflect.Int64Kind, protoreflect.Sint64Kind, protoreflect.Sfixed64Kind,
		protoreflect.Uint64Kind, protoreflect.Fixed64Kind:
		return "BIGINT"
	case protoreflect.FloatKind:
		return "REAL"
	case protoreflect.DoubleKind:
		return "DOUBLE PRECISION"
	default:
		return "TEXT"
	}
}
