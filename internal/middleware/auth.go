package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/betpro/server/internal/services"
)

type contextKey string

const UserContextKey contextKey = "user"

func Auth(authService *services.AuthService, profileCache services.AuthProfileCache) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if !strings.HasPrefix(authHeader, "Bearer ") {
				respondJSON(w, http.StatusUnauthorized, map[string]string{"error": "No token provided"})
				return
			}

			token := strings.TrimPrefix(authHeader, "Bearer ")
			claims, err := authService.VerifyToken(token)
			if err != nil {
				respondJSON(w, http.StatusUnauthorized, map[string]string{"error": "Invalid token"})
				return
			}

			profile, err := profileCache.Get(r.Context(), claims.UserID)
			if err != nil || profile == nil || !profile.IsActive {
				respondJSON(w, http.StatusUnauthorized, map[string]string{"error": "Account disabled"})
				return
			}

			claims.Role = profile.Role

			ctx := context.WithValue(r.Context(), UserContextKey, claims)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func RequireRole(roles ...string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			claims, ok := r.Context().Value(UserContextKey).(*services.Claims)
			if !ok {
				respondJSON(w, http.StatusUnauthorized, map[string]string{"error": "Unauthorized"})
				return
			}

			hasRole := false
			for _, role := range roles {
				if claims.Role == role {
					hasRole = true
					break
				}
			}

			if !hasRole {
				respondJSON(w, http.StatusForbidden, map[string]string{"error": "Forbidden"})
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

func GetUserFromContext(ctx context.Context) (*services.Claims, bool) {
	claims, ok := ctx.Value(UserContextKey).(*services.Claims)
	return claims, ok
}
