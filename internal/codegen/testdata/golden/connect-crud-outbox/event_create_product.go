package outbox

import (
	"time"

	"github.com/gofrs/uuid/v5"
	"github.com/riverqueue/river"
)

// CreateProductEventArgs contains the River job arguments for a Product creation event.
type CreateProductEventArgs struct {
	EntityID   uuid.UUID `json:"entity_id"`
	OccurredAt time.Time `json:"occurred_at"`
}

func (CreateProductEventArgs) Kind() string {
	return "create_product"
}

func (CreateProductEventArgs) InsertOpts() river.InsertOpts {
	return river.InsertOpts{}
}
