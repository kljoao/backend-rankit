package middlewares

import (
	"context"
	"net/http"
	"strings"

	"rankit/internal/ports"
)

type contextKey string

const UserIDKey contextKey = "userID"

// AuthMiddleware cria um middleware para validação de JWT.
func AuthMiddleware(tokenService ports.TokenService) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				http.Error(w, "Autenticação requerida", http.StatusUnauthorized)
				return
			}

			parts := strings.Split(authHeader, " ")
			if len(parts) != 2 || parts[0] != "Bearer" {
				http.Error(w, "Formato de token inválido (esperado: Bearer <token>)", http.StatusUnauthorized)
				return
			}

			tokenString := parts[1]
			userID, err := tokenService.ValidateToken(tokenString)
			if err != nil {
				http.Error(w, "Token inválido ou expirado: "+err.Error(), http.StatusUnauthorized)
				return
			}

			// Injeta ID no contexto
			ctx := context.WithValue(r.Context(), UserIDKey, userID)
			next.ServeHTTP(w, r.WithContext(ctx))
		}
		return http.HandlerFunc(fn)
	}
}
