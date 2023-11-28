package middler

import (
	"context"
	nhttp "net/http"

	"github.com/go-kratos/kratos/v2/transport/http"
)

const (
	ContextPathKey   = "path"
	ContextMethodKey = "method"
)

func Context() http.FilterFunc {
	return func(handler nhttp.Handler) nhttp.Handler {
		return nhttp.HandlerFunc(func(w nhttp.ResponseWriter, r *nhttp.Request) {
			// 把path、method写入context
			ctx := r.Context()
			ctx = context.WithValue(ctx, ContextPathKey, r.URL.Path)
			ctx = context.WithValue(ctx, ContextMethodKey, r.Method)
			r = r.WithContext(ctx)
			handler.ServeHTTP(w, r)
		})
	}
}

func GetPath(ctx context.Context) string {
	return ctx.Value(ContextPathKey).(string)
}

func GetMethod(ctx context.Context) string {
	return ctx.Value(ContextMethodKey).(string)
}
