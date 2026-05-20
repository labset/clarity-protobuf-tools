package workers

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/riverqueue/river"
	"github.com/segmentio/kafka-go"

	"github.com/acme/app/internal/acme/inventory/v1/outbox"
)

type createProductWorker struct {
	river.WorkerDefaults[outbox.CreateProductEventArgs]
	writer *kafka.Writer
}

func (w *createProductWorker) Work(ctx context.Context, job *river.Job[outbox.CreateProductEventArgs]) error {
	envelope := EventEnvelope{
		EntityID:   job.Args.EntityID.String(),
		Operation:  "create",
		OccurredAt: job.Args.OccurredAt,
	}

	data, err := json.Marshal(envelope)
	if err != nil {
		return fmt.Errorf("marshaling create product event: %w", err)
	}

	return w.writer.WriteMessages(ctx, kafka.Message{
		Topic: productTopic,
		Key:   []byte(envelope.EntityID),
		Value: data,
	})
}
