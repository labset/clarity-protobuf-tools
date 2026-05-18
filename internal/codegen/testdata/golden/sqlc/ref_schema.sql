CREATE SCHEMA IF NOT EXISTS acme_inventory_v1;

CREATE TABLE IF NOT EXISTS acme_inventory_v1.product (
  id UUID PRIMARY KEY NOT NULL,
  created_at TIMESTAMPTZ NOT NULL,
  updated_at TIMESTAMPTZ NOT NULL,
  deleted_at TIMESTAMPTZ,
  category_id UUID REFERENCES acme_inventory_v1.category(id) NOT NULL,
  supplier_id UUID NOT NULL,
  name TEXT NOT NULL
);
