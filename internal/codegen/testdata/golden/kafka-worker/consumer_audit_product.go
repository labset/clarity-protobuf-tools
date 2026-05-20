package consumers

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/segmentio/kafka-go"

	"github.com/acme/app/internal/acme/inventory/v1/workers"
)

// ProductAuditHandler handles audit events for Product.
type ProductAuditHandler interface {
	HandleProductEvent(ctx context.Context, envelope workers.EventEnvelope) error
}

// RunProductAuditConsumer starts a Kafka consumer for Product audit events.
func RunProductAuditConsumer(ctx context.Context, brokers []string, handler ProductAuditHandler) error {
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers: brokers,
		Topic:   "inventory.product.events.v1",
		GroupID: "inventory.product.audit",
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
