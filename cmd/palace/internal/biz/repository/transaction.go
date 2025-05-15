package repository

import (
	"context"
)

type Transaction interface {
	MainExec(ctx context.Context, fn func(ctx context.Context) error) error
	BizExec(ctx context.Context, fn func(ctx context.Context) error) error
	EventExec(ctx context.Context, fn func(ctx context.Context) error) error
}
