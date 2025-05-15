package middler

import (
	"net/http"
	"strings"

	"github.com/gorilla/handlers"

	"github.com/aide-family/moon/pkg/config"
)

// Cors Cross-domain middleware
func Cors(c *config.HTTPServer) func(http.Handler) http.Handler {
	return handlers.CORS(
		handlers.AllowedOriginValidator(func(origin string) bool {
			for _, s := range c.GetAllowOrigins() {
				if strings.EqualFold("*", s) || strings.EqualFold(origin, s) {
					return true
				}
			}
			return false
		}),
		handlers.AllowedHeaders(c.GetAllowHeaders()),
		handlers.AllowedMethods(c.GetAllowMethods()),
		handlers.AllowCredentials(),
	)
}
