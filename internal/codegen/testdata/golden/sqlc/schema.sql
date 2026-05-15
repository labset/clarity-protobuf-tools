CREATE SCHEMA IF NOT EXISTS acme_inventory;

CREATE TABLE acme_inventory.product (
  id UUID PRIMARY KEY NOT NULL,
  created_at TIMESTAMPTZ NOT NULL,
  updated_at TIMESTAMPTZ NOT NULL,
  name TEXT NOT NULL,
  price BIGINT NOT NULL
);
