package rules

import (
	"context"
	"path"

	"buf.build/go/bufplugin/check"
	"buf.build/go/bufplugin/check/checkutil"
	"github.com/labset/clarity-protobuf-tools/internal/protoutil"
	"google.golang.org/protobuf/reflect/protoreflect"
)

var refMessageRuleSpec = &check.RuleSpec{
	ID:      "LABSET_REF_MESSAGE",
	Default: true,
	Purpose: "Checks that every ROLE_ENTITY message in models.proto has a corresponding <Model>Ref message with ROLE_REFERENCE in refs.proto within the same package.",
	Type:    check.RuleTypeLint,
	Handler: checkutil.NewMessageRuleHandler(checkRefMessage, checkutil.WithoutImports()),
}

func checkRefMessage(
	_ context.Context,
	responseWriter check.ResponseWriter,
	request check.Request,
	messageDescriptor protoreflect.MessageDescriptor,
) error {
	if !protoutil.IsEntity(messageDescriptor) {
		return nil
	}

	fileName := path.Base(string(messageDescriptor.ParentFile().Path()))
	if fileName != "models.proto" {
		return nil
	}

	entityPkg := messageDescriptor.ParentFile().Package()
	refName := string(messageDescriptor.Name()) + "Ref"

	var refsFile protoreflect.FileDescriptor
	for _, fd := range request.FileDescriptors() {
		pfd := fd.ProtoreflectFileDescriptor()
		if pfd.Package() == entityPkg && path.Base(string(pfd.Path())) == "refs.proto" {
			refsFile = pfd
			break
		}
	}

	if refsFile == nil {
		responseWriter.AddAnnotation(
			check.WithMessagef(
				"Entity %q requires a %s message with ROLE_REFERENCE in refs.proto, but no refs.proto found in package %q.",
				messageDescriptor.FullName(),
				refName,
				entityPkg,
			),
			check.WithDescriptor(messageDescriptor),
		)
		return nil
	}

	var refMsg protoreflect.MessageDescriptor
	messages := refsFile.Messages()
	for i := range messages.Len() {
		msg := messages.Get(i)
		if string(msg.Name()) == refName {
			refMsg = msg
			break
		}
	}

	if refMsg == nil {
		responseWriter.AddAnnotation(
			check.WithMessagef(
				"Entity %q requires a %s message in refs.proto, but it was not found.",
				messageDescriptor.FullName(),
				refName,
			),
			check.WithDescriptor(messageDescriptor),
		)
		return nil
	}

	if !protoutil.IsReference(refMsg) {
		responseWriter.AddAnnotation(
			check.WithMessagef(
				"Message %s.%s in refs.proto must have ROLE_REFERENCE.",
				entityPkg,
				refName,
			),
			check.WithDescriptor(messageDescriptor),
		)
	}

	idField := refMsg.Fields().ByName("id")
	if idField == nil || idField.Kind() != protoreflect.StringKind {
		responseWriter.AddAnnotation(
			check.WithMessagef(
				"Message %s.%s must have a string field named \"id\".",
				entityPkg,
				refName,
			),
			check.WithDescriptor(messageDescriptor),
		)
	}

	return nil
}
