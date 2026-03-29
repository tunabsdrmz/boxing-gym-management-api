package handler

import (
	"net/http"

	"github.com/tunabsdrmz/boxing-gym-management/internal/repository"
)


type Handler struct {
	Auth interface {
		Register(w http.ResponseWriter, r *http.Request)
		Login(w http.ResponseWriter, r *http.Request)
	}
	Fighter interface{
		CreateFighter(w http.ResponseWriter, r *http.Request)
		GetFighterByID(w http.ResponseWriter, r *http.Request)
		GetAllFighters(w http.ResponseWriter, r *http.Request)
		UpdateFighter(w http.ResponseWriter, r *http.Request)
		DeleteFighter(w http.ResponseWriter, r *http.Request)
	}
	Trainer interface{
		CreateTrainer(w http.ResponseWriter, r *http.Request)
			GetTrainerByID(w http.ResponseWriter, r *http.Request)
		GetAllTrainers(w http.ResponseWriter, r *http.Request)
		UpdateTrainer(w http.ResponseWriter, r *http.Request)
		DeleteTrainer(w http.ResponseWriter, r *http.Request)
	}
}

func NewHandler(repository repository.Repository) Handler {
	return Handler{
		Auth:    &AuthHandler{repository: repository},
		Fighter: &FighterHandler{repository: repository},
		Trainer: &TrainerHandler{repository: repository},
	}
}