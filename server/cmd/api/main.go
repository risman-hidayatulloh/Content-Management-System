package main

import (
	"crypto/sha256"
	"fullstack-cms/config"
	"fullstack-cms/internal/db"
	"fullstack-cms/internal/http/admin"
	"fullstack-cms/internal/http/middleware"
	"fullstack-cms/internal/http/public"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/gorilla/csrf"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/zap"
)

func main() {
	cfg := config.Load()

	// Initialize zap logger
	logger, err := zap.NewProduction()
	if err != nil {
		panic(err)
	}
	defer logger.Sync()

	gormdb := db.Connect(cfg, logger)

	r := mux.NewRouter()
	r.Use(middleware.Recover(logger))
	r.Use(middleware.Metrics()) // prometheus per-request

	// health & metrics
	r.HandleFunc("/healthz", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(200)
		w.Write([]byte(`{"ok":true}`))
	}).Methods(http.MethodGet)

	r.HandleFunc("/csrf", func(w http.ResponseWriter, r *http.Request) {
		token := csrf.Token(r)
		w.Header().Set("X-CSRF-Token", token)
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"csrf_token":"` + token + `"}`))
	}).Methods(http.MethodGet)

	r.Handle("/metrics", promhttp.Handler())

	// static media
	fs := http.FileServer(http.Dir(cfg.App.MediaDir))
	r.PathPrefix("/media/").Handler(http.StripPrefix("/media/", fs))

	// public API
	public.Register(r, gormdb, logger, cfg)

	// admin API
	ar := r.PathPrefix("/admin").Subrouter()
	admin.Register(ar, gormdb, logger, cfg)

	// CORS (ketat ke origin admin UI)
	cors := handlers.CORS(
		handlers.AllowedOrigins([]string{cfg.App.AdminOrigin}),
		handlers.AllowedMethods([]string{"GET", "POST", "PATCH", "DELETE", "OPTIONS"}),
		handlers.AllowedHeaders([]string{"Content-Type", "Authorization", "X-CSRF-Token"}),
		handlers.AllowCredentials(),
	)

	// access log
	logged := handlers.CombinedLoggingHandler(os.Stdout, cors(r))

	// CSRF: lindungi semua state-changing (GET/HEAD/OPTIONS tetap lolos)
	key := []byte(cfg.App.CSRFKey)
	if l := len(key); l != 32 && l != 64 {
		sum := sha256.Sum256(key)
		key = sum[:]
	}

	csrfMW := csrf.Protect(
		key,
		csrf.Secure(cfg.App.Prod),
		csrf.SameSite(csrf.SameSiteLaxMode),
		csrf.CookieName("csrf_token"),
		csrf.RequestHeader("X-CSRF-Token"),
	)

	srv := &http.Server{
		Addr:         ":" + strconv.Itoa(cfg.App.Port),
		Handler:      csrfMW(logged),
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	logger.Info("server.start", zap.String("addr", srv.Addr))
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		logger.Fatal("server.fail", zap.Error(err))
	}
}
