package rules

import (
	"context"
	"path"

	"buf.build/go/bufplugin/check"
	"buf.build/go/bufplugin/check/checkutil"
	pluginV1 "github.com/labset/clarity-protobuf-tools/api/clarity/plugin/v1"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/descriptorpb"
)

var entityFileRuleSpec = &check.RuleSpec{
	ID:      "CLARITY_ENTITY_FILE",
	Default: true,
	Purpose: "Checks that ROLE_ENTITY is only used on messages defined in models.proto files.",
	Type:    check.RuleTypeLint,
	Handler: checkutil.NewMessageRuleHandler(checkEntityFile, checkutil.WithoutImports()),
}

func checkEntityFile(
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

	fileName := path.Base(string(messageDescriptor.ParentFile().Path()))
	if fileName != "models.proto" {
		responseWriter.AddAnnotation(
			check.WithMessagef(
				"Message %q with ROLE_ENTITY must be defined in a models.proto file, got %q.",
				messageDescriptor.FullName(),
				fileName,
			),
			check.WithDescriptor(messageDescriptor),
		)
	}

	return nil
}
