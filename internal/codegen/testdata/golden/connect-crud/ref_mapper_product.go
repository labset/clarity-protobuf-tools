package api

import (
	"github.com/gofrs/uuid/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/acme/app/internal/acme/inventory/v1/db"
	inventoryv1 "github.com/acme/inventory/v1"
	pluginv1 "github.com/labset/clarity-protobuf-tools/api/clarity/plugin/v1"
)

func productToProto(row db.AcmeInventoryV1Product) *inventoryv1.Product {
	return &inventoryv1.Product{
		Entity: &pluginv1.Entity{
			Id:        row.ID.String(),
			CreatedAt: timestamppb.New(row.CreatedAt.Time),
			UpdatedAt: timestamppb.New(row.UpdatedAt.Time),
		},
		Category: &inventoryv1.CategoryRef{Id: row.CategoryID.String()},
		Supplier: &inventoryv1.SupplierRef{Id: row.SupplierID.String()},
		Name:     row.Name,
	}
}

func productFromCreate(msg *inventoryv1.Product, id uuid.UUID, now pgtype.Timestamptz) db.CreateProductParams {
	return db.CreateProductParams{
		ID:         id,
		CreatedAt:  now,
		UpdatedAt:  now,
		CategoryID: uuid.FromStringOrNil(msg.GetCategory().GetId()),
		SupplierID: uuid.FromStringOrNil(msg.GetSupplier().GetId()),
		Name:       msg.GetName(),
	}
}

func productFromUpdate(msg *inventoryv1.Product, id uuid.UUID) db.UpdateProductParams {
	return db.UpdateProductParams{
		ID:         id,
		CategoryID: uuid.FromStringOrNil(msg.GetCategory().GetId()),
		SupplierID: uuid.FromStringOrNil(msg.GetSupplier().GetId()),
		Name:       msg.GetName(),
	}
}
