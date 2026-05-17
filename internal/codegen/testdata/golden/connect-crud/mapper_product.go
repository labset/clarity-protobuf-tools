package api

import (
	"github.com/gofrs/uuid/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"google.golang.org/protobuf/types/known/timestamppb"

	pluginv1 "github.com/labset/clarity-protobuf-tools/api/clarity/plugin/v1"
	"github.com/acme/app/internal/acme/inventory/v1/db"
	inventoryv1 "github.com/acme/inventory/v1"
)

func productToProto(row db.AcmeInventoryV1Product) *inventoryv1.Product {
	return &inventoryv1.Product{
		Entity: &pluginv1.Entity{
			Id:        row.ID.String(),
			CreatedAt: timestamppb.New(row.CreatedAt.Time),
			UpdatedAt: timestamppb.New(row.UpdatedAt.Time),
		},
		Name:  row.Name,
		Price: row.Price,
	}
}

func productFromCreate(msg *inventoryv1.Product, id uuid.UUID, now pgtype.Timestamptz) db.CreateAcmeInventoryV1ProductParams {
	return db.CreateAcmeInventoryV1ProductParams{
		ID:        id,
		CreatedAt: now,
		UpdatedAt: now,
		Name:      msg.GetName(),
		Price:     msg.GetPrice(),
	}
}

func productFromUpdate(msg *inventoryv1.Product, id uuid.UUID) db.UpdateAcmeInventoryV1ProductParams {
	return db.UpdateAcmeInventoryV1ProductParams{
		ID:    id,
		Name:  msg.GetName(),
		Price: msg.GetPrice(),
	}
}
