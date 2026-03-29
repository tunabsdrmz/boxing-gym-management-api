package handler

import (
	"database/sql"
	"errors"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/tunabsdrmz/boxing-gym-management/internal/auth"
	"github.com/tunabsdrmz/boxing-gym-management/internal/repository"
	"github.com/tunabsdrmz/boxing-gym-management/internal/utils"
)

type OpsHandler struct {
	repository repository.Repository
}

func userIDOr401(w http.ResponseWriter, r *http.Request) (string, bool) {
	uid, ok := auth.UserIDFromContext(r.Context())
	if !ok {
		utils.UnauthorizedErrorResponse(w, r, errors.New("missing user"))
		return "", false
	}
	return uid, true
}

type scheduleCreateBody struct {
	Title        string  `json:"title"`
	StartAt      string  `json:"start_at"`
	EndAt        string  `json:"end_at"`
	ResourceType string  `json:"resource_type"`
	TrainerID    *string `json:"trainer_id"`
	Notes        *string `json:"notes"`
}

func (h *OpsHandler) CreateScheduleEvent(w http.ResponseWriter, r *http.Request) {
	uid, ok := userIDOr401(w, r)
	if !ok {
		return
	}
	var body scheduleCreateBody
	if err := utils.ReadJSON(w, r, &body); err != nil {
		utils.BadRequestResponse(w, r, err)
		return
	}
	start, err := time.Parse(time.RFC3339, body.StartAt)
	if err != nil {
		utils.BadRequestResponse(w, r, errors.New("start_at must be RFC3339"))
		return
	}
	end, err := time.Parse(time.RFC3339, body.EndAt)
	if err != nil {
		utils.BadRequestResponse(w, r, errors.New("end_at must be RFC3339"))
		return
	}
	if body.Title == "" {
		utils.BadRequestResponse(w, r, errors.New("title required"))
		return
	}
	rt := body.ResourceType
	if rt == "" {
		rt = "general"
	}
	overlap, err := h.repository.Schedule.HasOverlap(r.Context(), nil, rt, start, end)
	if err != nil {
		utils.BadRequestResponse(w, r, err)
		return
	}
	if overlap {
		utils.ConflictResponse(w, r, errors.New("schedule overlaps an existing booking for this resource"))
		return
	}
	ev, err := h.repository.Schedule.Create(r.Context(), repository.CreateScheduleEventRequest{
		Title:        body.Title,
		StartAt:      start,
		EndAt:        end,
		ResourceType: rt,
		TrainerID:    body.TrainerID,
		Notes:        body.Notes,
		CreatedBy:    uid,
	})
	if err != nil {
		utils.InternalServerError(w, r, err)
		return
	}
	utils.JsonResponse(w, http.StatusCreated, ev, "event created")
}

func (h *OpsHandler) ListScheduleEvents(w http.ResponseWriter, r *http.Request) {
	var from, to *time.Time
	if fs := r.URL.Query().Get("from"); fs != "" {
		t, err := time.Parse(time.RFC3339, fs)
		if err != nil {
			utils.BadRequestResponse(w, r, errors.New("from must be RFC3339"))
			return
		}
		from = &t
	}
	if ts := r.URL.Query().Get("to"); ts != "" {
		t, err := time.Parse(time.RFC3339, ts)
		if err != nil {
			utils.BadRequestResponse(w, r, errors.New("to must be RFC3339"))
			return
		}
		to = &t
	}
	list, err := h.repository.Schedule.List(r.Context(), repository.ListScheduleRequest{From: from, To: to})
	if err != nil {
		utils.InternalServerError(w, r, err)
		return
	}
	utils.JsonResponse(w, http.StatusOK, list, "schedule retrieved")
}

func (h *OpsHandler) GetScheduleEvent(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	ev, err := h.repository.Schedule.GetByID(r.Context(), id)
	if errors.Is(err, sql.ErrNoRows) {
		utils.NotFoundResponse(w, r, err)
		return
	}
	if err != nil {
		utils.InternalServerError(w, r, err)
		return
	}
	utils.JsonResponse(w, http.StatusOK, ev, "event retrieved")
}

type scheduleUpdateBody struct {
	Title        *string `json:"title"`
	StartAt      *string `json:"start_at"`
	EndAt        *string `json:"end_at"`
	ResourceType *string `json:"resource_type"`
	TrainerID    *string `json:"trainer_id"`
	Notes        *string `json:"notes"`
}

