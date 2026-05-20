package outbox

import (
	"time"

	"github.com/gofrs/uuid/v5"
	"github.com/riverqueue/river"
)

// UpdateProductEventArgs contains the River job arguments for a Product update event.
type UpdateProductEventArgs struct {
	EntityID   uuid.UUID `json:"entity_id"`
	FieldMask  []string  `json:"field_mask,omitempty"`
	OccurredAt time.Time `json:"occurred_at"`
}

func (UpdateProductEventArgs) Kind() string {
	return "update_product"
}

func (UpdateProductEventArgs) InsertOpts() river.InsertOpts {
	return river.InsertOpts{}
}
