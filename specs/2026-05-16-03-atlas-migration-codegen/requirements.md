# Requirements: Atlas Migration Codegen

> Add an `atlas-sqlc` codegen mode that combines sqlc schema generation with Atlas migration configuration, and update the schema naming convention to include version.

## Context

The `protoc-gen-clarity` plugin currently supports a `sqlc` mode that generates SQL schema, queries, and sqlc config from proto entity definitions. To support database migrations via [Atlas](https://atlasgo.io/), we need a new mode that generates Atlas project configuration (`atlas.hcl`) alongside the sqlc output. Atlas will then handle migration diffing and application against the generated schema.

The current schema naming convention (`<provider>_<domain>`) does not include the version segment, which should be corrected across all modes.

## Requirements

### Functional

- [ ] FR-1: Update `packageMeta.Schema` to use `<provider>_<domain>_<version>` format (e.g., `acme_inventory_v1`) across all existing sqlc output
- [ ] FR-2: Add a new `atlas-sqlc` mode to `protoc-gen-clarity` that runs the sqlc generator and additionally generates Atlas configuration
- [ ] FR-3: Generate `atlas.hcl` at `<outputDir>/atlas.hcl` with:
  - Schema source referencing `sql/schema.sql` and `sql/baseline.sql`
  - Migration directory at `migrations/`
  - Dev database URL: `docker://postgres/17-alpine/dev?search_path=<provider>_<domain>_<version>`
- [ ] FR-4: Generate `sql/baseline.sql` containing `CREATE SCHEMA IF NOT EXISTS public;` to prevent Atlas from dropping the public schema during diffs
- [ ] FR-5: The `atlas.hcl` template lives under `internal/codegen/templates/atlas-sqlc/`
- [ ] FR-6: The `baseline.sql` template lives under `internal/codegen/templates/atlas-sqlc/`

### Non-Functional

- [ ] NFR-1: Follow existing codegen conventions — `generator_` file prefix, embedded templates, golden file tests
- [ ] NFR-2: The `atlas-sqlc` generator must reuse the sqlc generator rather than duplicating its logic

### Deferred

- [ ] DFR-1: Running `atlas migrate diff` as part of the plugin — Atlas CLI handles this separately
- [ ] DFR-2: Standalone `atlas` mode without sqlc (if needed in the future)

## Constraints

- Atlas requires a dev database URL for computing schema diffs; `docker://postgres/17-alpine` is the chosen provider
- The `baseline.sql` must be included as an Atlas schema source so that `public` schema is not dropped during diffs

## Out of Scope

- Atlas CLI installation or execution
- Migration file generation (Atlas handles this)
- Changes to the lint plugin or proto definitions
