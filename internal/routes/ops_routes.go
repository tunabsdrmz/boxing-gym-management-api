package routes

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/tunabsdrmz/boxing-gym-management/internal/auth"
	"github.com/tunabsdrmz/boxing-gym-management/internal/handler"
	"github.com/tunabsdrmz/boxing-gym-management/internal/middleware"
)

type opsRoutes struct {
	handler handler.Handler
}

func (o *opsRoutes) Register(r chi.Router, authenticate func(http.Handler) http.Handler) {
	read := middleware.RequireRoles(auth.RoleViewer, auth.RoleStaff, auth.RoleAdmin)
	write := middleware.RequireRoles(auth.RoleStaff, auth.RoleAdmin)

	r.Route("/schedule", func(r chi.Router) {
		r.Use(authenticate)
		r.With(read).Get("/events", o.handler.Ops.ListScheduleEvents)
		r.With(read).Get("/events/{id}", o.handler.Ops.GetScheduleEvent)
		r.With(write).Post("/events", o.handler.Ops.CreateScheduleEvent)
		r.With(write).Put("/events/{id}", o.handler.Ops.UpdateScheduleEvent)
		r.With(write).Delete("/events/{id}", o.handler.Ops.DeleteScheduleEvent)
	})

	r.Route("/attendance", func(r chi.Router) {
		r.Use(authenticate)
		r.With(read).Get("/", o.handler.Ops.ListAttendanceByDate)
		r.With(write).Post("/", o.handler.Ops.UpsertAttendance)
		r.With(write).Delete("/{id}", o.handler.Ops.DeleteAttendance)
	})

	r.Route("/announcements", func(r chi.Router) {
		r.Use(authenticate)
		r.With(read).Get("/active", o.handler.Ops.ListAnnouncementsActive)
		r.With(write).Get("/all", o.handler.Ops.ListAnnouncementsAll)
		r.With(write).Post("/", o.handler.Ops.CreateAnnouncement)
		r.With(write).Put("/{id}", o.handler.Ops.UpdateAnnouncement)
		r.With(write).Delete("/{id}", o.handler.Ops.DeleteAnnouncement)
	})
}
