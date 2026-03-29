package handler

import (
	"context"
	"errors"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/tunabsdrmz/boxing-gym-management/internal/auth"
	"github.com/tunabsdrmz/boxing-gym-management/internal/config"
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

type tokenResponseData struct {
	AccessToken  string          `json:"access_token"`
	RefreshToken string          `json:"refresh_token"`
	Token        string          `json:"token"`
	ExpiresIn    int             `json:"expires_in"`
	User         repository.User `json:"user"`
}

type refreshRequest struct {
	RefreshToken string `json:"refresh_token"`
}

type forgotPasswordRequest struct {
	Email string `json:"email"`
}

type resetPasswordRequest struct {
	Token       string `json:"token"`
	NewPassword string `json:"new_password"`
}

func refreshExpiry() time.Time {
	d := config.App.JWT.RefreshTTL
	if d <= 0 {
		d = 7 * 24 * time.Hour
	}
	return time.Now().Add(d)
}

func (h *AuthHandler) issueTokens(ctx context.Context, userID, role string) (tokenResponseData, error) {
	access, err := auth.NewAccessToken(userID, role)
	if err != nil {
		return tokenResponseData{}, err
	}
	plain, err := auth.RandomHex(32)
	if err != nil {
		return tokenResponseData{}, err
	}
	hash := auth.HashToken(plain)
	if err := h.repository.AuthToken.InsertRefreshToken(ctx, userID, hash, refreshExpiry()); err != nil {
		return tokenResponseData{}, err
	}
	u, err := h.repository.User.GetUserByID(ctx, userID)
	if err != nil {
		return tokenResponseData{}, err
	}
	return tokenResponseData{
		AccessToken:  access,
		RefreshToken: plain,
		Token:        access,
		ExpiresIn:    auth.AccessExpiresInSeconds(),
		User:         u,
	}, nil
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
	tokens, err := h.issueTokens(r.Context(), user.ID, user.Role)
	if err != nil {
		utils.InternalServerError(w, r, err)
		return
	}
	utils.JsonResponse(w, http.StatusCreated, tokens, "registered successfully")
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
	if row.Locked {
		utils.WriteJSONError(w, http.StatusForbidden, "account is locked")
		return
	}
	tokens, err := h.issueTokens(r.Context(), row.ID, row.Role)
	if err != nil {
		utils.InternalServerError(w, r, err)
		return
	}
	utils.JsonResponse(w, http.StatusOK, tokens, "login successful")
}

func (h *AuthHandler) Refresh(w http.ResponseWriter, r *http.Request) {
	var body refreshRequest
	if err := utils.ReadJSON(w, r, &body); err != nil {
		utils.BadRequestResponse(w, r, err)
		return
	}
	t := strings.TrimSpace(body.RefreshToken)
	if t == "" {
		utils.BadRequestResponse(w, r, errors.New("refresh_token required"))
		return
	}
	hash := auth.HashToken(t)
	userID, err := h.repository.AuthToken.ValidateRefreshToken(r.Context(), hash)
	if err != nil {
		utils.UnauthorizedErrorResponse(w, r, err)
		return
	}
	_ = h.repository.AuthToken.DeleteRefreshToken(r.Context(), hash)
	u, err := h.repository.User.GetUserByID(r.Context(), userID)
	if err != nil {
		utils.UnauthorizedErrorResponse(w, r, err)
		return
	}
	if u.Locked {
		utils.WriteJSONError(w, http.StatusForbidden, "account is locked")
		return
	}
	tokens, err := h.issueTokens(r.Context(), u.ID, u.Role)
	if err != nil {
		utils.InternalServerError(w, r, err)
		return
	}
	utils.JsonResponse(w, http.StatusOK, tokens, "token refreshed")
}

func (h *AuthHandler) ForgotPassword(w http.ResponseWriter, r *http.Request) {
	var body forgotPasswordRequest
	if err := utils.ReadJSON(w, r, &body); err != nil {
		utils.BadRequestResponse(w, r, err)
		return
	}
	email := strings.TrimSpace(body.Email)
	msg := "If an account exists for this email, password reset instructions will follow."
	if email == "" {
		utils.JsonResponse(w, http.StatusOK, map[string]any{}, msg)
		return
	}
	row, err := h.repository.User.GetUserByEmail(r.Context(), email)
	if err != nil {
		if errors.Is(err, repository.ErrInvalidCredentials) {
			utils.JsonResponse(w, http.StatusOK, map[string]any{}, msg)
			return
		}
		utils.InternalServerError(w, r, err)
		return
	}
	plain, err := auth.RandomHex(32)
	if err != nil {
		utils.InternalServerError(w, r, err)
		return
	}
	hash := auth.HashToken(plain)
	exp := time.Now().Add(1 * time.Hour)
	if err := h.repository.AuthToken.InsertPasswordResetToken(r.Context(), row.ID, hash, exp); err != nil {
		utils.InternalServerError(w, r, err)
		return
	}
	data := map[string]any{}
	if config.App.Dev.ReturnPasswordResetToken {
		data["reset_token"] = plain
		log.Println("password reset token (dev)", "email", row.Email, "token", plain)
	} else {
		log.Println("password reset issued", "email", row.Email, "user_id", row.ID)
	}
	utils.JsonResponse(w, http.StatusOK, data, msg)
}

func (h *AuthHandler) ResetPassword(w http.ResponseWriter, r *http.Request) {
	var body resetPasswordRequest
	if err := utils.ReadJSON(w, r, &body); err != nil {
		utils.BadRequestResponse(w, r, err)
		return
	}
	if len(body.NewPassword) < 8 {
		utils.BadRequestResponse(w, r, errors.New("new_password must be at least 8 characters"))
		return
	}
	t := strings.TrimSpace(body.Token)
	if t == "" {
		utils.BadRequestResponse(w, r, errors.New("token required"))
		return
	}
	hash := auth.HashToken(t)
	userID, err := h.repository.AuthToken.ConsumePasswordResetToken(r.Context(), hash)
	if err != nil {
		utils.BadRequestResponse(w, r, err)
		return
	}
	pw, err := bcrypt.GenerateFromPassword([]byte(body.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		utils.InternalServerError(w, r, err)
		return
	}
	if err := h.repository.User.SetPasswordHash(r.Context(), userID, string(pw)); err != nil {
		utils.InternalServerError(w, r, err)
		return
	}
	_ = h.repository.AuthToken.RevokeRefreshForUser(r.Context(), userID)
	utils.JsonResponse(w, http.StatusOK, map[string]any{}, "password updated; please sign in again")
}
