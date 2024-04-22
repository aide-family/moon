package server

import (
	"errors"
	"fmt"
	clu "github.com/aide-family/moon/app/kubemoon/internal/server/cluster"
	"github.com/aide-family/moon/pkg/helper/middler"
	"github.com/go-kratos/kratos/v2/transport/http"
	"k8s.io/apimachinery/pkg/util/proxy"
	"k8s.io/apiserver/pkg/endpoints/handlers/responsewriters"
	"k8s.io/client-go/rest"
	nhttp "net/http"
	"net/url"
)

func ProxyToKubernetes(set clu.Set) http.FilterFunc {
	return func(handler nhttp.Handler) nhttp.Handler {
		return nhttp.HandlerFunc(func(w nhttp.ResponseWriter, r *nhttp.Request) {
			// TODO print log @梧桐
			info, ok := middler.RequestInfoFrom(r.Context())
			if !ok {
				err := errors.New("unable to retrieve request info from request")
				responsewriters.InternalError(w, r, err)
			}
			if len(info.ClusterName) != 0 {
				cluster := set.Client(info.ClusterName)
				if cluster == nil {
					responsewriters.InternalError(w, r, fmt.Errorf("cluster %s is not exsit", cluster))
					return
				}
				//TODO: need to check cluster support crds
				config := cluster.Config()
				kubernetes, _ := url.Parse(config.Host)
				defaultTransport, err := rest.TransportFor(config)
				if err != nil {
					responsewriters.InternalError(w, r, fmt.Errorf("cluster %s is not exsit", cluster))
					return
				}

				s := *r.URL
				s.Host = kubernetes.Host
				s.Scheme = kubernetes.Scheme
				s.Path = info.Path
				// make sure we don't override kubernetes's authorization
				r.Header.Del("Authorization")
				httpProxy := proxy.NewUpgradeAwareHandler(&s, defaultTransport, true, false, nil)
				httpProxy.UpgradeTransport = proxy.NewUpgradeRequestRoundTripper(defaultTransport, defaultTransport)
				httpProxy.ServeHTTP(w, r)
				return
			}
			handler.ServeHTTP(w, r)
		})
	}
}
