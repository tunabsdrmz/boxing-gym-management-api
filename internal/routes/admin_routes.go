package routes

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/tunabsdrmz/boxing-gym-management/internal/auth"
	"github.com/tunabsdrmz/boxing-gym-management/internal/handler"
	"github.com/tunabsdrmz/boxing-gym-management/internal/middleware"
)

type adminRoutes struct {
	handler handler.Handler
}

func (a *adminRoutes) Register(r chi.Router, authenticate func(http.Handler) http.Handler) {
	adminOnly := middleware.RequireRoles(auth.RoleAdmin)

	r.Route("/admin", func(r chi.Router) {
		r.Use(authenticate)
		r.With(adminOnly).Get("/users", a.handler.Admin.ListUsers)
		r.With(adminOnly).Patch("/users/{id}", a.handler.Admin.PatchUser)
	})
}
