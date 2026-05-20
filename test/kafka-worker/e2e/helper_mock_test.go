package e2e

import (
	"context"

	"github.com/labset/clarity-protobuf-tools/test/kafka-worker/internal/test/inventory/v1/workers"
)

type mockAuditHandler struct{}

func (m *mockAuditHandler) HandleProductEvent(_ context.Context, _ workers.EventEnvelope) error {
	return nil
}

type mockIndexHandler struct{}

func (m *mockIndexHandler) HandleProductEvent(_ context.Context, _ workers.EventEnvelope) error {
	return nil
}
