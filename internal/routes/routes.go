package routes

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/tunabsdrmz/boxing-gym-management/internal/apidocs"
	"github.com/tunabsdrmz/boxing-gym-management/internal/handler"
)

type Routes struct {
	Auth interface {
		Register(r chi.Router)
	}
	Fighter interface {
		Register(r chi.Router, auth func(http.Handler) http.Handler)
	}
	Trainer interface {
		Register(r chi.Router, auth func(http.Handler) http.Handler)
	}
	Admin interface {
		Register(r chi.Router, auth func(http.Handler) http.Handler)
	}
	Ops interface {
		Register(r chi.Router, auth func(http.Handler) http.Handler)
	}
}

func NewRoutes(h handler.Handler) Routes {
	return Routes{
		Auth:    &authRoutes{handler: h},
		Fighter: &fighterRoutes{handler: h},
		Trainer: &trainerRoutes{handler: h},
		Admin:   &adminRoutes{handler: h},
		Ops:     &opsRoutes{handler: h},
	}
}

func NewRouter(h handler.Handler, auth func(http.Handler) http.Handler) *chi.Mux {
	r := chi.NewRouter()
	r.Get("/openapi.yaml", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/yaml; charset=utf-8")
		_, _ = w.Write(apidocs.OpenAPIYAML)
	})
	all := NewRoutes(h)
	all.Auth.Register(r)
	all.Fighter.Register(r, auth)
	all.Trainer.Register(r, auth)
	all.Admin.Register(r, auth)
	all.Ops.Register(r, auth)
	return r
}
