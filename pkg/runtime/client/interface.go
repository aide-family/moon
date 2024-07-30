package client

import (
	"context"

	"github.com/aide-family/moon/pkg/runtime"
	"github.com/aide-family/moon/pkg/runtime/meta"
)

type Object interface {
	meta.Object
	runtime.Object
}

type Reader interface {
	Get(context context.Context, key string, object Object) error
	List(context context.Context, object []Object) error
}

type Writer interface {
	Create(context context.Context, object Object) error
	Update(context context.Context, object Object) error
	Delete(context context.Context, object Object) error
}

type Client interface {
	Reader
	Writer
	Scheme() *runtime.Scheme
}
