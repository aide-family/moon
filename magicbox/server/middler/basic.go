package middler

import (
	"encoding/base64"
	nethttp "net/http"
	"strings"
)

// BasicAuthMiddleware creates a middleware that implements HTTP Basic Authentication.
func BasicAuthMiddleware(username, password string) func(nethttp.Handler) nethttp.Handler {
	return func(next nethttp.Handler) nethttp.Handler {
		return nethttp.HandlerFunc(func(w nethttp.ResponseWriter, r *nethttp.Request) {
			// Get the Authorization header
			auth := r.Header.Get("Authorization")
			if auth == "" {
				w.Header().Set("WWW-Authenticate", `Basic realm="Swagger Documentation"`)
				nethttp.Error(w, "Unauthorized", nethttp.StatusUnauthorized)
				return
			}

			// Check if it's Basic auth
			if !strings.HasPrefix(auth, "Basic ") {
				w.Header().Set("WWW-Authenticate", `Basic realm="Swagger Documentation"`)
				nethttp.Error(w, "Unauthorized", nethttp.StatusUnauthorized)
				return
			}

			// Decode the credentials
			encoded := strings.TrimPrefix(auth, "Basic ")
			decoded, err := base64.StdEncoding.DecodeString(encoded)
			if err != nil {
				w.Header().Set("WWW-Authenticate", `Basic realm="Swagger Documentation"`)
				nethttp.Error(w, "Unauthorized", nethttp.StatusUnauthorized)
				return
			}

			// Split username:password
			creds := strings.SplitN(string(decoded), ":", 2)
			if len(creds) != 2 || creds[0] != username || creds[1] != password {
				w.Header().Set("WWW-Authenticate", `Basic realm="Swagger Documentation"`)
				nethttp.Error(w, "Unauthorized", nethttp.StatusUnauthorized)
				return
			}

			// Authentication successful, proceed to next handler
			next.ServeHTTP(w, r)
		})
	}
}
