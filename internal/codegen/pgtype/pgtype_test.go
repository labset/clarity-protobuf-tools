package pgtype

import (
	"testing"

	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protodesc"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/reflect/protoregistry"
	"google.golang.org/protobuf/types/descriptorpb"
)

func fieldDesc(t *testing.T, fds *descriptorpb.FileDescriptorProto) protoreflect.MessageDescriptor {
	t.Helper()
	fd, err := protodesc.NewFile(fds, new(protoregistry.Files))
	if err != nil {
		t.Fatalf("failed to create file descriptor: %v", err)
	}
	return fd.Messages().Get(0)
}

func TestMapField_Scalars(t *testing.T) {
	tests := []struct {
		name     string
		protoTyp descriptorpb.FieldDescriptorProto_Type
		want     string
	}{
		{"string", descriptorpb.FieldDescriptorProto_TYPE_STRING, "TEXT"},
		{"bytes", descriptorpb.FieldDescriptorProto_TYPE_BYTES, "BYTEA"},
		{"bool", descriptorpb.FieldDescriptorProto_TYPE_BOOL, "BOOLEAN"},
		{"int32", descriptorpb.FieldDescriptorProto_TYPE_INT32, "INTEGER"},
		{"sint32", descriptorpb.FieldDescriptorProto_TYPE_SINT32, "INTEGER"},
		{"sfixed32", descriptorpb.FieldDescriptorProto_TYPE_SFIXED32, "INTEGER"},
		{"uint32", descriptorpb.FieldDescriptorProto_TYPE_UINT32, "INTEGER"},
		{"fixed32", descriptorpb.FieldDescriptorProto_TYPE_FIXED32, "INTEGER"},
		{"int64", descriptorpb.FieldDescriptorProto_TYPE_INT64, "BIGINT"},
		{"sint64", descriptorpb.FieldDescriptorProto_TYPE_SINT64, "BIGINT"},
		{"sfixed64", descriptorpb.FieldDescriptorProto_TYPE_SFIXED64, "BIGINT"},
		{"uint64", descriptorpb.FieldDescriptorProto_TYPE_UINT64, "BIGINT"},
		{"fixed64", descriptorpb.FieldDescriptorProto_TYPE_FIXED64, "BIGINT"},
		{"float", descriptorpb.FieldDescriptorProto_TYPE_FLOAT, "REAL"},
		{"double", descriptorpb.FieldDescriptorProto_TYPE_DOUBLE, "DOUBLE PRECISION"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			msg := fieldDesc(t, &descriptorpb.FileDescriptorProto{
				Name:    proto.String("test.proto"),
				Package: proto.String("test"),
				Syntax:  proto.String("proto3"),
				MessageType: []*descriptorpb.DescriptorProto{
					{
						Name: proto.String("TestMessage"),
						Field: []*descriptorpb.FieldDescriptorProto{
							{
								Name:   proto.String("field"),
								Number: proto.Int32(1),
								Type:   tt.protoTyp.Enum(),
							},
						},
					},
				},
			})
			col := MapField(msg.Fields().Get(0))
			if col.Type != tt.want {
				t.Errorf("MapField() type = %q, want %q", col.Type, tt.want)
			}
			if col.Name != "field" {
				t.Errorf("MapField() name = %q, want %q", col.Name, "field")
			}
		})
	}
}

func TestMapField_Enum(t *testing.T) {
	msg := fieldDesc(t, &descriptorpb.FileDescriptorProto{
		Name:    proto.String("test.proto"),
		Package: proto.String("test"),
		Syntax:  proto.String("proto3"),
		EnumType: []*descriptorpb.EnumDescriptorProto{
			{
				Name: proto.String("Status"),
				Value: []*descriptorpb.EnumValueDescriptorProto{
					{Name: proto.String("STATUS_UNSPECIFIED"), Number: proto.Int32(0)},
					{Name: proto.String("STATUS_ACTIVE"), Number: proto.Int32(1)},
					{Name: proto.String("STATUS_INACTIVE"), Number: proto.Int32(2)},
				},
			},
		},
		MessageType: []*descriptorpb.DescriptorProto{
			{
				Name: proto.String("TestMessage"),
				Field: []*descriptorpb.FieldDescriptorProto{
					{
						Name:     proto.String("status"),
						Number:   proto.Int32(1),
						Type:     descriptorpb.FieldDescriptorProto_TYPE_ENUM.Enum(),
						TypeName: proto.String(".test.Status"),
					},
				},
			},
		},
	})
	col := MapField(msg.Fields().Get(0))
	if col.Type != "TEXT" {
		t.Errorf("MapField() type = %q, want TEXT", col.Type)
	}
	wantCheck := "status IN ('STATUS_UNSPECIFIED', 'STATUS_ACTIVE', 'STATUS_INACTIVE')"
	if col.Check != wantCheck {
		t.Errorf("MapField() check = %q, want %q", col.Check, wantCheck)
	}
}

