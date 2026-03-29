package handler

import (
	"errors"
	"net/http"
	"strings"

	"github.com/tunabsdrmz/boxing-gym-management/internal/auth"
	"github.com/tunabsdrmz/boxing-gym-management/internal/repository"
	"github.com/tunabsdrmz/boxing-gym-management/internal/utils"
	"golang.org/x/crypto/bcrypt"
)

type AuthHandler struct {
	repository repository.Repository
}

type registerRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type loginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type authResponseData struct {
	Token string          `json:"token"`
	User  repository.User `json:"user"`
}

func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var body registerRequest
	if err := utils.ReadJSON(w, r, &body); err != nil {
		utils.BadRequestResponse(w, r, err)
		return
	}
	email := strings.TrimSpace(body.Email)
	if email == "" || len(body.Password) < 8 {
		utils.BadRequestResponse(w, r, errors.New("valid email and password (min 8 characters) required"))
		return
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(body.Password), bcrypt.DefaultCost)
	if err != nil {
		utils.InternalServerError(w, r, err)
		return
	}
	user, err := h.repository.User.CreateUser(r.Context(), repository.CreateUserRequest{
		Email:        email,
		PasswordHash: string(hash),
		Role:         auth.RoleViewer,
	})
	if err != nil {
		if errors.Is(err, repository.ErrEmailTaken) {
			utils.ConflictResponse(w, r, err)
			return
		}
		utils.InternalServerError(w, r, err)
		return
	}
	token, err := auth.NewToken(user.ID, user.Role)
	if err != nil {
		utils.InternalServerError(w, r, err)
		return
	}
	utils.JsonResponse(w, http.StatusCreated, authResponseData{Token: token, User: user}, "registered successfully")
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var body loginRequest
	if err := utils.ReadJSON(w, r, &body); err != nil {
		utils.BadRequestResponse(w, r, err)
		return
	}
	email := strings.TrimSpace(body.Email)
	if email == "" || body.Password == "" {
		utils.BadRequestResponse(w, r, errors.New("email and password required"))
		return
	}
	row, err := h.repository.User.GetUserByEmail(r.Context(), email)
	if err != nil {
		if errors.Is(err, repository.ErrInvalidCredentials) {
			utils.UnauthorizedErrorResponse(w, r, err)
			return
		}
		utils.InternalServerError(w, r, err)
		return
	}
	if bcrypt.CompareHashAndPassword([]byte(row.PasswordHash), []byte(body.Password)) != nil {
		utils.UnauthorizedErrorResponse(w, r, repository.ErrInvalidCredentials)
		return
	}
	token, err := auth.NewToken(row.ID, row.Role)
	if err != nil {
		utils.InternalServerError(w, r, err)
		return
	}
	utils.JsonResponse(w, http.StatusOK, authResponseData{Token: token, User: row.User}, "login successful")
}
