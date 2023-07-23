package middler

import (
	"github.com/gorilla/handlers"
	"net/http"
)

func Cors() func(http.Handler) http.Handler {
	return handlers.CORS(
		handlers.AllowedOriginValidator(func(origin string) bool {
			return true
		}),
		handlers.AllowedHeaders([]string{
			"Origin",
			"Content-Length",
			"Content-Type",
			"Authorization",
			"content-type-original",
			"x-requested-with",
		}),
		handlers.AllowedMethods([]string{"GET", "POST", "PUT", "DELETE", "PATCH", "OPTIONS"}),
		handlers.AllowCredentials(),
	)
}
