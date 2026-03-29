package main

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/tunabsdrmz/boxing-gym-management/internal/config"
	"github.com/tunabsdrmz/boxing-gym-management/internal/handler"
	appmw "github.com/tunabsdrmz/boxing-gym-management/internal/middleware"
	"github.com/tunabsdrmz/boxing-gym-management/internal/repository"
	"github.com/tunabsdrmz/boxing-gym-management/internal/static"
)

type application struct {
	config     config.Config
	repository repository.Repository
	handler    handler.Handler
}

func (app *application) mount(apiRouter *chi.Mux) http.Handler {
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.RealIP)
	r.Use(middleware.Recoverer)
	r.Use(middleware.RequestID)

	if len(config.App.CORS.AllowedOrigins) > 0 {
		r.Use(cors.Handler(cors.Options{
			AllowedOrigins:   config.App.CORS.AllowedOrigins,
			AllowedMethods:   []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
			AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-Request-ID"},
			ExposedHeaders:   []string{"X-Request-ID"},
			AllowCredentials: false,
			MaxAge:           300,
		}))
	}
	if config.App.Security.RateLimitRPM > 0 {
		r.Use(appmw.RateLimitPerIP(config.App.Security.RateLimitRPM))
	}

	r.Use(middleware.Timeout(60 * time.Second))

	r.Get("/healthcheck", healthcheck)

	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = w.Write(static.IndexHTML)
	})

	r.Mount("/api/v1", apiRouter)

	return r
}

func (app *application) run(mux http.Handler) error {
	srv := &http.Server{
		Addr:         app.config.Port,
		Handler:      mux,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  10 * time.Second,
	}

	return srv.ListenAndServe()
}
