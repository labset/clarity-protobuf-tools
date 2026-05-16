env "local" {
  src = [
    "file://sql/baseline.sql",
    "file://sql/schema.sql",
  ]
  dev = "docker://postgres/17-alpine/dev?search_path=acme_inventory_v1"
  migration {
    dir = "file://migrations"
  }
}
