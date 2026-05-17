# clarity-protobuf-tools

A toolchain for working with [Clarity](https://github.com/labset)-annotated protobuf messages. It provides:

- **clarity-lint-plugin** — a buf lint plugin that validates proto messages annotated with clarity options
- **protoc-gen-clarity** — a protoc code generation plugin with multiple modes

## Quick Start

```bash
# install mise (manages go, buf, golangci-lint, goreleaser)
# see https://mise.jdx.dev/getting-started.html

mise install           # install tool versions
mise run deps          # download Go module dependencies
mise run generate      # generate Go code from proto files
mise run build         # build binaries (includes tests)
mise run install       # install plugins locally
```

### Dev Loop

```bash
# edit proto annotations or codegen templates, then:
mise run build         # compile + test
mise run lint          # lint Go + proto files
mise run lint:fix      # auto-fix lint issues
```

## Proto Annotations

Messages are annotated using `clarity.plugin.v1.ClarityMessageOptions`:

```protobuf
import "clarity/plugin/v1/options.proto";
import "clarity/plugin/v1/entity.proto";

message Product {
  option (clarity.plugin.v1.message) = {
    role: ROLE_ENTITY
    operations: [OPERATION_CREATE, OPERATION_GET, OPERATION_LIST, OPERATION_UPDATE, OPERATION_DELETE]
  };

  clarity.plugin.v1.Entity entity = 1;
  string name = 2;
  int64 price = 3;
}
```

### Roles

| Role | Description |
|------|-------------|
| `ROLE_ENTITY` | Marks a message as a database entity — must have an `entity` field of type `clarity.plugin.v1.Entity` at field number 1 |

### Operations

| Operation | Description |
|-----------|-------------|
| `OPERATION_CREATE` | Generate a create RPC with typed `item` request |
| `OPERATION_GET` | Generate a get-by-id RPC with `buf.validate` UUID constraint |
| `OPERATION_LIST` | Generate a list RPC with cursor-based pagination |
| `OPERATION_UPDATE` | Generate an update RPC with `item` + `FieldMask` for partial updates |
| `OPERATION_DELETE` | Generate a soft-delete RPC with validated id |

Operations are opt-in — only specified operations produce service output.

## Lint Plugin

The `clarity-lint-plugin` validates:

- `ROLE_ENTITY` messages must be in files named `models.proto`
- `ROLE_ENTITY` messages must have an `entity` field of type `clarity.plugin.v1.Entity` at field number 1
- `ROLE_ENTITY` messages must be in a `<provider>.<domain>.<version>` package

## Codegen Modes

### sqlc

Generates PostgreSQL schema, [sqlc](https://sqlc.dev/) queries, and sqlc config from `ROLE_ENTITY` messages.

```
protoc --clarity_out=. --clarity_opt=mode=sqlc proto/*.proto
```

For a message in package `acme.inventory.v1`, generates:

```
internal/acme/inventory/v1/
├── sql/
│   ├── schema.sql        # CREATE SCHEMA + CREATE TABLE
│   └── queries/
│       └── product.sql   # CRUD queries (get, list, create, update, delete, soft_delete)
└── sqlc.yaml             # sqlc config targeting PostgreSQL with pgx
```

### atlas-sqlc

Extends `sqlc` mode with [Atlas](https://atlasgo.io/) migration configuration.

```
protoc --clarity_out=. --clarity_opt=mode=atlas-sqlc proto/*.proto
```

Generates everything from `sqlc` plus:

```
internal/acme/inventory/v1/
├── atlas.hcl             # Atlas project config (schema source, migration dir, dev DB)
└── sql/
    └── baseline.sql      # Prevents Atlas from dropping the public schema
```

### service

Generates `.proto` service definitions from `ROLE_ENTITY` messages with opted-in operations.

```
protoc --clarity_out=. --clarity_opt=mode=service proto/*.proto
```

For a message in package `acme.inventory.v1` with all operations, generates:

```
acme/inventory/v1/
├── service_product.proto       # service ProductService { ... }
├── rpc_create_product.proto    # CreateProductRequest/Response
├── rpc_get_product.proto       # GetProductRequest/Response
├── rpc_list_product.proto      # ListProductsRequest/Response
├── rpc_update_product.proto    # UpdateProductRequest/Response
└── rpc_delete_product.proto    # DeleteProductRequest/Response
```

Generated services follow these conventions:

- **Create/Get/Update** responses use a typed `item` field
- **List** response uses `repeated items` with `next_page_token` for cursor pagination
- **Update** request includes `FieldMask` for partial updates
- **Get/Update/Delete** requests validate `id` with `buf.validate` UUID constraint
- **Delete** is always soft delete at the API layer

### Type Mapping (sqlc/atlas-sqlc)

| Proto Type | PostgreSQL Type |
|------------|----------------|
| `string` | `TEXT` |
| `bytes` | `BYTEA` |
| `bool` | `BOOLEAN` |
| `int32`, `sint32`, `sfixed32`, `uint32`, `fixed32` | `INTEGER` |
| `int64`, `sint64`, `sfixed64`, `uint64`, `fixed64` | `BIGINT` |
| `float` | `REAL` |
| `double` | `DOUBLE PRECISION` |
| `google.protobuf.Timestamp` | `TIMESTAMPTZ` |
| `google.protobuf.Duration` | `INTERVAL` |
| `google.protobuf.Struct` / `Value` | `JSONB` |
| `enum` | `TEXT` with `CHECK` constraint |
| `repeated <scalar>` | Array (e.g. `TEXT[]`) |
| `repeated <message>`, nested message, `map` | `JSONB` |
| `oneof` | Nullable columns per variant |

## Usage with Buf

### buf.yaml

```yaml
version: v2

modules:
  - path: protos

deps:
  - buf.build/labset/clarity-protobuf-tools

lint:
  use:
    - STANDARD
  plugins:
    - plugin: clarity-lint-plugin

breaking:
  use:
    - FILE
```

### buf.gen.yaml

#### sqlc mode

```yaml
version: v2
inputs:
  - directory: protos

plugins:
  - local: protoc-gen-clarity
    out: internal
    opt:
      - mode=sqlc
```

#### atlas-sqlc mode

```yaml
version: v2
inputs:
  - directory: protos

plugins:
  - local: protoc-gen-clarity
    out: internal
    opt:
      - mode=atlas-sqlc
```

#### service mode

Generates `.proto` service definitions from entity messages — output alongside your source protos:

```yaml
version: v2
inputs:
  - directory: protos

plugins:
  - local: protoc-gen-clarity
    out: protos
    opt:
      - mode=service
```

### Running generation

```bash
# generate with a specific config
buf generate --template buf.gen.yaml

# or with multiple configs in sequence
buf generate --template buf.gen.yaml
buf generate --template buf.gen.services.yaml
```

## Development

### Requirements

- [mise](https://mise.jdx.dev/) — manages `go`, `buf`, `golangci-lint` and `goreleaser` versions (see `.mise.toml`)

```bash
mise install
```

### Tasks

```bash
mise run deps          # download Go module dependencies
mise run generate      # generate Go code from proto files
mise run build         # build binaries (includes tests)
mise run install       # install plugins locally
mise run lint          # lint Go + proto files
mise run lint:fix      # auto-fix lint issues
mise run format        # format Go + proto files
mise run buf:lint      # lint proto files only
mise run buf:format    # format proto files only
mise run go:lint       # lint Go code only
mise run go:format     # format Go code only
```
