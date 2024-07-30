package middleware

import (
	"net/http"

	"github.com/gorilla/handlers"
)

// Cors 跨域中间件
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
			"accept",
			"origin",
			"user-agent",
			"*",
			"User-Agent",
			"Referer",
			"Accept-Encoding",
			"Accept-Language",
			"X-Requested-With",
			"X-Requested-With",
			"Source-Type",
		}),
		handlers.AllowedMethods([]string{"GET", "POST", "PUT", "DELETE", "PATCH", "OPTIONS"}),
		handlers.AllowCredentials(),
	)
}
