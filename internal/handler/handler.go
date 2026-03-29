package handler

import (
	"net/http"

	"github.com/tunabsdrmz/boxing-gym-management/internal/repository"
)

type Handler struct {
	Auth interface {
		Register(w http.ResponseWriter, r *http.Request)
		Login(w http.ResponseWriter, r *http.Request)
		Refresh(w http.ResponseWriter, r *http.Request)
		ForgotPassword(w http.ResponseWriter, r *http.Request)
		ResetPassword(w http.ResponseWriter, r *http.Request)
	}
	Fighter interface {
		CreateFighter(w http.ResponseWriter, r *http.Request)
		GetFighterByID(w http.ResponseWriter, r *http.Request)
		GetAllFighters(w http.ResponseWriter, r *http.Request)
		ExportFightersCSV(w http.ResponseWriter, r *http.Request)
		AddAssistantTrainer(w http.ResponseWriter, r *http.Request)
		RemoveAssistantTrainer(w http.ResponseWriter, r *http.Request)
		UpdateFighter(w http.ResponseWriter, r *http.Request)
		DeleteFighter(w http.ResponseWriter, r *http.Request)
	}
	Trainer interface {
		CreateTrainer(w http.ResponseWriter, r *http.Request)
		GetTrainerByID(w http.ResponseWriter, r *http.Request)
		GetAllTrainers(w http.ResponseWriter, r *http.Request)
		UpdateTrainer(w http.ResponseWriter, r *http.Request)
		DeleteTrainer(w http.ResponseWriter, r *http.Request)
	}
	Admin interface {
		ListUsers(w http.ResponseWriter, r *http.Request)
		PatchUser(w http.ResponseWriter, r *http.Request)
	}
	Ops interface {
		CreateScheduleEvent(w http.ResponseWriter, r *http.Request)
		ListScheduleEvents(w http.ResponseWriter, r *http.Request)
		GetScheduleEvent(w http.ResponseWriter, r *http.Request)
		UpdateScheduleEvent(w http.ResponseWriter, r *http.Request)
		DeleteScheduleEvent(w http.ResponseWriter, r *http.Request)
		UpsertAttendance(w http.ResponseWriter, r *http.Request)
		ListAttendanceByDate(w http.ResponseWriter, r *http.Request)
		DeleteAttendance(w http.ResponseWriter, r *http.Request)
		CreateAnnouncement(w http.ResponseWriter, r *http.Request)
		ListAnnouncementsActive(w http.ResponseWriter, r *http.Request)
		ListAnnouncementsAll(w http.ResponseWriter, r *http.Request)
		UpdateAnnouncement(w http.ResponseWriter, r *http.Request)
		DeleteAnnouncement(w http.ResponseWriter, r *http.Request)
	}
}

func NewHandler(repo repository.Repository) Handler {
	return Handler{
		Auth:    &AuthHandler{repository: repo},
		Fighter: &FighterHandler{repository: repo},
		Trainer: &TrainerHandler{repository: repo},
		Admin:   &AdminHandler{repository: repo},
		Ops:     &OpsHandler{repository: repo},
	}
}
