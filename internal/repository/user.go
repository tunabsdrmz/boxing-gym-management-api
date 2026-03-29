package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/lib/pq"
)

type User struct {
	ID           string    `json:"id"`
	Email        string    `json:"email"`
	Role         string    `json:"role"`
	Locked       bool      `json:"locked"`
	LockedReason *string   `json:"locked_reason,omitempty"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// UserWithSecret is used only for login (password verification); never serialize password to JSON.
type UserWithSecret struct {
	User
	PasswordHash string `json:"-"`
}

type CreateUserRequest struct {
	Email        string
	PasswordHash string
	Role         string
}

var (
	ErrEmailTaken           = errors.New("email already registered")
	ErrInvalidCredentials   = errors.New("invalid email or password")
	ErrAccountLocked        = errors.New("account is locked")
	ErrUserNotFound         = errors.New("user not found")
	ErrInvalidRole          = errors.New("invalid role")
)

type UserRepository struct {
	db *sql.DB
}

func (u *UserRepository) CreateUser(ctx context.Context, req CreateUserRequest) (User, error) {
	email := strings.ToLower(strings.TrimSpace(req.Email))
	if email == "" || req.PasswordHash == "" || req.Role == "" {
		return User{}, errors.New("invalid user payload")
	}
	query := `
		INSERT INTO users (email, password_hash, role)
		VALUES ($1, $2, $3)
		RETURNING id::text, email, role, locked, locked_reason, created_at, updated_at
	`
	var out User
	var lockedReason sql.NullString
	err := u.db.QueryRowContext(ctx, query, email, req.PasswordHash, req.Role).Scan(
		&out.ID, &out.Email, &out.Role, &out.Locked, &lockedReason, &out.CreatedAt, &out.UpdatedAt,
	)
	if err != nil {
		var pqErr *pq.Error
		if errors.As(err, &pqErr) && pqErr.Code == "23505" {
			return User{}, ErrEmailTaken
		}
		return User{}, err
	}
	if lockedReason.Valid {
		s := lockedReason.String
		out.LockedReason = &s
	}
	return out, nil
}

func (u *UserRepository) GetUserByEmail(ctx context.Context, email string) (UserWithSecret, error) {
	norm := strings.ToLower(strings.TrimSpace(email))
	query := `
		SELECT id::text, email, role, password_hash, locked, locked_reason, created_at, updated_at
		FROM users
		WHERE email = $1
	`
	var row UserWithSecret
	var lockedReason sql.NullString
	err := u.db.QueryRowContext(ctx, query, norm).Scan(
		&row.ID, &row.Email, &row.Role, &row.PasswordHash, &row.Locked, &lockedReason, &row.CreatedAt, &row.UpdatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return UserWithSecret{}, ErrInvalidCredentials
	}
	if err != nil {
		return UserWithSecret{}, err
	}
	if lockedReason.Valid {
		s := lockedReason.String
		row.LockedReason = &s
	}
	return row, nil
}

func (u *UserRepository) GetUserByID(ctx context.Context, id string) (User, error) {
	query := `
		SELECT id::text, email, role, locked, locked_reason, created_at, updated_at
		FROM users
		WHERE id = $1::uuid
	`
	var out User
	var lockedReason sql.NullString
	err := u.db.QueryRowContext(ctx, query, id).Scan(
		&out.ID, &out.Email, &out.Role, &out.Locked, &lockedReason, &out.CreatedAt, &out.UpdatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return User{}, ErrUserNotFound
	}
	if err != nil {
		return User{}, err
	}
	if lockedReason.Valid {
		s := lockedReason.String
		out.LockedReason = &s
	}
	return out, nil
}

type ListUsersResult struct {
	Users []User
	Total int
}

func (u *UserRepository) ListUsers(ctx context.Context, limit, offset int) (ListUsersResult, error) {
	var total int
	if err := u.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM users`).Scan(&total); err != nil {
		return ListUsersResult{}, err
	}
	rows, err := u.db.QueryContext(ctx, `
		SELECT id::text, email, role, locked, locked_reason, created_at, updated_at
		FROM users
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2
	`, limit, offset)
	if err != nil {
		return ListUsersResult{}, err
	}
	defer rows.Close()
	var users []User
	for rows.Next() {
		var usr User
		var lockedReason sql.NullString
		if err := rows.Scan(&usr.ID, &usr.Email, &usr.Role, &usr.Locked, &lockedReason, &usr.CreatedAt, &usr.UpdatedAt); err != nil {
			return ListUsersResult{}, err
		}
		if lockedReason.Valid {
			s := lockedReason.String
			usr.LockedReason = &s
		}
		users = append(users, usr)
	}
	return ListUsersResult{Users: users, Total: total}, rows.Err()
}

type UpdateUserAdminRequest struct {
	ID           string
	Role         *string
	Locked       *bool
	LockedReason *string
}

func (u *UserRepository) UpdateUserAdmin(ctx context.Context, req UpdateUserAdminRequest) (User, error) {
	var setParts []string
	var args []any
	n := 1
	if req.Role != nil {
		r := strings.TrimSpace(*req.Role)
		if r != "admin" && r != "staff" && r != "viewer" {
			return User{}, ErrInvalidRole
		}
		setParts = append(setParts, fmt.Sprintf("role = $%d", n))
		args = append(args, r)
		n++
	}
	if req.Locked != nil {
		setParts = append(setParts, fmt.Sprintf("locked = $%d", n))
		args = append(args, *req.Locked)
		n++
	}
	if req.LockedReason != nil {
		setParts = append(setParts, fmt.Sprintf("locked_reason = $%d", n))
		args = append(args, *req.LockedReason)
		n++
	}
	if len(setParts) == 0 {
		return u.GetUserByID(ctx, req.ID)
	}
	setParts = append(setParts, "updated_at = now()")
	query := fmt.Sprintf(`
		UPDATE users SET %s WHERE id = $%d::uuid
		RETURNING id::text, email, role, locked, locked_reason, created_at, updated_at
	`, strings.Join(setParts, ", "), n)
	args = append(args, req.ID)
	var out User
	var lockedReason sql.NullString
	err := u.db.QueryRowContext(ctx, query, args...).Scan(
		&out.ID, &out.Email, &out.Role, &out.Locked, &lockedReason, &out.CreatedAt, &out.UpdatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return User{}, ErrUserNotFound
	}
	if err != nil {
		return User{}, err
	}
	if lockedReason.Valid {
		s := lockedReason.String
		out.LockedReason = &s
	}
	return out, nil
}

func (u *UserRepository) SetPasswordHash(ctx context.Context, userID, passwordHash string) error {
	res, err := u.db.ExecContext(ctx, `
		UPDATE users SET password_hash = $1, updated_at = now() WHERE id = $2::uuid
	`, passwordHash, userID)
	if err != nil {
		return err
	}
	aff, _ := res.RowsAffected()
	if aff == 0 {
		return ErrUserNotFound
	}
	return nil
}
