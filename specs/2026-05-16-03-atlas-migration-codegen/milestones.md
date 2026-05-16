# Milestones: Atlas Migration Codegen

## M1: Update schema naming to include version

- **Status:** pending
- **Description:** Change `packageMeta.Schema` from `<provider>_<domain>` to `<provider>_<domain>_<version>` and update all downstream references (schema.sql, queries, sqlc.yaml templates). Update golden files to reflect the new naming.
- **Acceptance Criteria:**
  - [ ] `packageMeta.Schema` returns `<provider>_<domain>_<version>`
  - [ ] Generated schema.sql uses versioned schema name
  - [ ] Generated queries reference versioned schema name
  - [ ] Golden tests pass with updated expectations

## M2: Add atlas-sqlc mode with atlas.hcl and baseline.sql generation

- **Status:** pending
- **Description:** Implement a new `atlas-sqlc` generator that delegates to the sqlc generator and additionally produces `atlas.hcl` and `sql/baseline.sql`. Register the mode in `GeneratorForMode`.
- **Acceptance Criteria:**
  - [ ] `--mode=atlas-sqlc` is accepted by the plugin
  - [ ] `atlas.hcl` is generated with correct source, migration dir, and dev URL
  - [ ] `sql/baseline.sql` is generated with `CREATE SCHEMA IF NOT EXISTS public;`
  - [ ] Templates live under `internal/codegen/templates/atlas-sqlc/`
  - [ ] Golden file tests cover the new output
  - [ ] Existing `sqlc` mode continues to work unchanged
