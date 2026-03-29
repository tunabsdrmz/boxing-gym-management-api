package auth

import "context"

type ctxKey int

const (
	ctxKeyUserID ctxKey = iota
	ctxKeyRole
)

func ContextWithUserID(ctx context.Context, userID string) context.Context {
	return context.WithValue(ctx, ctxKeyUserID, userID)
}

func UserIDFromContext(ctx context.Context) (string, bool) {
	v, ok := ctx.Value(ctxKeyUserID).(string)
	return v, ok && v != ""
}

func ContextWithRole(ctx context.Context, role string) context.Context {
	return context.WithValue(ctx, ctxKeyRole, role)
}

func RoleFromContext(ctx context.Context) (string, bool) {
	v, ok := ctx.Value(ctxKeyRole).(string)
	return v, ok && v != ""
}
