package auth

import (
	"context"
	"net/http"
	"strings"

	"github.com/gabriel-ferreira26/test-go/internal/httpx"
)

type contextKey int

const userIDKey contextKey = iota

// RequireAuth wraps next so it only runs when the request carries a valid
// "Authorization: Bearer <token>" header. On success, the authenticated
// user's ID is stashed in the request context for handlers to read via
// UserIDFromContext.
func (s *Store) RequireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token, ok := bearerToken(r)
		if !ok {
			httpx.WriteError(w, http.StatusUnauthorized, "missing bearer token")
			return
		}

		userID, ok := s.UserIDForToken(token)
		if !ok {
			httpx.WriteError(w, http.StatusUnauthorized, "invalid or expired token")
			return
		}

		ctx := context.WithValue(r.Context(), userIDKey, userID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func bearerToken(r *http.Request) (string, bool) {
	const prefix = "Bearer "
	h := r.Header.Get("Authorization")
	if !strings.HasPrefix(h, prefix) {
		return "", false
	}
	return strings.TrimPrefix(h, prefix), true
}

// UserIDFromContext returns the authenticated user's ID set by RequireAuth.
func UserIDFromContext(ctx context.Context) (int, bool) {
	id, ok := ctx.Value(userIDKey).(int)
	return id, ok
}
