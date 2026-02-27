package middler

import (
	nethttp "net/http"
	"strings"

	"github.com/go-kratos/kratos/v2/transport/http"
	"github.com/gorilla/handlers"
)

type CorsConfig struct {
	AllowOrigins    []string
	AllowHeaders    []string
	AllowMethods    []string
	ExposeHeaders   []string
	MaxAge          int
	AllowOriginFunc func(origin string) bool
}

func defaultAllowOriginFunc(c *CorsConfig) func(origin string) bool {
	return func(origin string) bool {
		for _, s := range c.AllowOrigins {
			if strings.EqualFold("*", s) || strings.EqualFold(origin, s) {
				return true
			}
		}
		return false
	}
}

// Cors Cross-domain middleware
func Cors(c *CorsConfig) func(nethttp.Handler) nethttp.Handler {
	if c.AllowOriginFunc == nil {
		c.AllowOriginFunc = defaultAllowOriginFunc(c)
	}

	return handlers.CORS(
		handlers.AllowedOriginValidator(c.AllowOriginFunc),
		handlers.AllowedHeaders(c.AllowHeaders),
		handlers.AllowedMethods(c.AllowMethods),
		handlers.ExposedHeaders(c.ExposeHeaders),
		handlers.MaxAge(c.MaxAge),
		handlers.AllowCredentials(),
	)
}

func DefaultCors() http.ServerOption {
	return http.Filter(Cors(&CorsConfig{
		AllowOrigins: []string{"*"},
		AllowMethods: []string{nethttp.MethodGet, nethttp.MethodPost, nethttp.MethodPut, nethttp.MethodDelete, nethttp.MethodOptions},
		MaxAge:       600,
	}))
}
