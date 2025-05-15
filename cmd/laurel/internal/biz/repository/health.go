package repository

import "context"

type Health interface {
	PingCache(ctx context.Context) error
}
