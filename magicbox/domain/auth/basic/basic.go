// Package basic is the basic auth package for the magicbox service.
package basic

import (
	nethttp "net/http"
	"strings"

	"github.com/aide-family/magicbox/config"
	"github.com/aide-family/magicbox/pointer"
	"github.com/aide-family/magicbox/server/middler"
	klog "github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/transport/http"
)

type HandlerBinding struct {
	Name      string
	Enabled   bool
	BasicAuth *config.BasicAuthConfig
	Handler   nethttp.Handler
	Path      string
}

// BindHandlerWithAuth binds a handler with basic auth.
// If the basic auth is not enabled, it will return the handler without basic auth.
// If the basic auth is enabled, it will return the handler with basic auth.
func BindHandlerWithAuth(httpSrv *http.Server, binding HandlerBinding) {
	if !binding.Enabled {
		klog.Debugf("%s is not enabled", binding.Name)
		return
	}

	handler := binding.Handler
	basicAuth := binding.BasicAuth
	if pointer.IsNotNil(basicAuth) && strings.EqualFold(basicAuth.GetEnabled(), "true") {
		handler = middler.BasicAuthMiddleware(basicAuth.GetUsername(), basicAuth.GetPassword())(handler)
		klog.Debugf("%s route: %s (Basic Auth: %s:%s)", binding.Name, binding.Path, basicAuth.GetUsername(), basicAuth.GetPassword())
	} else {
		klog.Debugf("%s route: %s (No Basic Auth)", binding.Name, binding.Path)
	}

	httpSrv.HandlePrefix(binding.Path, handler)
}
