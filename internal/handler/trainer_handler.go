package handler

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/tunabsdrmz/boxing-gym-management/internal/repository"
	"github.com/tunabsdrmz/boxing-gym-management/internal/utils"
)

type TrainerHandler struct {
	repository repository.Repository
}

func (h *TrainerHandler) CreateTrainer(w http.ResponseWriter, r *http.Request) {
	var req repository.CreateTrainerRequest
	err := utils.ReadJSON(w, r, &req)
	if err != nil {
		utils.BadRequestResponse(w, r, err)
		return
	}
	trainer, err := h.repository.Trainer.CreateTrainer(r.Context(), req)
	if err != nil {
		utils.InternalServerError(w, r, err)
		return
	}
	utils.JsonResponse(w, http.StatusCreated, trainer, "Trainer created successfully")
}

func (h *TrainerHandler) GetTrainerByID(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		id = r.URL.Query().Get("id")
	}
	if id == "" {
		utils.BadRequestResponse(w, r, errors.New("trainer id is required"))
		return
	}
	req := repository.GetTrainerRequest{ID: id}
	trainer, err := h.repository.Trainer.GetTrainerByID(r.Context(), req)
	if err != nil {
		utils.NotFoundResponse(w, r, err)
		return
	}
	utils.JsonResponse(w, http.StatusOK, trainer, "Trainer retrieved successfully")
}

func (h *TrainerHandler) GetAllTrainers(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	limit, offset, err := ResolveListPagination(q.Get("limit"), q.Get("offset"), q.Get("page"))
	if err != nil {
		utils.BadRequestResponse(w, r, err)
		return
	}
	req := repository.GetAllTrainersRequest{
		Limit:  strconv.Itoa(limit),
		Offset: strconv.Itoa(offset),
	}
	result, err := h.repository.Trainer.GetAllTrainers(r.Context(), req)
	if err != nil {
		utils.InternalServerError(w, r, err)
		return
	}
	data := trainerListData{
		Trainers: result.Trainers,
		Pagination: listPagination{
			Limit:      limit,
			TotalPages: paginationTotalPages(result.Total, limit),
			Page:       paginationCurrentPage(offset, limit),
			Offset:     offset,
			Total:      result.Total,
		},
	}
	utils.JsonResponse(w, http.StatusOK, data, "Trainers retrieved successfully")
}

func (h *TrainerHandler) UpdateTrainer(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	if id == "" {
		utils.BadRequestResponse(w, r, errors.New("trainer id is required"))
		return
	}
	req := repository.UpdateTrainerRequest{ID: id}
	err := utils.ReadJSON(w, r, &req)
	if err != nil {
		utils.BadRequestResponse(w, r, err)
		return
	}
	trainer, err := h.repository.Trainer.UpdateTrainer(r.Context(), req)
	if err != nil {
		utils.InternalServerError(w, r, err)
		return
	}
	utils.JsonResponse(w, http.StatusOK, trainer, "Trainer updated successfully")
}

func (h *TrainerHandler) DeleteTrainer(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	if id == "" {
		utils.BadRequestResponse(w, r, errors.New("trainer id is required"))
		return
	}
	req := repository.DeleteTrainerRequest{ID: id}
	err := h.repository.Trainer.DeleteTrainer(r.Context(), req)
	if err != nil {
		utils.NotFoundResponse(w, r, err)
		return
	}
	utils.JsonResponse(w, http.StatusOK, nil, "Trainer deleted successfully")
}
