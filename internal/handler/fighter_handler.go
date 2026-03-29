package handler

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/tunabsdrmz/boxing-gym-management/internal/repository"
	"github.com/tunabsdrmz/boxing-gym-management/internal/utils"
)

type FighterHandler struct {
	repository repository.Repository
}

func (h *FighterHandler) CreateFighter(w http.ResponseWriter, r *http.Request) {
	var req repository.CreateFighterRequest
	err := utils.ReadJSON(w, r, &req)
	if err != nil {
		utils.BadRequestResponse(w, r, err)
		return
	}
	fighter, err := h.repository.Fighter.CreateFighter(r.Context(), req)
	if err != nil {
		utils.WriteJSONError(w, http.StatusInternalServerError, err.Error())
		return
	}
	utils.JsonResponse(w, http.StatusCreated, fighter, "Fighter created successfully")
}

func (h *FighterHandler) GetFighterByID(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		id = r.URL.Query().Get("id")
	}
	if id == "" {
		utils.BadRequestResponse(w, r, errors.New("fighter id is required"))
		return
	}
	req := repository.GetFighterRequest{ID: id}
	fighter, err := h.repository.Fighter.GetFighterByID(r.Context(), req)
	if err != nil {
		utils.NotFoundResponse(w, r, err)
		return
	}
	utils.JsonResponse(w, http.StatusOK, fighter, "Fighter retrieved successfully")
}

func (h *FighterHandler) GetAllFighters(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	limit, offset, err := ResolveListPagination(q.Get("limit"), q.Get("offset"), q.Get("page"))
	if err != nil {
		utils.BadRequestResponse(w, r, err)
		return
	}
	req := repository.GetAllFightersRequest{
		Limit:  strconv.Itoa(limit),
		Offset: strconv.Itoa(offset),
	}
	result, err := h.repository.Fighter.GetAllFighters(r.Context(), req)
	if err != nil {
		utils.InternalServerError(w, r, err)
		return
	}
	data := fighterListData{
		Fighters: result.Fighters,
		Pagination: listPagination{
			Limit:      limit,
			TotalPages: paginationTotalPages(result.Total, limit),
			Page:       paginationCurrentPage(offset, limit),
			Offset:     offset,
			Total:      result.Total,
		},
	}
	utils.JsonResponse(w, http.StatusOK, data, "Fighters retrieved successfully")
}

func (h *FighterHandler) UpdateFighter(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	if id == "" {
		utils.BadRequestResponse(w, r, errors.New("fighter id is required"))
		return
	}
	req := repository.UpdateFighterRequest{ID: id}
	err := utils.ReadJSON(w, r, &req)
	if err != nil {
		utils.BadRequestResponse(w, r, err)
		return
	}
	fighter, err := h.repository.Fighter.UpdateFighter(r.Context(), req)
	if err != nil {
		utils.InternalServerError(w, r, err)
		return
	}
	utils.JsonResponse(w, http.StatusOK, fighter, "Fighter updated successfully")
}

func (h *FighterHandler) DeleteFighter(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	if id == "" {
		utils.BadRequestResponse(w, r, errors.New("fighter id is required"))
		return
	}
	req := repository.DeleteFighterRequest{ID: id}
	err := h.repository.Fighter.DeleteFighter(r.Context(), req)
	if err != nil {
		utils.NotFoundResponse(w, r, err)
		return
	}
	utils.JsonResponse(w, http.StatusOK, nil, "Fighter deleted successfully")
}
