package admin

import (
	"net/http"

	"github.com/golang-jwt/jwt/v5"

	"fullstack-cms/internal/http/middleware"
)

func currentUserID(r *http.Request) string {
	claims := middleware.UserClaims(r)
	if claims == nil {
		return "unknown"
	}
	if m, ok := claims.(jwt.MapClaims); ok {
		if v, ok2 := m["uid"].(string); ok2 {
			return v
		}
	}
	return "unknown"
}
