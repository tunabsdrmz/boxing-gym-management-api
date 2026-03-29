package handler

import (
	"context"
	"encoding/csv"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/tunabsdrmz/boxing-gym-management/internal/repository"
	"github.com/tunabsdrmz/boxing-gym-management/internal/utils"
)

type FighterHandler struct {
	repository repository.Repository
}

func (h *FighterHandler) attachAssistantsBatch(ctx context.Context, fighters []repository.Fighter) ([]repository.Fighter, error) {
	if len(fighters) == 0 {
		return fighters, nil
	}
	ids := make([]string, len(fighters))
	for i := range fighters {
		ids[i] = fighters[i].ID
	}
	m, err := h.repository.FighterAssistant.ListByFighterIDs(ctx, ids)
	if err != nil {
		return nil, err
	}
	for i := range fighters {
		if a, ok := m[fighters[i].ID]; ok {
			fighters[i].AssistantTrainers = a
		}
	}
	return fighters, nil
}

func (h *FighterHandler) CreateFighter(w http.ResponseWriter, r *http.Request) {
	var req repository.CreateFighterRequest
	if err := utils.ReadJSON(w, r, &req); err != nil {
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
	fighter, err := h.repository.Fighter.GetFighterByID(r.Context(), repository.GetFighterRequest{ID: id})
	if err != nil {
		utils.NotFoundResponse(w, r, err)
		return
	}
	assist, err := h.repository.FighterAssistant.ListByFighterID(r.Context(), id)
	if err != nil {
		utils.InternalServerError(w, r, err)
		return
	}
	fighter.AssistantTrainers = assist
	utils.JsonResponse(w, http.StatusOK, fighter, "Fighter retrieved successfully")
}

func (h *FighterHandler) GetAllFighters(w http.ResponseWriter, r *http.Request) {
	req, limit, offset, err := fighterListRequest(r.URL.Query(), nil, nil)
	if err != nil {
		utils.BadRequestResponse(w, r, err)
		return
	}
	result, err := h.repository.Fighter.GetAllFighters(r.Context(), req)
	if err != nil {
		utils.InternalServerError(w, r, err)
		return
	}
	fighters, err := h.attachAssistantsBatch(r.Context(), result.Fighters)
	if err != nil {
		utils.InternalServerError(w, r, err)
		return
	}
	data := fighterListData{
		Fighters: fighters,
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

func (h *FighterHandler) ExportFightersCSV(w http.ResponseWriter, r *http.Request) {
	limit := 10_000
	offset := 0
	req, _, _, err := fighterListRequest(r.URL.Query(), &limit, &offset)
	if err != nil {
		utils.BadRequestResponse(w, r, err)
		return
	}
	result, err := h.repository.Fighter.GetAllFighters(r.Context(), req)
	if err != nil {
		utils.InternalServerError(w, r, err)
		return
	}
	fighters, err := h.attachAssistantsBatch(r.Context(), result.Fighters)
	if err != nil {
		utils.InternalServerError(w, r, err)
		return
	}

	w.Header().Set("Content-Type", "text/csv; charset=utf-8")
	w.Header().Set("Content-Disposition", `attachment; filename="fighters.csv"`)
	cw := csv.NewWriter(w)
	_ = cw.Write([]string{
		"id", "name", "age", "weight", "wins", "losses", "trainer_id", "fighter_status", "weight_class",
		"license_number", "contract_end", "emergency_contact_name", "emergency_contact_phone",
		"health_notes", "assistant_trainers",
	})
	for _, f := range fighters {
		var ecn, ecp, wc, lic, ce, hn string
		if f.EmergencyContactName != nil {
			ecn = *f.EmergencyContactName
		}
		if f.EmergencyContactPhone != nil {
			ecp = *f.EmergencyContactPhone
		}
		if f.WeightClass != nil {
			wc = *f.WeightClass
		}
		if f.LicenseNumber != nil {
			lic = *f.LicenseNumber
		}
		if f.ContractEnd != nil {
			ce = *f.ContractEnd
		}
		if f.HealthNotes != nil {
			hn = *f.HealthNotes
		}
		var asst []string
		for _, a := range f.AssistantTrainers {
			asst = append(asst, a.TrainerID+":"+a.Role)
		}
		_ = cw.Write([]string{
			f.ID, f.Name, strconv.Itoa(f.Age), fmt.Sprint(f.Weight),
			strconv.Itoa(f.Wins), strconv.Itoa(f.Losses), f.TrainerID, f.FighterStatus,
			wc, lic, ce, ecn, ecp, hn, strings.Join(asst, ";"),
		})
	}
	cw.Flush()
}

type assistantBody struct {
	TrainerID string `json:"trainer_id"`
	Role      string `json:"role"`
}

func (h *FighterHandler) AddAssistantTrainer(w http.ResponseWriter, r *http.Request) {
	fighterID := chi.URLParam(r, "id")
	if fighterID == "" {
		utils.BadRequestResponse(w, r, errors.New("fighter id is required"))
		return
	}
	var body assistantBody
	if err := utils.ReadJSON(w, r, &body); err != nil {
		utils.BadRequestResponse(w, r, err)
		return
	}
	tid := strings.TrimSpace(body.TrainerID)
	if tid == "" {
		utils.BadRequestResponse(w, r, errors.New("trainer_id required"))
		return
	}
	role := strings.TrimSpace(body.Role)
	if role == "" {
		role = "assistant"
	}
	if role != "assistant" && role != "corner" {
		utils.BadRequestResponse(w, r, errors.New("role must be assistant or corner"))
		return
	}
	primary, err := h.repository.FighterAssistant.PrimaryTrainerID(r.Context(), fighterID)
	if err != nil {
		utils.NotFoundResponse(w, r, err)
		return
	}
	if tid == primary {
		utils.BadRequestResponse(w, r, errors.New("trainer is already the primary coach"))
		return
	}
	if err := h.repository.FighterAssistant.Add(r.Context(), fighterID, tid, role); err != nil {
		utils.InternalServerError(w, r, err)
		return
	}
	fighter, err := h.repository.Fighter.GetFighterByID(r.Context(), repository.GetFighterRequest{ID: fighterID})
	if err != nil {
		utils.InternalServerError(w, r, err)
		return
	}
	assist, _ := h.repository.FighterAssistant.ListByFighterID(r.Context(), fighterID)
	fighter.AssistantTrainers = assist
	utils.JsonResponse(w, http.StatusOK, fighter, "assistant trainer linked")
}

func (h *FighterHandler) RemoveAssistantTrainer(w http.ResponseWriter, r *http.Request) {
	fighterID := chi.URLParam(r, "id")
	trainerID := chi.URLParam(r, "trainerID")
	if fighterID == "" || trainerID == "" {
		utils.BadRequestResponse(w, r, errors.New("fighter id and trainer id required"))
		return
	}
	if err := h.repository.FighterAssistant.Remove(r.Context(), fighterID, trainerID); err != nil {
		utils.NotFoundResponse(w, r, err)
		return
	}
	utils.JsonResponse(w, http.StatusOK, nil, "assistant trainer removed")
}

func (h *FighterHandler) UpdateFighter(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	if id == "" {
		utils.BadRequestResponse(w, r, errors.New("fighter id is required"))
		return
	}
	req := repository.UpdateFighterRequest{ID: id}
	if err := utils.ReadJSON(w, r, &req); err != nil {
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
	if err := h.repository.Fighter.DeleteFighter(r.Context(), repository.DeleteFighterRequest{ID: id}); err != nil {
		utils.NotFoundResponse(w, r, err)
		return
	}
	utils.JsonResponse(w, http.StatusOK, nil, "Fighter deleted successfully")
}
