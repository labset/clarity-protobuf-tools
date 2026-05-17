package clarity

import (
	pluginV1 "github.com/labset/clarity-protobuf-tools/api/clarity/plugin/v1"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/descriptorpb"
)

// messageRole returns the clarity Role for a message descriptor, or ROLE_UNSPECIFIED
// if the message has no clarity options.
func messageRole(md protoreflect.MessageDescriptor) pluginV1.Role {
	opts, ok := md.Options().(*descriptorpb.MessageOptions)
	if !ok {
		return pluginV1.Role_ROLE_UNSPECIFIED
	}
	if !proto.HasExtension(opts, pluginV1.E_Message) {
		return pluginV1.Role_ROLE_UNSPECIFIED
	}
	ext := proto.GetExtension(opts, pluginV1.E_Message)
	clarityOpts, ok := ext.(*pluginV1.ClarityMessageOptions)
	if !ok || clarityOpts == nil {
		return pluginV1.Role_ROLE_UNSPECIFIED
	}
	return clarityOpts.GetRole()
}

// IsEntity returns true if the message has ROLE_ENTITY.
func IsEntity(md protoreflect.MessageDescriptor) bool {
	return messageRole(md) == pluginV1.Role_ROLE_ENTITY
}

// Operations returns the operations configured for a message descriptor.
func Operations(md protoreflect.MessageDescriptor) []pluginV1.Operation {
	opts, ok := md.Options().(*descriptorpb.MessageOptions)
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
	return clarityOpts.GetOperations()
}
