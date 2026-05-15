package rules

import (
	"context"
	"fmt"

	"buf.build/go/bufplugin/check"
	"buf.build/go/bufplugin/check/checkutil"
	pluginV1 "github.com/labset/go-protoc-gen-plugin/api/clarity/plugin/v1"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/descriptorpb"
)

var entityFieldRuleSpec = &check.RuleSpec{
	ID:      "CLARITY_ENTITY_FIELD",
	Default: true,
	Purpose: "Checks that messages with ROLE_ENTITY have a field named entity of type clarity.plugin.v1.Entity at field number 1.",
	Type:    check.RuleTypeLint,
	Handler: checkutil.NewMessageRuleHandler(checkEntityField, checkutil.WithoutImports()),
}

func checkEntityField(
	_ context.Context,
	responseWriter check.ResponseWriter,
	_ check.Request,
	messageDescriptor protoreflect.MessageDescriptor,
) error {
	opts, ok := messageDescriptor.Options().(*descriptorpb.MessageOptions)
	if !ok {
		return nil
	}
	if !proto.HasExtension(opts, pluginV1.E_Message) {
		return nil
	}
	ext := proto.GetExtension(opts, pluginV1.E_Message)
	clarityOpts, ok := ext.(*pluginV1.ClarityMessageOptions)
	if !ok || clarityOpts == nil {
		return nil
	}
	if clarityOpts.GetRole() != pluginV1.Role_ROLE_ENTITY {
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

	if entityField.Kind() != protoreflect.MessageKind || entityField.Message().FullName() != "clarity.plugin.v1.Entity" {
		responseWriter.AddAnnotation(
			check.WithMessagef(
				"Message %q field \"entity\" must be of type clarity.plugin.v1.Entity, got %s.",
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
	return fmt.Sprintf("%s", fd.Kind())
}
