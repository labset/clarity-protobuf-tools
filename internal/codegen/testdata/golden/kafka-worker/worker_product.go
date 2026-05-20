package workers

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/riverqueue/river"
	"github.com/segmentio/kafka-go"

	"github.com/acme/app/internal/acme/inventory/v1/outbox"
)

const productTopic = "inventory.product.events.v1"

// ProductWorkerDeps holds dependencies for Product Kafka workers.
type ProductWorkerDeps struct {
	Writer *kafka.Writer
}

// RegisterProductWorkers registers River workers that publish Product outbox events to Kafka.
func RegisterProductWorkers(workers *river.Workers, deps ProductWorkerDeps) {
	river.AddWorker(workers, &createProductWorker{writer: deps.Writer})
	river.AddWorker(workers, &updateProductWorker{writer: deps.Writer})
	river.AddWorker(workers, &deleteProductWorker{writer: deps.Writer})
}

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
