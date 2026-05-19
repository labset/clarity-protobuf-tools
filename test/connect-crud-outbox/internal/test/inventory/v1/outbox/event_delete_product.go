package outbox

import (
	"time"

	"github.com/gofrs/uuid/v5"
	"github.com/riverqueue/river"
)

// DeleteProductEventArgs contains the River job arguments for a Product deletion event.
type DeleteProductEventArgs struct {
	EntityID   uuid.UUID `json:"entity_id"`
	OccurredAt time.Time `json:"occurred_at"`
}

func (DeleteProductEventArgs) Kind() string {
	return "delete_product"
}

func (DeleteProductEventArgs) InsertOpts() river.InsertOpts {
	return river.InsertOpts{}
}
