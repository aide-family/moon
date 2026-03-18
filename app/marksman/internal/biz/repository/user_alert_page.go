package repository

import (
	"context"

	"github.com/bwmarrin/snowflake"
)

// UserAlertPage provides access to a user's followed alert page list (single table).
type UserAlertPage interface {
	// GetUserAlertPageUIDs returns the current user's followed alert page UIDs in order.
	GetUserAlertPageUIDs(ctx context.Context, userUID snowflake.ID) ([]snowflake.ID, error)
	// SaveUserAlertPages replaces the user's followed alert pages with the given UIDs (order by index).
	SaveUserAlertPages(ctx context.Context, userUID snowflake.ID, alertPageUIDs []snowflake.ID) error
}
