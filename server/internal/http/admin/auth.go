package admin

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/gorilla/csrf"
	"github.com/gorilla/mux"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

	"fullstack-cms/config"
	"fullstack-cms/internal/content"
	"fullstack-cms/internal/http/middleware"
)

func Register(r *mux.Router, db *gorm.DB, _ any, cfg config.Config) {
	// login
	r.HandleFunc("/auth/login", login(db, cfg)).Methods(http.MethodPost)

	// protected
	p := r.NewRoute().Subrouter()
	p.Use(middleware.JWT([]byte(cfg.App.JWTSecret)))

	// Content Types CRUD
	registerContentTypes(p, db)

	// Entries CRUD + publish/schedule + rollback
	registerEntries(p, db)

	// media upload
	p.HandleFunc("/media", uploadMedia(db, cfg)).Methods(http.MethodPost)
}

func login(db *gorm.DB, cfg config.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req struct{ Email, Password string }
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "bad json", http.StatusBadRequest)
			return
		}
		var u content.User
		if err := db.Preload("Roles.Permissions").Where("email = ?", req.Email).First(&u).Error; err != nil {
			http.Error(w, "invalid", http.StatusUnauthorized)
			return
		}
		if bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(req.Password)) != nil {
			http.Error(w, "invalid", http.StatusUnauthorized)
			return
		}

		claims := jwt.MapClaims{"uid": u.ID, "exp": time.Now().Add(8 * time.Hour).Unix()}
		t := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
		tokenStr, _ := t.SignedString([]byte(cfg.App.JWTSecret))

		http.SetCookie(w, &http.Cookie{
			Name:     "access_token",
			Value:    tokenStr,
			Path:     "/",
			HttpOnly: true,
			Secure:   cfg.App.Prod,
			SameSite: http.SameSiteLaxMode,
			MaxAge:   int((8 * time.Hour).Seconds()),
		})
		// expose CSRF token untuk SPA
		w.Header().Set("X-CSRF-Token", csrf.Token(r))
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"ok":true}`))
	}
}
