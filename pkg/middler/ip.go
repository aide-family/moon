package middler

import (
	"context"
	"github.com/go-kratos/kratos/v2/middleware"
	"github.com/go-kratos/kratos/v2/transport"
	"github.com/go-kratos/kratos/v2/transport/http"
	"github.com/prometheus/client_golang/prometheus"
	nhttp "net/http"
)

func IpMetric(counter *prometheus.CounterVec) middleware.Middleware {
	return func(handler middleware.Handler) middleware.Handler {
		return func(ctx context.Context, req interface{}) (interface{}, error) {
			if tp, ok := transport.FromServerContext(ctx); ok {
				counter.WithLabelValues(tp.RequestHeader().Get("RemoteAddr")).Inc()
			}
			return handler(ctx, req)
		}
	}
}

func LocalHttpRequestFilter() http.FilterFunc {
	return func(next nhttp.Handler) nhttp.Handler {
		return nhttp.HandlerFunc(func(w nhttp.ResponseWriter, req *nhttp.Request) {
			// 获取请求IP
			remoteAddr := req.RemoteAddr
			req.Header.Set("RemoteAddr", remoteAddr)
			next.ServeHTTP(w, req)
		})
	}
}
