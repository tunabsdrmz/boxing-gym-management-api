package middleware

import (
	"net/http"

	"github.com/tunabsdrmz/boxing-gym-management/internal/auth"
	"github.com/tunabsdrmz/boxing-gym-management/internal/utils"
)

func RequireRoles(allowed ...string) func(http.Handler) http.Handler {
	set := make(map[string]struct{}, len(allowed))
	for _, a := range allowed {
		set[a] = struct{}{}
	}
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			role, ok := auth.RoleFromContext(r.Context())
			if !ok {
				utils.ForbiddenResponse(w, r)
				return
			}
			if _, ok := set[role]; !ok {
				utils.ForbiddenResponse(w, r)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
