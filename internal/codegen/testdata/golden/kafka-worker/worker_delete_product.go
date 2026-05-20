package workers

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/riverqueue/river"
	"github.com/segmentio/kafka-go"

	"github.com/acme/app/internal/acme/inventory/v1/outbox"
)

type deleteProductWorker struct {
	river.WorkerDefaults[outbox.DeleteProductEventArgs]
	writer *kafka.Writer
}

func (w *deleteProductWorker) Work(ctx context.Context, job *river.Job[outbox.DeleteProductEventArgs]) error {
	envelope := EventEnvelope{
		EntityID:   job.Args.EntityID.String(),
		Operation:  "delete",
		OccurredAt: job.Args.OccurredAt,
	}

	data, err := json.Marshal(envelope)
	if err != nil {
		return fmt.Errorf("marshaling delete product event: %w", err)
	}

	return w.writer.WriteMessages(ctx, kafka.Message{
		Topic: productTopic,
		Key:   []byte(envelope.EntityID),
		Value: data,
	})
}
