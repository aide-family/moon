package repository

import (
	"context"

	"github.com/bwmarrin/snowflake"
)

// AlertPushDedup tracks recently pushed firing alerts to avoid duplicate rabbit notifications.
type AlertPushDedup interface {
	ExistsFiringPush(ctx context.Context, namespaceUID snowflake.ID, fingerprint string) (bool, error)
	MarkFiringPush(ctx context.Context, namespaceUID snowflake.ID, fingerprint string) error
	ClearFiringPush(ctx context.Context, namespaceUID snowflake.ID, fingerprint string) error
}
