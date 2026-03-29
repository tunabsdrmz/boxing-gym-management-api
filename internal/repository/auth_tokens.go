package repository

import (
	"context"
	"database/sql"
	"errors"
	"time"
)

type AuthTokenRepository struct {
	db *sql.DB
}

func (a *AuthTokenRepository) InsertRefreshToken(ctx context.Context, userID, tokenHash string, expiresAt time.Time) error {
	_, err := a.db.ExecContext(ctx, `
		INSERT INTO refresh_tokens (user_id, token_hash, expires_at)
		VALUES ($1::uuid, $2, $3)
	`, userID, tokenHash, expiresAt)
	return err
}

func (a *AuthTokenRepository) DeleteRefreshToken(ctx context.Context, tokenHash string) error {
	_, err := a.db.ExecContext(ctx, `DELETE FROM refresh_tokens WHERE token_hash = $1`, tokenHash)
	return err
}

// ValidateRefreshToken returns the user id if the hash exists and is not expired.
func (a *AuthTokenRepository) ValidateRefreshToken(ctx context.Context, tokenHash string) (userID string, err error) {
	var uid string
	var exp time.Time
	err = a.db.QueryRowContext(ctx, `
		SELECT user_id::text, expires_at FROM refresh_tokens WHERE token_hash = $1
	`, tokenHash).Scan(&uid, &exp)
	if errors.Is(err, sql.ErrNoRows) {
		return "", ErrInvalidRefreshToken
	}
	if err != nil {
		return "", err
	}
	if time.Now().After(exp) {
		_, _ = a.db.ExecContext(ctx, `DELETE FROM refresh_tokens WHERE token_hash = $1`, tokenHash)
		return "", ErrInvalidRefreshToken
	}
	return uid, nil
}

func (a *AuthTokenRepository) InsertPasswordResetToken(ctx context.Context, userID, tokenHash string, expiresAt time.Time) error {
	_, err := a.db.ExecContext(ctx, `DELETE FROM password_reset_tokens WHERE user_id = $1::uuid`, userID)
	if err != nil {
		return err
	}
	_, err = a.db.ExecContext(ctx, `
		INSERT INTO password_reset_tokens (user_id, token_hash, expires_at)
		VALUES ($1::uuid, $2, $3)
	`, userID, tokenHash, expiresAt)
	return err
}

func (a *AuthTokenRepository) ConsumePasswordResetToken(ctx context.Context, tokenHash string) (userID string, err error) {
	var uid string
	var exp time.Time
	err = a.db.QueryRowContext(ctx, `
		SELECT user_id::text, expires_at FROM password_reset_tokens WHERE token_hash = $1
	`, tokenHash).Scan(&uid, &exp)
	if errors.Is(err, sql.ErrNoRows) {
		return "", ErrInvalidResetToken
	}
	if err != nil {
		return "", err
	}
	if time.Now().After(exp) {
		_, _ = a.db.ExecContext(ctx, `DELETE FROM password_reset_tokens WHERE token_hash = $1`, tokenHash)
		return "", ErrInvalidResetToken
	}
	_, err = a.db.ExecContext(ctx, `DELETE FROM password_reset_tokens WHERE token_hash = $1`, tokenHash)
	if err != nil {
		return "", err
	}
	return uid, nil
}

var (
	ErrInvalidRefreshToken = errors.New("invalid or expired refresh token")
	ErrInvalidResetToken   = errors.New("invalid or expired reset token")
)

func (a *AuthTokenRepository) RevokeRefreshForUser(ctx context.Context, userID string) error {
	_, err := a.db.ExecContext(ctx, `DELETE FROM refresh_tokens WHERE user_id = $1::uuid`, userID)
	return err
}
