package workers

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/segmentio/kafka-go"
)

// ProductAuditHandler handles audit events for Product.
type ProductAuditHandler interface {
	HandleProductEvent(ctx context.Context, envelope EventEnvelope) error
}

// ProductIndexHandler handles index events for Product.
type ProductIndexHandler interface {
	HandleProductEvent(ctx context.Context, envelope EventEnvelope) error
}

// ProductConsumerDeps holds dependencies for Product Kafka consumers.
type ProductConsumerDeps struct {
	Brokers []string
}

// RunProductAuditConsumer starts a Kafka consumer for Product audit events.
func RunProductAuditConsumer(ctx context.Context, deps ProductConsumerDeps, handler ProductAuditHandler) error {
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers: deps.Brokers,
		Topic:   productTopic,
		GroupID: "inventory.product.audit",
	})
	defer reader.Close()
	return consumeProductEvents(ctx, reader, handler.HandleProductEvent)
}

// RunProductIndexConsumer starts a Kafka consumer for Product index events.
func RunProductIndexConsumer(ctx context.Context, deps ProductConsumerDeps, handler ProductIndexHandler) error {
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers: deps.Brokers,
		Topic:   productTopic,
		GroupID: "inventory.product.index",
	})
	defer reader.Close()
	return consumeProductEvents(ctx, reader, handler.HandleProductEvent)
}

func consumeProductEvents(ctx context.Context, reader *kafka.Reader, handle func(context.Context, EventEnvelope) error) error {
	for {
		msg, err := reader.ReadMessage(ctx)
		if err != nil {
			return err
		}
		var envelope EventEnvelope
		if err := json.Unmarshal(msg.Value, &envelope); err != nil {
			return fmt.Errorf("unmarshaling product event: %w", err)
		}
		if err := handle(ctx, envelope); err != nil {
			return fmt.Errorf("handling product event: %w", err)
		}
	}
}
