package middler

import (
	"context"
	"fmt"
	"github.com/go-kratos/kratos/v2/transport/http"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/apiserver/pkg/endpoints/handlers/responsewriters"
	"k8s.io/apiserver/pkg/endpoints/request"
	nhttp "net/http"
)

type RequestInfo struct {
	ClusterName string
	request.RequestInfo
}

func ParseRequest(req *nhttp.Request) (*RequestInfo, error) {
	factory := request.RequestInfoFactory{
		APIPrefixes:          sets.NewString("apis", "api"),
		GrouplessAPIPrefixes: sets.NewString("api"),
	}
	info, err := factory.NewRequestInfo(req)
	if err != nil {
		return nil, err
	}
	clusterName := req.URL.Query().Get("clusters")
	reqInfo := &RequestInfo{
		ClusterName: clusterName,
		RequestInfo: *info,
	}
	return reqInfo, nil
}

type requestInfoKeyType int

// requestInfoKey is the RequestInfo key for the context. It's of private type here. Because
// keys are interfaces and interfaces are equal when the type and the value is equal, this
// does not conflict with the keys defined in pkg/api.
const requestInfoKey requestInfoKeyType = iota

// WithRequestInfo returns a copy of parent in which the request info value is set
func WithRequestInfo(parent context.Context, info *RequestInfo) context.Context {
	return context.WithValue(parent, requestInfoKey, info)
}

// RequestInfoFrom returns the value of the RequestInfo key on the ctx
func RequestInfoFrom(ctx context.Context) (*RequestInfo, bool) {
	info, ok := ctx.Value(requestInfoKey).(*RequestInfo)
	return info, ok
}

func ParseRequestForKubernetes() http.FilterFunc {
	return func(handler nhttp.Handler) nhttp.Handler {
		return nhttp.HandlerFunc(func(w nhttp.ResponseWriter, r *nhttp.Request) {
			ctx := r.Context()
			_, ok := RequestInfoFrom(ctx)
			if !ok {
				info, err := ParseRequest(r)
				if err != nil {
					// TODO print log @梧桐
					responsewriters.InternalError(w, r, fmt.Errorf("failed to crate RequestInfo: %v", err))
					return
				}
				r = r.WithContext(WithRequestInfo(ctx, info))
			}
			handler.ServeHTTP(w, r)
		})
	}
}
