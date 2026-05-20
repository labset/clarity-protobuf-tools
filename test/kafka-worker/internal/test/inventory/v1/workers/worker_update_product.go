package workers

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/riverqueue/river"
	"github.com/segmentio/kafka-go"

	"github.com/labset/clarity-protobuf-tools/test/kafka-worker/internal/test/inventory/v1/outbox"
)

type updateProductWorker struct {
	river.WorkerDefaults[outbox.UpdateProductEventArgs]
	writer *kafka.Writer
}

func (w *updateProductWorker) Work(ctx context.Context, job *river.Job[outbox.UpdateProductEventArgs]) error {
	envelope := EventEnvelope{
		EntityID:   job.Args.EntityID.String(),
		Operation:  "update",
		FieldMask:  job.Args.FieldMask,
		OccurredAt: job.Args.OccurredAt,
	}

	data, err := json.Marshal(envelope)
	if err != nil {
		return fmt.Errorf("marshaling update product event: %w", err)
	}

	return w.writer.WriteMessages(ctx, kafka.Message{
		Topic: productTopic,
		Key:   []byte(envelope.EntityID),
		Value: data,
	})
}
