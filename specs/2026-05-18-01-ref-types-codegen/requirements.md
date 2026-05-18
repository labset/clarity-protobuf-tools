# Requirements: Ref Types Codegen

> Add support for reference type fields in proto entities, generating proper foreign key columns, queries, and mapper conversions across the codegen pipeline.

## Context

The current codegen pipeline handles scalar and enum fields on entities but has no concept of inter-entity references. In practice, entities often reference other entities (e.g., a `Product` has a `Category`). By convention, a `refs.proto` file defines `<Model>Ref` wrapper messages annotated with `ROLE_REFERENCE`. Entity fields that use these ref types should produce `_id UUID` columns in the SQL schema, with optional foreign key constraints controlled via a field-level annotation. This needs to flow through the full codegen stack: schema, queries, and mappers.

## Requirements

### Functional

- [ ] FR-1: Add `ROLE_REFERENCE` value to the `Role` enum in `options.proto`
- [ ] FR-2: Add a `ClarityFieldOptions` message with a `bool foreign_key` field and extend `google.protobuf.FieldOptions` in `options.proto`
- [ ] FR-3: Add a lint rule enforcing that for every entity in `models.proto`, a corresponding `<Model>Ref` message with `ROLE_REFERENCE` exists in `refs.proto`
- [ ] FR-4: In SQL schema codegen (`sqlc` mode), ref-type fields produce a `<snake_name>_id UUID NOT NULL` column
- [ ] FR-5: When a ref field has `foreign_key = true`, emit a `REFERENCES <derived_table>(id)` constraint on the column
- [ ] FR-6: The referenced table name is derived from the ref message name (e.g., `CategoryRef` -> `category` table in the same schema)
- [ ] FR-7: In SQLC query codegen, include `_id` columns in insert, update, select, and list queries
- [ ] FR-8: In connect-crud mapper codegen, ref fields map `msg.Get<Ref>().GetId()` to `uuid.UUID` and reverse (`<Ref>Ref{Id: row.<Name>ID.String()}`)

### Non-Functional

- [ ] NFR-1: Follow existing codegen conventions -- `generator_` prefix, embedded templates, golden file tests
- [ ] NFR-2: `ROLE_REFERENCE` messages are only recognised in files named `refs.proto`
- [ ] NFR-3: `foreign_key` defaults to `false` -- no FK constraint unless explicitly opted in

### Deferred

- [ ] DFR-1: Cascade delete behaviour configuration on FK constraints
- [ ] DFR-2: Composite/multi-field references
- [ ] DFR-3: Self-referential entities (e.g., parent category)
- [ ] DFR-4: Ref field nullability (optional references)

## Constraints

- Must remain a standard protoc/buf plugin
- FK target table is always derived from the ref message name, not configurable
- The same `<Model>Ref` can be used by multiple entities with different FK settings

## Out of Scope

- Runtime foreign key resolution / joins
- Service proto codegen changes (ref fields flow through naturally from the entity message)
- Cross-store reference validation
