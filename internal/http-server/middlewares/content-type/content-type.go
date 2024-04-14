package content_type

import (
	"net/http"
)

type CustomMiddleware func(next http.Handler) http.Handler

func SetContentTypeApplicationJson(next http.Handler) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		const op = "middlewares.auth.SetContentTypeApplicationJson"
		rw.Header().Set("Content-Type", "application/json")
	})
}
