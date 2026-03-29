package auth

import (
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/tunabsdrmz/boxing-gym-management/internal/config"
)

var ErrInvalidToken = errors.New("invalid token")

type Claims struct {
	UserID string `json:"sub"`
	Role   string `json:"role"`
	jwt.RegisteredClaims
}

const defaultTokenTTL = 24 * time.Hour

func NewToken(userID, role string) (string, error) {
	if config.App.JWT.Secret == "" {
		return "", errors.New("jwt secret is not configured")
	}
	now := time.Now()
	claims := Claims{
		UserID: userID,
		Role:   role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(defaultTokenTTL)),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
		},
	}
	t := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return t.SignedString([]byte(config.App.JWT.Secret))
}

func ParseToken(tokenString string) (*Claims, error) {
	if config.App.JWT.Secret == "" {
		return nil, errors.New("jwt secret is not configured")
	}
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return []byte(config.App.JWT.Secret), nil
	})
	if err != nil {
		return nil, ErrInvalidToken
	}
	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, ErrInvalidToken
	}
	if claims.UserID == "" || claims.Role == "" {
		return nil, ErrInvalidToken
	}
	return claims, nil
}
