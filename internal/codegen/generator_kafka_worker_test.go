package codegen

import (
	"os"
	"testing"

	pluginV1 "github.com/labset/clarity-protobuf-tools/api/clarity/plugin/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/compiler/protogen"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/descriptorpb"
	"google.golang.org/protobuf/types/pluginpb"
)

func loadKafkaWorkerGolden(t *testing.T, name string) string {
	t.Helper()
	data, err := os.ReadFile("testdata/golden/kafka-worker/" + name)
	require.NoError(t, err)
	return string(data)
}

func newTestKafkaWorkerGenerator(outputDir string) *kafkaWorkerGenerator {
	return &kafkaWorkerGenerator{
		goModule:  "github.com/acme/app",
		outputDir: outputDir,
	}
}

func testKafkaWorkerProtoFile(
	t *testing.T,
	ops []pluginV1.Operation,
	subs []pluginV1.Subscriber,
) *descriptorpb.FileDescriptorProto {
	t.Helper()
	mopts := &descriptorpb.MessageOptions{}
	proto.SetExtension(mopts, pluginV1.E_Message, &pluginV1.ClarityMessageOptions{
		Role:        pluginV1.Role_ROLE_ENTITY,
		Operations:  ops,
		Subscribers: subs,
	})
	return &descriptorpb.FileDescriptorProto{
		Name:    proto.String("acme/inventory/v1/models.proto"),
		Package: proto.String("acme.inventory.v1"),
		Syntax:  proto.String("proto3"),
		Options: &descriptorpb.FileOptions{
			GoPackage: proto.String("github.com/acme/inventory/v1;inventoryv1"),
		},
		Dependency: []string{
			"labset/options/v1/options.proto",
			"labset/data/v1/entity.proto",
		},
		EnumType: []*descriptorpb.EnumDescriptorProto{
			{
				Name: proto.String("ProductStatus"),
				Value: []*descriptorpb.EnumValueDescriptorProto{
					{Name: proto.String("PRODUCT_STATUS_UNSPECIFIED"), Number: proto.Int32(0)},
					{Name: proto.String("PRODUCT_STATUS_ACTIVE"), Number: proto.Int32(1)},
					{Name: proto.String("PRODUCT_STATUS_ARCHIVED"), Number: proto.Int32(2)},
				},
			},
		},
		MessageType: []*descriptorpb.DescriptorProto{
			{
				Name:    proto.String("Product"),
				Options: mopts,
				Field: []*descriptorpb.FieldDescriptorProto{
					{
						Name:     proto.String("entity"),
						Number:   proto.Int32(1),
						Type:     descriptorpb.FieldDescriptorProto_TYPE_MESSAGE.Enum(),
						TypeName: proto.String(".clarity.plugin.v1.Entity"),
					},
					{
						Name:   proto.String("name"),
						Number: proto.Int32(2),
						Type:   descriptorpb.FieldDescriptorProto_TYPE_STRING.Enum(),
					},
					{
						Name:   proto.String("price"),
						Number: proto.Int32(3),
						Type:   descriptorpb.FieldDescriptorProto_TYPE_INT64.Enum(),
					},
					{
						Name:     proto.String("status"),
						Number:   proto.Int32(4),
						Type:     descriptorpb.FieldDescriptorProto_TYPE_ENUM.Enum(),
						TypeName: proto.String(".acme.inventory.v1.ProductStatus"),
					},
				},
			},
		},
	}
}

func runKafkaWorkerGenerator(
	t *testing.T,
	gen *kafkaWorkerGenerator,
	ops []pluginV1.Operation,
	subs []pluginV1.Subscriber,
) map[string]string {
	t.Helper()
	deps := collectFileDescriptors(t,
		"labset/options/v1/options.proto",
		"labset/data/v1/entity.proto",
	)
	modelsFile := testKafkaWorkerProtoFile(t, ops, subs)
	req := &pluginpb.CodeGeneratorRequest{
		FileToGenerate: []string{"acme/inventory/v1/models.proto"},
		ProtoFile:      append(deps, modelsFile),
	}
	plugin, err := protogen.Options{}.New(req)
	require.NoError(t, err)

	err = gen.Generate(plugin)
	require.NoError(t, err)

	resp := plugin.Response()
	require.NotNil(t, resp)

	files := make(map[string]string)
	for _, f := range resp.GetFile() {
		files[f.GetName()] = f.GetContent()
	}
	return files
}

