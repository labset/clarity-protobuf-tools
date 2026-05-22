package rules

import (
	"context"

	"buf.build/go/bufplugin/check"
	"buf.build/go/bufplugin/check/checkutil"
	"github.com/labset/clarity-protobuf-tools/internal/protoutil"
	"google.golang.org/protobuf/reflect/protoreflect"
)

var entityFieldRuleSpec = &check.RuleSpec{
	ID:      "LABSET_ENTITY_FIELD",
	Default: true,
	Purpose: "Checks that messages with ROLE_ENTITY have a field named entity of type labset.data.v1.Entity at field number 1.",
	Type:    check.RuleTypeLint,
	Handler: checkutil.NewMessageRuleHandler(checkEntityField, checkutil.WithoutImports()),
}

func checkEntityField(
	_ context.Context,
	responseWriter check.ResponseWriter,
	_ check.Request,
	messageDescriptor protoreflect.MessageDescriptor,
) error {
	if !protoutil.IsEntity(messageDescriptor) {
		return nil
	}

	entityField := messageDescriptor.Fields().ByName("entity")
	if entityField == nil {
		responseWriter.AddAnnotation(
			check.WithMessagef(
				"Message %q with ROLE_ENTITY must have a field named \"entity\".",
				messageDescriptor.FullName(),
			),
			check.WithDescriptor(messageDescriptor),
		)
		return nil
	}

	if entityField.Number() != 1 {
		responseWriter.AddAnnotation(
			check.WithMessagef(
				"Message %q field \"entity\" must be at field number 1, got %d.",
				messageDescriptor.FullName(),
				entityField.Number(),
			),
			check.WithDescriptor(entityField),
		)
	}

	if entityField.Kind() != protoreflect.MessageKind ||
		entityField.Message().FullName() != "labset.data.v1.Entity" {
		responseWriter.AddAnnotation(
			check.WithMessagef(
				"Message %q field \"entity\" must be of type labset.data.v1.Entity, got %s.",
				messageDescriptor.FullName(),
				fieldTypeName(entityField),
			),
			check.WithDescriptor(entityField),
		)
	}

	return nil
}

func fieldTypeName(fd protoreflect.FieldDescriptor) string {
	if fd.Kind() == protoreflect.MessageKind || fd.Kind() == protoreflect.GroupKind {
		return string(fd.Message().FullName())
	}
	if fd.Kind() == protoreflect.EnumKind {
		return string(fd.Enum().FullName())
	}
	return fd.Kind().String()
}
