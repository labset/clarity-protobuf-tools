package consumers

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/segmentio/kafka-go"

	"github.com/acme/app/internal/acme/inventory/v1/workers"
)

// ProductIndexHandler handles index events for Product.
type ProductIndexHandler interface {
	HandleProductEvent(ctx context.Context, envelope workers.EventEnvelope) error
}

// RunProductIndexConsumer starts a Kafka consumer for Product index events.
func RunProductIndexConsumer(ctx context.Context, brokers []string, handler ProductIndexHandler) error {
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers: brokers,
		Topic:   "inventory.product.events.v1",
		GroupID: "inventory.product.index",
	})
	defer reader.Close()

	for {
		msg, err := reader.ReadMessage(ctx)
		if err != nil {
			return err
		}
		var envelope workers.EventEnvelope
		if err := json.Unmarshal(msg.Value, &envelope); err != nil {
			return fmt.Errorf("unmarshaling product event: %w", err)
		}
		if err := handler.HandleProductEvent(ctx, envelope); err != nil {
			return fmt.Errorf("handling product event: %w", err)
		}
	}
}
