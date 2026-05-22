package rules

import (
	"context"
	"path"
	"strings"

	"buf.build/go/bufplugin/check"
	"buf.build/go/bufplugin/check/checkutil"
	"github.com/labset/clarity-protobuf-tools/internal/protoutil"
	"google.golang.org/protobuf/reflect/protoreflect"
)

var entityFileRuleSpec = &check.RuleSpec{
	ID:      "LABSET_ENTITY_FILE",
	Default: true,
	Purpose: "Checks that ROLE_ENTITY is only used on messages defined in models.proto files under a <provider>.<domain>.<version> package.",
	Type:    check.RuleTypeLint,
	Handler: checkutil.NewMessageRuleHandler(checkEntityFile, checkutil.WithoutImports()),
}

func checkEntityFile(
	_ context.Context,
	responseWriter check.ResponseWriter,
	_ check.Request,
	messageDescriptor protoreflect.MessageDescriptor,
) error {
	if !protoutil.IsEntity(messageDescriptor) {
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

	pkg := string(messageDescriptor.ParentFile().Package())
	parts := strings.Split(pkg, ".")
	if len(parts) < 3 {
		responseWriter.AddAnnotation(
			check.WithMessagef(
				"Message %q with ROLE_ENTITY must be in a <provider>.<domain>.<version> package, got %q.",
				messageDescriptor.FullName(),
				pkg,
			),
			check.WithDescriptor(messageDescriptor),
		)
	}

	return nil
}
