package e2e

import (
	"testing"

	"github.com/riverqueue/river"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/labset/clarity-protobuf-tools/test/kafka-worker/internal/test/inventory/v1/consumers"
	"github.com/labset/clarity-protobuf-tools/test/kafka-worker/internal/test/inventory/v1/workers"
)

func TestRegisterProductWorkers(t *testing.T) {
	w := river.NewWorkers()
	workers.RegisterProductWorkers(w, workers.ProductWorkerDeps{})
	assert.NotNil(t, w)
}

func TestProductConsumerHandlerInterfaces(t *testing.T) {
	var _ consumers.ProductAuditHandler = (*mockAuditHandler)(nil)
	var _ consumers.ProductIndexHandler = (*mockIndexHandler)(nil)

	require.True(t, true, "handler interfaces compile correctly")
}
