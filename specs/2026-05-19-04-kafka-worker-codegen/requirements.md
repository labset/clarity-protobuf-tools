# Requirements: Kafka Worker Codegen

> Generate River workers that publish entity mutation events to Kafka topics, with consumer group stubs driven by a proto-level subscriber annotation.

## Context

The existing `connect-crud-outbox` codegen mode generates transactional outbox events using River queue. These events are enqueued atomically alongside data mutations but currently require manually written River workers to process them. Real-world usage demands fan-out to multiple independent consumers (audit logging, search indexing, webhook emission, notifications) — each with independent failure modes, retry policies, and scaling characteristics.

Kafka consumer groups are the natural fit for this pattern. A single River worker publishes events to a per-entity Kafka topic, and independent consumer groups process them. A new `subscribers` annotation on `ClarityMessageOptions` declares which consumer groups an entity requires, driving codegen of both the Kafka-publishing River worker and typed consumer group stubs.

## Requirements

### Functional

- [x] FR-1: Extend `ClarityMessageOptions` with a `repeated Subscriber subscribers` field and a `Subscriber` enum (`AUDIT`, `INDEX`, `WEBHOOK`, `NOTIFICATION`)
- [x] FR-2: Introduce a new codegen mode `kafka-worker` that reads the `subscribers` annotation from entity messages
- [x] FR-3: Generate a River worker per entity that consumes the outbox event and publishes to a Kafka topic named `<domain>.<entity_snake>.events.<version>` (e.g. `inventory.product.events.v1`)
- [x] FR-4: Kafka message envelope includes: entity ID, operation (create/update/delete), field mask (for updates), occurred_at timestamp, and a `Payload` field reserved for future entity payload inclusion
- [x] FR-5: Kafka message key is the entity ID to guarantee per-entity ordering within a partition
- [x] FR-6: Generate a typed consumer handler interface per subscriber per entity (e.g. `ProductAuditHandler`, `ProductIndexHandler`)
- [x] FR-7: Generate consumer group registration that wires handler implementations to a Kafka consumer with group ID `<domain>.<entity_snake>.<subscriber_snake>` (e.g. `inventory.product.audit`)
- [x] FR-8: If `subscribers` is empty or absent on an entity, the `kafka-worker` mode produces no output for that entity

### Non-Functional

- [x] NFR-1: Follow existing codegen conventions — `generator_` file prefix, `embed` for templates, golden file tests
- [x] NFR-2: Templates live under `internal/codegen/templates/kafka-worker/`
- [x] NFR-3: Golden test files scoped under `internal/codegen/testdata/golden/kafka-worker/`
- [x] NFR-4: Use `segmentio/kafka-go` as the Kafka client (pure Go, no CGO dependency)

### Deferred

- [ ] DFR-1: Schema registry integration (Confluent Schema Registry / protobuf schema evolution) — evaluate once event format stabilises
- [ ] DFR-2: Dead-letter topic codegen for failed consumer processing
- [ ] DFR-3: Consumer retry policies and backoff configuration
- [ ] DFR-4: Populate entity payload in Kafka envelope — requires worker to fetch entity from DB via store, adding store/mapper dependencies to the generated worker

## Constraints

- River worker receives the existing outbox event args (EntityID, OccurredAt, FieldMask) — no changes to `connect-crud-outbox` output
- Consumer handler implementations are manual — codegen produces interfaces and registration only
- One Kafka topic per entity, not per subscriber or per operation
- The `subscribers` annotation is only meaningful when the entity also has mutating operations

## Out of Scope

- Kafka cluster provisioning or topic creation
- Consumer deployment or runtime orchestration
- Kafka producer/consumer configuration (auth, batching, compression) — left to application-level wiring
- Avro or protobuf-binary encoding (proto-JSON for now)
