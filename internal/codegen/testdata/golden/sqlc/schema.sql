CREATE SCHEMA IF NOT EXISTS acme_inventory_v1;

CREATE TABLE IF NOT EXISTS acme_inventory_v1.product (
  id UUID PRIMARY KEY NOT NULL,
  created_at TIMESTAMPTZ NOT NULL,
  updated_at TIMESTAMPTZ NOT NULL,
  deleted_at TIMESTAMPTZ,
  name TEXT NOT NULL,
  price BIGINT NOT NULL
);

CREATE TABLE IF NOT EXISTS acme_inventory_v1.order (
  id UUID PRIMARY KEY NOT NULL,
  created_at TIMESTAMPTZ NOT NULL,
  updated_at TIMESTAMPTZ NOT NULL,
  deleted_at TIMESTAMPTZ,
  quantity INTEGER NOT NULL,
  billing_address TEXT,
  shipping_address TEXT
);
