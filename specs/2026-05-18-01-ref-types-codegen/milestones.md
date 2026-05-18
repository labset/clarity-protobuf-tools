# Milestones: Ref Types Codegen

## M1: Proto Options -- ROLE_REFERENCE and Field Options

- **Status:** pending
- **Description:** Add `ROLE_REFERENCE` to the `Role` enum and introduce a `ClarityFieldOptions` message with a `bool foreign_key` field, extending `google.protobuf.FieldOptions`. Regenerate Go code from the updated proto.
- **Acceptance Criteria:**
  - [ ] `ROLE_REFERENCE` value exists in the `Role` enum in `options.proto`
  - [ ] `ClarityFieldOptions` message with `foreign_key` bool field exists in `options.proto`
  - [ ] `google.protobuf.FieldOptions` is extended with `clarity.plugin.v1.field` using `ClarityFieldOptions`
  - [ ] Generated Go code compiles and includes the new types

## M2: Linter -- Enforce refs.proto Convention

- **Status:** pending
- **Description:** Add a lint rule that enforces every entity defined in `models.proto` has a corresponding `<Model>Ref` message with `ROLE_REFERENCE` in `refs.proto` within the same package.
- **Acceptance Criteria:**
  - [ ] New `check_ref_message.go` rule in `internal/rules/`
  - [ ] Rule registered in `checks.go`
  - [ ] Rule reports a diagnostic when an entity in `models.proto` has no matching `<Model>Ref` in `refs.proto`
  - [ ] Rule reports a diagnostic when a ref message exists but lacks `ROLE_REFERENCE`
  - [ ] Rule passes when all entities have valid corresponding refs

## M3: SQL Schema and SQLC Query Codegen

- **Status:** pending
- **Description:** Update the `sqlc` codegen mode so that ref-type fields on entities produce `<snake_name>_id UUID NOT NULL` columns in the schema, with an optional `REFERENCES <table>(id)` constraint when the field is annotated with `foreign_key = true`. Include the `_id` columns in generated SQLC queries.
- **Acceptance Criteria:**
  - [ ] Ref-type fields produce `<snake_name>_id UUID NOT NULL` columns in `schema.sql`
  - [ ] Fields with `foreign_key = true` emit a `REFERENCES <derived_table>(id)` constraint
  - [ ] Fields without `foreign_key` produce a plain UUID column with no constraint
  - [ ] Referenced table name is derived from ref message name (e.g., `CategoryRef` -> `category`)
  - [ ] Insert, update, select, and list queries include the `_id` columns
  - [ ] Golden file tests cover both FK and non-FK ref fields

## M4: Connect-CRUD Mapper Codegen

- **Status:** pending
- **Description:** Update the connect-crud mapper codegen to handle ref-type fields, converting between proto ref messages and UUID columns in the `toProto`, `fromCreate`, and `fromUpdate` mapper functions.
- **Acceptance Criteria:**
  - [ ] `toProto` maps `row.<Name>ID` to `<Ref>Ref{Id: row.<Name>ID.String()}`
  - [ ] `fromCreate` maps `msg.Get<Ref>().GetId()` to `uuid.UUID`
  - [ ] `fromUpdate` maps `msg.Get<Ref>().GetId()` to `uuid.UUID`
  - [ ] Golden file tests cover ref fields in mapper output
