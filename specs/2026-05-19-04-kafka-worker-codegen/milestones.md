# Milestones: Kafka Worker Codegen

## M1: Subscriber annotation

- **Status:** done
- **Description:** Extend the proto options with the Subscriber enum and repeated field on ClarityMessageOptions. Regenerate Go bindings.
- **Acceptance Criteria:**
  - [x] `Subscriber` enum added with values: `SUBSCRIBER_UNSPECIFIED`, `SUBSCRIBER_AUDIT`, `SUBSCRIBER_INDEX`, `SUBSCRIBER_WEBHOOK`, `SUBSCRIBER_NOTIFICATION`
  - [x] `repeated Subscriber subscribers` field added to `ClarityMessageOptions`
  - [x] Go bindings regenerated under `api/clarity/plugin/v1/`
  - [x] Build passes (`mise run build`)

## M2: River worker codegen

- **Status:** done
- **Description:** Generate a River worker per entity that consumes the outbox event and publishes to a Kafka topic. The worker is only generated when the entity has subscribers declared.
- **Acceptance Criteria:**
  - [x] New `kafka-worker` codegen mode registered in the plugin
  - [x] River worker template consumes outbox event args and publishes to Kafka
  - [x] Topic naming follows `<domain>.<entity_snake>.events.<version>` convention
  - [x] Kafka message key is the entity ID
  - [x] Kafka message envelope includes: entity ID, operation, field mask, occurred_at, proto-JSON payload
  - [x] No output generated when `subscribers` is empty or absent
  - [x] Golden file tests pass for generated worker files

## M3: Consumer group stubs

- **Status:** done
- **Description:** Generate typed consumer handler interfaces and consumer group registration per subscriber per entity.
- **Acceptance Criteria:**
  - [x] Handler interface generated per subscriber per entity (e.g. `ProductAuditHandler`, `ProductIndexHandler`)
  - [x] Consumer group registration generated with group ID `<domain>.<entity_snake>.<subscriber_snake>`
  - [x] Registration wires handler implementations to a Kafka consumer
  - [x] Golden file tests pass for generated consumer files

## M4: Integration and build

- **Status:** done
- **Description:** End-to-end golden file test with a test proto that declares subscribers, validating the complete output tree and full build.
- **Acceptance Criteria:**
  - [x] Test proto with `subscribers` annotation added under test fixtures
  - [x] E2E golden file test validates complete output tree
  - [x] Build passes (`mise run build`)
