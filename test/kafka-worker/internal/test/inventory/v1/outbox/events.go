package outbox

import (
	"time"

	"github.com/gofrs/uuid/v5"
	"github.com/riverqueue/river"
)

type CreateProductEventArgs struct {
	EntityID   uuid.UUID `json:"entity_id"`
	OccurredAt time.Time `json:"occurred_at"`
}

func (CreateProductEventArgs) Kind() string      { return "create_product" }
func (CreateProductEventArgs) InsertOpts() river.InsertOpts { return river.InsertOpts{} }

type UpdateProductEventArgs struct {
	EntityID   uuid.UUID `json:"entity_id"`
	FieldMask  []string  `json:"field_mask"`
	OccurredAt time.Time `json:"occurred_at"`
}

func (UpdateProductEventArgs) Kind() string      { return "update_product" }
func (UpdateProductEventArgs) InsertOpts() river.InsertOpts { return river.InsertOpts{} }

type DeleteProductEventArgs struct {
	EntityID   uuid.UUID `json:"entity_id"`
	OccurredAt time.Time `json:"occurred_at"`
}

func (DeleteProductEventArgs) Kind() string      { return "delete_product" }
func (DeleteProductEventArgs) InsertOpts() river.InsertOpts { return river.InsertOpts{} }
