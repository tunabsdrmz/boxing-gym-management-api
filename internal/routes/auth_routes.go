package routes

import (
	"github.com/go-chi/chi/v5"
	"github.com/tunabsdrmz/boxing-gym-management/internal/handler"
)

type authRoutes struct {
	handler handler.Handler
}

func (a *authRoutes) Register(r chi.Router) {
	r.Route("/auth", func(r chi.Router) {
		r.Post("/register", a.handler.Auth.Register)
		r.Post("/login", a.handler.Auth.Login)
	})
}
