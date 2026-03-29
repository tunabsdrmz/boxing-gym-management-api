package middleware

import (
	"errors"
	"net/http"
	"strings"

	"github.com/tunabsdrmz/boxing-gym-management/internal/auth"
	"github.com/tunabsdrmz/boxing-gym-management/internal/utils"
)

func Authenticate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		h := r.Header.Get("Authorization")
		parts := strings.SplitN(h, " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") || strings.TrimSpace(parts[1]) == "" {
			utils.UnauthorizedErrorResponse(w, r, errors.New("missing bearer token"))
			return
		}
		raw := strings.TrimSpace(parts[1])
		claims, err := auth.ParseToken(raw)
		if err != nil {
			utils.UnauthorizedErrorResponse(w, r, err)
			return
		}
		ctx := auth.ContextWithUserID(r.Context(), claims.UserID)
		ctx = auth.ContextWithRole(ctx, claims.Role)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