func buildFileWithDep(t *testing.T, dep *descriptorpb.FileDescriptorProto, main *descriptorpb.FileDescriptorProto) protoreflect.MessageDescriptor {
	t.Helper()
	reg := new(protoregistry.Files)
	depFD, err := protodesc.NewFile(dep, reg)
	if err != nil {
		t.Fatalf("failed to create dep file descriptor: %v", err)
	}
	reg.RegisterFile(depFD)
	mainFD, err := protodesc.NewFile(main, reg)
	if err != nil {
		t.Fatalf("failed to create main file descriptor: %v", err)
	}
	return mainFD.Messages().Get(0)
}

func TestMapField_WellKnownTypes(t *testing.T) {
	tests := []struct {
		name     string
		typeName string
		depFile  string
		want     string
	}{
		{"timestamp", ".google.protobuf.Timestamp", "google/protobuf/timestamp.proto", "TIMESTAMPTZ"},
		{"duration", ".google.protobuf.Duration", "google/protobuf/duration.proto", "INTERVAL"},
		{"struct", ".google.protobuf.Struct", "google/protobuf/struct.proto", "JSONB"},
		{"value", ".google.protobuf.Value", "google/protobuf/value.proto", "JSONB"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dep := &descriptorpb.FileDescriptorProto{
				Name:    proto.String(tt.depFile),
				Package: proto.String("google.protobuf"),
				Syntax:  proto.String("proto3"),
				MessageType: []*descriptorpb.DescriptorProto{
					{Name: proto.String(tt.typeName[len(".google.protobuf."):])},
				},
			}
			main := &descriptorpb.FileDescriptorProto{
				Name:       proto.String("test.proto"),
				Package:    proto.String("test"),
				Syntax:     proto.String("proto3"),
				Dependency: []string{tt.depFile},
				MessageType: []*descriptorpb.DescriptorProto{
					{
						Name: proto.String("TestMessage"),
						Field: []*descriptorpb.FieldDescriptorProto{
							{
								Name:     proto.String("field"),
								Number:   proto.Int32(1),
								Type:     descriptorpb.FieldDescriptorProto_TYPE_MESSAGE.Enum(),
								TypeName: proto.String(tt.typeName),
							},
						},
					},
				},
			}
			msg := buildFileWithDep(t, dep, main)
			col := MapField(msg.Fields().Get(0))
			if col.Type != tt.want {
				t.Errorf("MapField() type = %q, want %q", col.Type, tt.want)
			}
		})
	}
}

func TestMapField_NestedMessage(t *testing.T) {
	dep := &descriptorpb.FileDescriptorProto{
		Name:    proto.String("other.proto"),
		Package: proto.String("test"),
		Syntax:  proto.String("proto3"),
		MessageType: []*descriptorpb.DescriptorProto{
			{Name: proto.String("Address")},
		},
	}
	main := &descriptorpb.FileDescriptorProto{
		Name:       proto.String("test.proto"),
		Package:    proto.String("test"),
		Syntax:     proto.String("proto3"),
		Dependency: []string{"other.proto"},
		MessageType: []*descriptorpb.DescriptorProto{
			{
				Name: proto.String("TestMessage"),
				Field: []*descriptorpb.FieldDescriptorProto{
					{
						Name:     proto.String("address"),
						Number:   proto.Int32(1),
						Type:     descriptorpb.FieldDescriptorProto_TYPE_MESSAGE.Enum(),
						TypeName: proto.String(".test.Address"),
					},
				},
			},
		},
	}
	msg := buildFileWithDep(t, dep, main)
	col := MapField(msg.Fields().Get(0))
	if col.Type != "JSONB" {
		t.Errorf("MapField() type = %q, want JSONB", col.Type)
	}
}

