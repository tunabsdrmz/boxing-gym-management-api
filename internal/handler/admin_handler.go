package handler

import (
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/tunabsdrmz/boxing-gym-management/internal/repository"
	"github.com/tunabsdrmz/boxing-gym-management/internal/utils"
)

type AdminHandler struct {
	repository repository.Repository
}

type adminUserListData struct {
	Users      []repository.User `json:"users"`
	Pagination listPagination    `json:"pagination"`
}

type adminPatchUserRequest struct {
	Role         *string `json:"role"`
	Locked       *bool   `json:"locked"`
	LockedReason *string `json:"locked_reason"`
}

func (h *AdminHandler) ListUsers(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	limit, offset, err := ResolveListPagination(q.Get("limit"), q.Get("offset"), q.Get("page"))
	if err != nil {
		utils.BadRequestResponse(w, r, err)
		return
	}
	res, err := h.repository.User.ListUsers(r.Context(), limit, offset)
	if err != nil {
		utils.InternalServerError(w, r, err)
		return
	}
	data := adminUserListData{
		Users: res.Users,
		Pagination: listPagination{
			Limit:      limit,
			TotalPages: paginationTotalPages(res.Total, limit),
			Page:       paginationCurrentPage(offset, limit),
			Offset:     offset,
			Total:      res.Total,
		},
	}
	utils.JsonResponse(w, http.StatusOK, data, "users retrieved")
}

func (h *AdminHandler) PatchUser(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		utils.BadRequestResponse(w, r, errors.New("user id required"))
		return
	}
	var body adminPatchUserRequest
	if err := utils.ReadJSON(w, r, &body); err != nil {
		utils.BadRequestResponse(w, r, err)
		return
	}
	if body.Role == nil && body.Locked == nil && body.LockedReason == nil {
		utils.BadRequestResponse(w, r, errors.New("no fields to update"))
		return
	}
	u, err := h.repository.User.UpdateUserAdmin(r.Context(), repository.UpdateUserAdminRequest{
		ID:           id,
		Role:         body.Role,
		Locked:       body.Locked,
		LockedReason: body.LockedReason,
	})
	if err != nil {
		if errors.Is(err, repository.ErrUserNotFound) {
			utils.NotFoundResponse(w, r, err)
			return
		}
		if errors.Is(err, repository.ErrInvalidRole) {
			utils.BadRequestResponse(w, r, err)
			return
		}
		utils.InternalServerError(w, r, err)
		return
	}
	utils.JsonResponse(w, http.StatusOK, u, "user updated")
}
