package runtime

import "context"

type Reader interface {
	Get(ctx context.Context, key string, out any) error
	List(ctx context.Context, out any) error
}

type Writer interface {
	Create(ctx context.Context, object any) error
	Update(ctx context.Context, object any) error
	Delete(ctx context.Context, object any) error
}