func TestMapField_RepeatedScalar(t *testing.T) {
	msg := fieldDesc(t, &descriptorpb.FileDescriptorProto{
		Name:    proto.String("test.proto"),
		Package: proto.String("test"),
		Syntax:  proto.String("proto3"),
		MessageType: []*descriptorpb.DescriptorProto{
			{
				Name: proto.String("TestMessage"),
				Field: []*descriptorpb.FieldDescriptorProto{
					{
						Name:   proto.String("tags"),
						Number: proto.Int32(1),
						Type:   descriptorpb.FieldDescriptorProto_TYPE_STRING.Enum(),
						Label:  descriptorpb.FieldDescriptorProto_LABEL_REPEATED.Enum(),
					},
				},
			},
		},
	})
	col := MapField(msg.Fields().Get(0))
	if col.Type != "TEXT[]" {
		t.Errorf("MapField() type = %q, want TEXT[]", col.Type)
	}
}

func TestMapField_RepeatedMessage(t *testing.T) {
	dep := &descriptorpb.FileDescriptorProto{
		Name:    proto.String("other.proto"),
		Package: proto.String("test"),
		Syntax:  proto.String("proto3"),
		MessageType: []*descriptorpb.DescriptorProto{
			{Name: proto.String("Item")},
		},
	}
	main := &descriptorpb.FileDescriptorProto{
		Name:       proto.String("test.proto"),
		Package:    proto.String("test"),
		Syntax:     proto.String("proto3"),
		Dependency: []string{"other.proto"},
		MessageType: []*descriptorpb.DescriptorProto{
			{
				Name: proto.String("TestMessage"),
				Field: []*descriptorpb.FieldDescriptorProto{
					{
						Name:     proto.String("items"),
						Number:   proto.Int32(1),
						Type:     descriptorpb.FieldDescriptorProto_TYPE_MESSAGE.Enum(),
						TypeName: proto.String(".test.Item"),
						Label:    descriptorpb.FieldDescriptorProto_LABEL_REPEATED.Enum(),
					},
				},
			},
		},
	}
	msg := buildFileWithDep(t, dep, main)
	col := MapField(msg.Fields().Get(0))
	if col.Type != "JSONB" {
		t.Errorf("MapField() type = %q, want JSONB", col.Type)
	}
}

func TestMapField_Map(t *testing.T) {
	msg := fieldDesc(t, &descriptorpb.FileDescriptorProto{
		Name:    proto.String("test.proto"),
		Package: proto.String("test"),
		Syntax:  proto.String("proto3"),
		MessageType: []*descriptorpb.DescriptorProto{
			{
				Name: proto.String("TestMessage"),
				Field: []*descriptorpb.FieldDescriptorProto{
					{
						Name:     proto.String("metadata"),
						Number:   proto.Int32(1),
						Type:     descriptorpb.FieldDescriptorProto_TYPE_MESSAGE.Enum(),
						TypeName: proto.String(".test.TestMessage.MetadataEntry"),
						Label:    descriptorpb.FieldDescriptorProto_LABEL_REPEATED.Enum(),
					},
				},
				NestedType: []*descriptorpb.DescriptorProto{
					{
						Name: proto.String("MetadataEntry"),
						Field: []*descriptorpb.FieldDescriptorProto{
							{
								Name:   proto.String("key"),
								Number: proto.Int32(1),
								Type:   descriptorpb.FieldDescriptorProto_TYPE_STRING.Enum(),
							},
							{
								Name:   proto.String("value"),
								Number: proto.Int32(2),
								Type:   descriptorpb.FieldDescriptorProto_TYPE_STRING.Enum(),
							},
						},
						Options: &descriptorpb.MessageOptions{
							MapEntry: proto.Bool(true),
						},
					},
				},
			},
		},
	})
	col := MapField(msg.Fields().Get(0))
	if col.Type != "JSONB" {
		t.Errorf("MapField() type = %q, want JSONB", col.Type)
	}
}

func TestColumnSQL(t *testing.T) {
	tests := []struct {
		name string
		col  Column
		want string
	}{
		{
			"not null",
			Column{Name: "name", Type: "TEXT"},
			"name TEXT NOT NULL",
		},
		{
			"nullable",
			Column{Name: "name", Type: "TEXT", Nullable: true},
			"name TEXT",
		},
		{
			"with check",
			Column{Name: "status", Type: "TEXT", Check: "status IN ('A', 'B')"},
			"status TEXT NOT NULL CHECK (status IN ('A', 'B'))",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.col.ColumnSQL()
			if got != tt.want {
				t.Errorf("ColumnSQL() = %q, want %q", got, tt.want)
			}
		})
	}
}