func (h *OpsHandler) UpdateScheduleEvent(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	cur, err := h.repository.Schedule.GetByID(r.Context(), id)
	if errors.Is(err, sql.ErrNoRows) {
		utils.NotFoundResponse(w, r, err)
		return
	}
	if err != nil {
		utils.InternalServerError(w, r, err)
		return
	}
	var body scheduleUpdateBody
	if err := utils.ReadJSON(w, r, &body); err != nil {
		utils.BadRequestResponse(w, r, err)
		return
	}
	start := cur.StartAt
	end := cur.EndAt
	resType := cur.ResourceType
	if body.StartAt != nil {
		t, err := time.Parse(time.RFC3339, *body.StartAt)
		if err != nil {
			utils.BadRequestResponse(w, r, errors.New("start_at must be RFC3339"))
			return
		}
		start = t
	}
	if body.EndAt != nil {
		t, err := time.Parse(time.RFC3339, *body.EndAt)
		if err != nil {
			utils.BadRequestResponse(w, r, errors.New("end_at must be RFC3339"))
			return
		}
		end = t
	}
	if body.ResourceType != nil {
		resType = *body.ResourceType
	}
	ex := id
	overlap, err := h.repository.Schedule.HasOverlap(r.Context(), &ex, resType, start, end)
	if err != nil {
		utils.BadRequestResponse(w, r, err)
		return
	}
	if overlap {
		utils.ConflictResponse(w, r, errors.New("schedule overlaps an existing booking for this resource"))
		return
	}
	var startPtr *time.Time
	if body.StartAt != nil {
		startPtr = &start
	}
	var endPtr *time.Time
	if body.EndAt != nil {
		endPtr = &end
	}
	ev, err := h.repository.Schedule.Update(r.Context(), repository.UpdateScheduleEventRequest{
		ID:           id,
		Title:        body.Title,
		StartAt:      startPtr,
		EndAt:        endPtr,
		ResourceType: body.ResourceType,
		TrainerID:    body.TrainerID,
		Notes:        body.Notes,
	})
	if err != nil {
		utils.InternalServerError(w, r, err)
		return
	}
	utils.JsonResponse(w, http.StatusOK, ev, "event updated")
}

func (h *OpsHandler) DeleteScheduleEvent(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if err := h.repository.Schedule.Delete(r.Context(), id); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			utils.NotFoundResponse(w, r, err)
			return
		}
		utils.InternalServerError(w, r, err)
		return
	}
	utils.JsonResponse(w, http.StatusOK, nil, "event deleted")
}

// --- Attendance ---

type attendanceUpsertBody struct {
	GymDate   string  `json:"gym_date"`
	FighterID string  `json:"fighter_id"`
	Present   bool    `json:"present"`
	Notes     *string `json:"notes"`
}

func (h *OpsHandler) UpsertAttendance(w http.ResponseWriter, r *http.Request) {
	uid, ok := userIDOr401(w, r)
	if !ok {
		return
	}
	var body attendanceUpsertBody
	if err := utils.ReadJSON(w, r, &body); err != nil {
		utils.BadRequestResponse(w, r, err)
		return
	}
	if body.GymDate == "" || body.FighterID == "" {
		utils.BadRequestResponse(w, r, errors.New("gym_date and fighter_id required"))
		return
	}
	if _, err := time.Parse("2006-01-02", body.GymDate); err != nil {
		utils.BadRequestResponse(w, r, errors.New("gym_date must be YYYY-MM-DD"))
		return
	}
	rec, err := h.repository.Attendance.Upsert(r.Context(), repository.UpsertAttendanceRequest{
		GymDate:    body.GymDate,
		FighterID:  body.FighterID,
		Present:    body.Present,
		Notes:      body.Notes,
		RecordedBy: uid,
	})
	if err != nil {
		utils.InternalServerError(w, r, err)
		return
	}
	utils.JsonResponse(w, http.StatusOK, rec, "attendance saved")
}

