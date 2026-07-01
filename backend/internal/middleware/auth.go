package middleware

import (
	"context"
	"net/http"

	"github.com/Andrii-K-17/light-chat/internal/response"
	"github.com/golang-jwt/jwt/v5"
)

// contextKey defines a custom type for context keys to avoid collisions.
type contextKey string

// UserIDKey is the context key used to store and retrieve the user ID.
const UserIDKey contextKey = "user_id"

// Auth returns a middleware that validates a JWT from a cookie.
func Auth(secret string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
			cookie, err := req.Cookie("token")
			if err != nil {
				response.Error(res, http.StatusUnauthorized, "unauthorized")
				return
			}

			parsedToken, err := jwt.Parse(cookie.Value, func(token *jwt.Token) (any, error) {
				if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
					return nil, jwt.ErrSignatureInvalid
				}
				return []byte(secret), nil
			})

			if err != nil || !parsedToken.Valid {
				response.Error(res, http.StatusUnauthorized, "unauthorized")
				return
			}

			claims, ok := parsedToken.Claims.(jwt.MapClaims)
			if !ok {
				response.Error(res, http.StatusUnauthorized, "unauthorized")
				return
			}

			userID := int(claims["user_id"].(float64))
			ctx := context.WithValue(req.Context(), UserIDKey, userID)
			next.ServeHTTP(res, req.WithContext(ctx))
		})
	}
}

// GetUserID extracts the user ID from the context.
func GetUserID(ctx context.Context) int {
	userID, ok := ctx.Value(UserIDKey).(int)
	if !ok {
		return 0
	}
	return userID
}
