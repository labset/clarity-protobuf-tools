package workers

import (
	"github.com/riverqueue/river"
	"github.com/segmentio/kafka-go"
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