func TestKafkaWorkerGenerator_AllOperations(t *testing.T) {
	allOps := []pluginV1.Operation{
		pluginV1.Operation_OPERATION_CREATE,
		pluginV1.Operation_OPERATION_GET,
		pluginV1.Operation_OPERATION_LIST,
		pluginV1.Operation_OPERATION_UPDATE,
		pluginV1.Operation_OPERATION_DELETE,
	}
	subs := []pluginV1.Subscriber{
		pluginV1.Subscriber_SUBSCRIBER_AUDIT,
		pluginV1.Subscriber_SUBSCRIBER_INDEX,
	}

	files := runKafkaWorkerGenerator(t, newTestKafkaWorkerGenerator(""), allOps, subs)

	// envelope + register + 3 workers + 2 consumers = 7
	assert.Len(t, files, 7)

	// Workers
	assert.Equal(t,
		loadKafkaWorkerGolden(t, "envelope.go"),
		files["internal/acme/inventory/v1/workers/envelope.go"],
	)
	assert.Equal(t,
		loadKafkaWorkerGolden(t, "register_product.go"),
		files["internal/acme/inventory/v1/workers/register_product.go"],
	)
	assert.Equal(t,
		loadKafkaWorkerGolden(t, "worker_create_product.go"),
		files["internal/acme/inventory/v1/workers/worker_create_product.go"],
	)
	assert.Equal(t,
		loadKafkaWorkerGolden(t, "worker_update_product.go"),
		files["internal/acme/inventory/v1/workers/worker_update_product.go"],
	)
	assert.Equal(t,
		loadKafkaWorkerGolden(t, "worker_delete_product.go"),
		files["internal/acme/inventory/v1/workers/worker_delete_product.go"],
	)

	// Consumers
	assert.Equal(t,
		loadKafkaWorkerGolden(t, "consumer_audit_product.go"),
		files["internal/acme/inventory/v1/consumers/consumer_audit_product.go"],
	)
	assert.Equal(t,
		loadKafkaWorkerGolden(t, "consumer_index_product.go"),
		files["internal/acme/inventory/v1/consumers/consumer_index_product.go"],
	)
}

func TestKafkaWorkerGenerator_NoSubscribers(t *testing.T) {
	ops := []pluginV1.Operation{
		pluginV1.Operation_OPERATION_CREATE,
		pluginV1.Operation_OPERATION_GET,
	}
	files := runKafkaWorkerGenerator(t, newTestKafkaWorkerGenerator(""), ops, nil)
	assert.Len(t, files, 0)
}

func TestKafkaWorkerGenerator_CreateOnly(t *testing.T) {
	ops := []pluginV1.Operation{pluginV1.Operation_OPERATION_CREATE}
	subs := []pluginV1.Subscriber{pluginV1.Subscriber_SUBSCRIBER_AUDIT}

	files := runKafkaWorkerGenerator(t, newTestKafkaWorkerGenerator(""), ops, subs)

	// envelope + register + 1 worker + 1 consumer = 4
	assert.Len(t, files, 4)
	assert.Contains(t, files, "internal/acme/inventory/v1/workers/register_product.go")
	assert.Contains(t, files, "internal/acme/inventory/v1/workers/worker_create_product.go")
	assert.NotContains(t, files, "internal/acme/inventory/v1/workers/worker_update_product.go")
	assert.NotContains(t, files, "internal/acme/inventory/v1/workers/worker_delete_product.go")
	assert.Contains(t, files, "internal/acme/inventory/v1/consumers/consumer_audit_product.go")
	assert.NotContains(t, files, "internal/acme/inventory/v1/consumers/consumer_index_product.go")
}

func TestKafkaWorkerGenerator_ReadOnlyOpsWithSubscribers(t *testing.T) {
	ops := []pluginV1.Operation{
		pluginV1.Operation_OPERATION_GET,
		pluginV1.Operation_OPERATION_LIST,
	}
	subs := []pluginV1.Subscriber{pluginV1.Subscriber_SUBSCRIBER_AUDIT}

	files := runKafkaWorkerGenerator(t, newTestKafkaWorkerGenerator(""), ops, subs)
	assert.Len(t, files, 0)
}

func TestKafkaWorkerGenerator_OutputDir(t *testing.T) {
	ops := []pluginV1.Operation{pluginV1.Operation_OPERATION_CREATE}
	subs := []pluginV1.Subscriber{pluginV1.Subscriber_SUBSCRIBER_AUDIT}

	files := runKafkaWorkerGenerator(t, newTestKafkaWorkerGenerator("custom/out"), ops, subs)

	assert.Contains(t, files, "custom/out/internal/acme/inventory/v1/workers/register_product.go")
	assert.Contains(
		t,
		files,
		"custom/out/internal/acme/inventory/v1/workers/worker_create_product.go",
	)
	assert.Contains(t, files, "custom/out/internal/acme/inventory/v1/workers/envelope.go")
	assert.Contains(
		t,
		files,
		"custom/out/internal/acme/inventory/v1/consumers/consumer_audit_product.go",
	)

	workerContent := files["custom/out/internal/acme/inventory/v1/workers/worker_create_product.go"]
	assert.Contains(t, workerContent,
		"\"github.com/acme/app/custom/out/internal/acme/inventory/v1/outbox\"",
	)

	consumerContent := files["custom/out/internal/acme/inventory/v1/consumers/consumer_audit_product.go"]
	assert.Contains(t, consumerContent,
		"\"github.com/acme/app/custom/out/internal/acme/inventory/v1/workers\"",
	)
}
