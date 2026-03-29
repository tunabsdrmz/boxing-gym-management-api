package routes

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/tunabsdrmz/boxing-gym-management/internal/auth"
	"github.com/tunabsdrmz/boxing-gym-management/internal/handler"
	"github.com/tunabsdrmz/boxing-gym-management/internal/middleware"
)

type trainerRoutes struct {
	handler handler.Handler
}

func (t *trainerRoutes) Register(r chi.Router, authenticate func(http.Handler) http.Handler) {
	read := middleware.RequireRoles(auth.RoleViewer, auth.RoleStaff, auth.RoleAdmin)
	write := middleware.RequireRoles(auth.RoleStaff, auth.RoleAdmin)

	r.Route("/trainers", func(r chi.Router) {
		r.Group(func(r chi.Router) {
			r.Use(authenticate)
			r.With(read).Get("/all", t.handler.Trainer.GetAllTrainers)
			r.With(read).Get("/{id}", t.handler.Trainer.GetTrainerByID)
		})
		r.Group(func(r chi.Router) {
			r.Use(authenticate)
			r.With(write).Post("/create", t.handler.Trainer.CreateTrainer)
			r.With(write).Put("/profile/update", t.handler.Trainer.UpdateTrainer)
			r.With(write).Delete("/profile/delete", t.handler.Trainer.DeleteTrainer)
		})
	})
}
