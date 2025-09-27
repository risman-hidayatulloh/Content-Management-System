package middleware

import (
	"context"
	"net/http"

	"github.com/golang-jwt/jwt/v5"
)

type ctxKey string

const userKey ctxKey = "claims"

func JWT(secret []byte) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			c, err := r.Cookie("access_token")
			if err != nil || c.Value == "" {
				http.Error(w, "unauthorized", http.StatusUnauthorized)
				return
			}
			tok, err := jwt.Parse(c.Value, func(t *jwt.Token) (interface{}, error) { return secret, nil })
			if err != nil || !tok.Valid {
				http.Error(w, "unauthorized", http.StatusUnauthorized)
				return
			}
			ctx := context.WithValue(r.Context(), userKey, tok.Claims)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func UserClaims(r *http.Request) jwt.Claims {
	if v := r.Context().Value(userKey); v != nil {
		return v.(jwt.Claims)
	}
	return nil
}