func (h *OpsHandler) ListAttendanceByDate(w http.ResponseWriter, r *http.Request) {
	d := r.URL.Query().Get("gym_date")
	if d == "" {
		d = time.Now().UTC().Format("2006-01-02")
	}
	if _, err := time.Parse("2006-01-02", d); err != nil {
		utils.BadRequestResponse(w, r, errors.New("gym_date must be YYYY-MM-DD"))
		return
	}
	list, err := h.repository.Attendance.ListByDate(r.Context(), d)
	if err != nil {
		utils.InternalServerError(w, r, err)
		return
	}
	utils.JsonResponse(w, http.StatusOK, list, "attendance retrieved")
}

func (h *OpsHandler) DeleteAttendance(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if err := h.repository.Attendance.Delete(r.Context(), id); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			utils.NotFoundResponse(w, r, err)
			return
		}
		utils.InternalServerError(w, r, err)
		return
	}
	utils.JsonResponse(w, http.StatusOK, nil, "attendance deleted")
}

// --- Announcements ---

type announcementCreateBody struct {
	Title     string  `json:"title"`
	Body      string  `json:"body"`
	ExpiresAt *string `json:"expires_at"`
}

func (h *OpsHandler) CreateAnnouncement(w http.ResponseWriter, r *http.Request) {
	uid, ok := userIDOr401(w, r)
	if !ok {
		return
	}
	var body announcementCreateBody
	if err := utils.ReadJSON(w, r, &body); err != nil {
		utils.BadRequestResponse(w, r, err)
		return
	}
	if body.Title == "" || body.Body == "" {
		utils.BadRequestResponse(w, r, errors.New("title and body required"))
		return
	}
	var exp *time.Time
	if body.ExpiresAt != nil && *body.ExpiresAt != "" {
		t, err := time.Parse(time.RFC3339, *body.ExpiresAt)
		if err != nil {
			utils.BadRequestResponse(w, r, errors.New("expires_at must be RFC3339"))
			return
		}
		exp = &t
	}
	a, err := h.repository.Announcement.Create(r.Context(), repository.CreateAnnouncementRequest{
		Title:     body.Title,
		Body:      body.Body,
		ExpiresAt: exp,
		CreatedBy: uid,
	})
	if err != nil {
		utils.InternalServerError(w, r, err)
		return
	}
	utils.JsonResponse(w, http.StatusCreated, a, "announcement created")
}

func (h *OpsHandler) ListAnnouncementsActive(w http.ResponseWriter, r *http.Request) {
	list, err := h.repository.Announcement.ListActive(r.Context())
	if err != nil {
		utils.InternalServerError(w, r, err)
		return
	}
	utils.JsonResponse(w, http.StatusOK, list, "announcements retrieved")
}

func (h *OpsHandler) ListAnnouncementsAll(w http.ResponseWriter, r *http.Request) {
	list, err := h.repository.Announcement.ListAll(r.Context())
	if err != nil {
		utils.InternalServerError(w, r, err)
		return
	}
	utils.JsonResponse(w, http.StatusOK, list, "announcements retrieved")
}

type announcementUpdateBody struct {
	Title     *string `json:"title"`
	Body      *string `json:"body"`
	ExpiresAt *string `json:"expires_at"`
}

func (h *OpsHandler) UpdateAnnouncement(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var body announcementUpdateBody
	if err := utils.ReadJSON(w, r, &body); err != nil {
		utils.BadRequestResponse(w, r, err)
		return
	}
	var exp *time.Time
	clearExp := false
	if body.ExpiresAt != nil {
		if *body.ExpiresAt == "" {
			clearExp = true
		} else {
			t, err := time.Parse(time.RFC3339, *body.ExpiresAt)
			if err != nil {
				utils.BadRequestResponse(w, r, errors.New("expires_at must be RFC3339"))
				return
			}
			exp = &t
		}
	}
	a, err := h.repository.Announcement.Update(r.Context(), repository.UpdateAnnouncementRequest{
		ID:           id,
		Title:        body.Title,
		Body:         body.Body,
		ExpiresAt:    exp,
		ClearExpires: clearExp,
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			utils.NotFoundResponse(w, r, err)
			return
		}
		utils.InternalServerError(w, r, err)
		return
	}
	utils.JsonResponse(w, http.StatusOK, a, "announcement updated")
}

func (h *OpsHandler) DeleteAnnouncement(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if err := h.repository.Announcement.Delete(r.Context(), id); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			utils.NotFoundResponse(w, r, err)
			return
		}
		utils.InternalServerError(w, r, err)
		return
	}
	utils.JsonResponse(w, http.StatusOK, nil, "announcement deleted")
}
