package routes

import (
	"net/http"

	"github.com/go-chi/chi/v5"
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
}

func NewRoutes(h handler.Handler) Routes {
	return Routes{
		Auth:    &authRoutes{handler: h},
		Fighter: &fighterRoutes{handler: h},
		Trainer: &trainerRoutes{handler: h},
	}
}

func NewRouter(h handler.Handler, auth func(http.Handler) http.Handler) *chi.Mux {
	r := chi.NewRouter()
	routes := NewRoutes(h)
	routes.Auth.Register(r)
	routes.Fighter.Register(r, auth)
	routes.Trainer.Register(r, auth)
	return r
}
