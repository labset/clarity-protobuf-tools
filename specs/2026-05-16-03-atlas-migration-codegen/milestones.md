# Milestones: Atlas Migration Codegen

## M1: Update schema naming to include version

- **Status:** done
- **Description:** Change `packageMeta.Schema` from `<provider>_<domain>` to `<provider>_<domain>_<version>` and update all downstream references (schema.sql, queries, sqlc.yaml templates). Update golden files to reflect the new naming.
- **Acceptance Criteria:**
  - [x] `packageMeta.Schema` returns `<provider>_<domain>_<version>`
  - [x] Generated schema.sql uses versioned schema name
  - [x] Generated queries reference versioned schema name
  - [x] Golden tests pass with updated expectations

## M2: Add atlas-sqlc mode with atlas.hcl and baseline.sql generation

- **Status:** done
- **Description:** Implement a new `atlas-sqlc` generator that delegates to the sqlc generator and additionally produces `atlas.hcl` and `sql/baseline.sql`. Register the mode in `GeneratorForMode`.
- **Acceptance Criteria:**
  - [x] `--mode=atlas-sqlc` is accepted by the plugin
  - [x] `atlas.hcl` is generated with correct source, migration dir, and dev URL
  - [x] `sql/baseline.sql` is generated with `CREATE SCHEMA IF NOT EXISTS public;`
  - [x] Templates live under `internal/codegen/templates/atlas-sqlc/`
  - [x] Golden file tests cover the new output
  - [x] Existing `sqlc` mode continues to work unchanged
