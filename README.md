# clarity-protobuf-tools

A toolchain for working with [Clarity](https://github.com/labset)-annotated protobuf messages. It provides:

- **clarity-lint-plugin** — a buf lint plugin that validates proto messages annotated with clarity options
- **protoc-gen-clarity** — a protoc code generation plugin with multiple modes

## Proto Annotations

Messages are annotated using `clarity.plugin.v1.ClarityMessageOptions`:

```protobuf
import "clarity/plugin/v1/options.proto";
import "clarity/plugin/v1/entity.proto";

message Product {
  option (clarity.plugin.v1.message) = {role: ROLE_ENTITY};

  clarity.plugin.v1.Entity entity = 1;
  string name = 2;
  int64 price = 3;
}
```

### Roles

| Role | Description |
|------|-------------|
| `ROLE_ENTITY` | Marks a message as a database entity — must have an `entity` field of type `clarity.plugin.v1.Entity` at field number 1 |

## Codegen Modes

### sqlc

Generates PostgreSQL schema, [sqlc](https://sqlc.dev/) queries, and sqlc config from `ROLE_ENTITY` messages.

```
protoc --clarity_out=. --clarity_opt=mode=sqlc,output_dir=. proto/*.proto
```

For a message in package `acme.inventory.v1`, this generates:

```
internal/acme/inventory/v1/
├── sql/
│   ├── schema.sql        # CREATE SCHEMA + CREATE TABLE statements
│   └── queries/
│       └── product.sql   # sqlc-annotated CRUD queries (get, list, create, update, delete)
└── sqlc.yaml             # sqlc config targeting PostgreSQL with pgx
```

#### Type Mapping

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
mise run buf:lint      # lint proto files
mise run buf:format    # format proto files
mise run lint          # lint Go code
mise run build         # build binaries (includes tests)
mise run install       # install plugins locally
```
