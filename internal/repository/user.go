package repository

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"time"

	"github.com/lib/pq"
)

type User struct {
	ID        string    `json:"id"`
	Email     string    `json:"email"`
	Role      string    `json:"role"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// UserWithSecret is used only for login (password verification); never serialize to JSON.
type UserWithSecret struct {
	User
	PasswordHash string
}

type CreateUserRequest struct {
	Email        string
	PasswordHash string
	Role         string
}

var (
	ErrEmailTaken         = errors.New("email already registered")
	ErrInvalidCredentials = errors.New("invalid email or password")
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
		RETURNING id::text, email, role, created_at, updated_at
	`
	var out User
	err := u.db.QueryRowContext(ctx, query, email, req.PasswordHash, req.Role).Scan(
		&out.ID, &out.Email, &out.Role, &out.CreatedAt, &out.UpdatedAt,
	)
	if err != nil {
		var pqErr *pq.Error
		if errors.As(err, &pqErr) && pqErr.Code == "23505" {
			return User{}, ErrEmailTaken
		}
		return User{}, err
	}
	return out, nil
}

func (u *UserRepository) GetUserByEmail(ctx context.Context, email string) (UserWithSecret, error) {
	norm := strings.ToLower(strings.TrimSpace(email))
	query := `
		SELECT id::text, email, role, password_hash, created_at, updated_at
		FROM users
		WHERE email = $1
	`
	var row UserWithSecret
	err := u.db.QueryRowContext(ctx, query, norm).Scan(
		&row.ID, &row.Email, &row.Role, &row.PasswordHash, &row.CreatedAt, &row.UpdatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return UserWithSecret{}, ErrInvalidCredentials
	}
	if err != nil {
		return UserWithSecret{}, err
	}
	return row, nil
}
