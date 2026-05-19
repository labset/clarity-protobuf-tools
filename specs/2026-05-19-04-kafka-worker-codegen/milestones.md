# Milestones: Kafka Worker Codegen

## M1: Subscriber annotation

- **Status:** pending
- **Description:** Extend the proto options with the Subscriber enum and repeated field on ClarityMessageOptions. Regenerate Go bindings.
- **Acceptance Criteria:**
  - [ ] `Subscriber` enum added with values: `SUBSCRIBER_UNSPECIFIED`, `SUBSCRIBER_AUDIT`, `SUBSCRIBER_INDEX`, `SUBSCRIBER_WEBHOOK`, `SUBSCRIBER_NOTIFICATION`
  - [ ] `repeated Subscriber subscribers` field added to `ClarityMessageOptions`
  - [ ] Go bindings regenerated under `api/clarity/plugin/v1/`
  - [ ] Build passes (`mise run build`)

## M2: River worker codegen

- **Status:** pending
- **Description:** Generate a River worker per entity that consumes the outbox event and publishes to a Kafka topic. The worker is only generated when the entity has subscribers declared.
- **Acceptance Criteria:**
  - [ ] New `kafka-worker` codegen mode registered in the plugin
  - [ ] River worker template consumes outbox event args and publishes to Kafka
  - [ ] Topic naming follows `<domain>.<entity_snake>.events.<version>` convention
  - [ ] Kafka message key is the entity ID
  - [ ] Kafka message envelope includes: entity ID, operation, field mask, occurred_at, proto-JSON payload
  - [ ] No output generated when `subscribers` is empty or absent
  - [ ] Golden file tests pass for generated worker files

## M3: Consumer group stubs

- **Status:** pending
- **Description:** Generate typed consumer handler interfaces and consumer group registration per subscriber per entity.
- **Acceptance Criteria:**
  - [ ] Handler interface generated per subscriber per entity (e.g. `ProductAuditHandler`, `ProductIndexHandler`)
  - [ ] Consumer group registration generated with group ID `<domain>.<entity_snake>.<subscriber_snake>`
  - [ ] Registration wires handler implementations to a Kafka consumer
  - [ ] Golden file tests pass for generated consumer files

## M4: Integration and build

- **Status:** pending
- **Description:** End-to-end golden file test with a test proto that declares subscribers, validating the complete output tree and full build.
- **Acceptance Criteria:**
  - [ ] Test proto with `subscribers` annotation added under test fixtures
  - [ ] E2E golden file test validates complete output tree
  - [ ] Build passes (`mise run build`)
