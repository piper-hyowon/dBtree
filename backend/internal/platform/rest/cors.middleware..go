package rest

import (
	"net/http"
)

type CORSConfig struct {
	AllowedOrigins   []string
	AllowCredentials bool
}

func NewCORSMiddleware(config CORSConfig) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			origin := r.Header.Get("Origin")

			allowOrigin := ""
			for _, allowed := range config.AllowedOrigins {
				if allowed == "*" {
					allowOrigin = "*"
					break
				}
				if allowed == origin {
					allowOrigin = origin
					break
				}
			}

			// Set CORS headers if origin is allowed
			if allowOrigin != "" {
				w.Header().Set("Access-Control-Allow-Origin", allowOrigin)
				w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS, PATCH")
				w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, Accept, X-Requested-With")
				w.Header().Set("Access-Control-Expose-Headers", "Retry-After, Content-Length")

				// Only set credentials if not using wildcard
				if config.AllowCredentials && allowOrigin != "*" {
					w.Header().Set("Access-Control-Allow-Credentials", "true")
				}
			}

			// Handle preflight requests
			if r.Method == "OPTIONS" {
				// Always respond to OPTIONS with 200/204 if origin is allowed
				if allowOrigin != "" {
					w.WriteHeader(http.StatusNoContent) // 204 is more appropriate for OPTIONS
				} else {
					// Origin not allowed - still need to respond but without CORS headers
					w.WriteHeader(http.StatusForbidden)
				}
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
