package routes

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/tunabsdrmz/boxing-gym-management/internal/auth"
	"github.com/tunabsdrmz/boxing-gym-management/internal/handler"
	"github.com/tunabsdrmz/boxing-gym-management/internal/middleware"
)

type fighterRoutes struct {
	handler handler.Handler
}

func (f *fighterRoutes) Register(r chi.Router, authenticate func(http.Handler) http.Handler) {
	read := middleware.RequireRoles(auth.RoleViewer, auth.RoleStaff, auth.RoleAdmin)
	write := middleware.RequireRoles(auth.RoleStaff, auth.RoleAdmin)

	r.Route("/fighters", func(r chi.Router) {
		r.Group(func(r chi.Router) {
			r.Use(authenticate)
			r.With(read).Get("/all", f.handler.Fighter.GetAllFighters)
			r.With(read).Get("/{id}", f.handler.Fighter.GetFighterByID)
		})
		r.Group(func(r chi.Router) {
			r.Use(authenticate)
			r.With(write).Post("/create", f.handler.Fighter.CreateFighter)
			r.With(write).Put("/profile/update", f.handler.Fighter.UpdateFighter)
			r.With(write).Delete("/profile/delete", f.handler.Fighter.DeleteFighter)
		})
	})
}
