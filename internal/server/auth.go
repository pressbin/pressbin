package server

import (
	"net/http"
	"strings"

	"pressbin.dev/pressbin/internal/store"
)

func (s *Server) requireAuth(permission string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			raw := strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer ")
			raw = strings.TrimSpace(raw)
			if raw == "" {
				writeError(w, 401, "missing api key")
				return
			}

			key, err := s.store.FindKeyByRaw(raw)
			if err != nil || !key.HasPermission(permission) {
				writeError(w, 403, "forbidden")
				return
			}

			go func(k store.APIKey) {
				_ = s.store.TouchKey(k.ID)
			}(key)

			next.ServeHTTP(w, r)
		})
	}
}
