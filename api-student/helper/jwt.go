package helper

import (
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var (
	ErrTokenInvalid = errors.New("token tidak valid")
	ErrTokenExpired = errors.New("token kedaluwarsa")
)

type Claims struct {
	UserID   int    `json:"uid"`
	Username string `json:"usr"`
	Role     string `json:"rol"`
	jwt.RegisteredClaims
}

type TokenIssuer struct {
	secret []byte
	ttl    time.Duration
}

func NewTokenIssuer(secret string, ttl time.Duration) *TokenIssuer {
	return &TokenIssuer{
		secret: []byte(secret),
		ttl:    ttl,
	}
}

func (t *TokenIssuer) Issue(userID int, username, role string) (string, time.Time, error) {
	exp := time.Now().Add(t.ttl)
	claims := Claims{
		UserID:   userID,
		Username: username,
		Role:     role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(exp),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    "api-student",
			Subject:   fmt.Sprintf("%d", userID),
		},
	}

	tok := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := tok.SignedString(t.secret)
	if err != nil {
		return "", time.Time{}, fmt.Errorf("gagal menandatangani token: %w", err)
	}

	return signed, exp, nil
}

func (t *TokenIssuer) Parse(raw string) (*Claims, error) {
	claims := &Claims{}

	parser := jwt.NewParser(
		jwt.WithValidMethods([]string{"HS256"}),
	)

	_, err := parser.ParseWithClaims(raw, claims, func(token *jwt.Token) (any, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, ErrTokenInvalid
		}
		return t.secret, nil
	})

	if err != nil {
		switch {
		case errors.Is(err, jwt.ErrTokenExpired):
			return nil, ErrTokenExpired
		default:
			return nil, ErrTokenInvalid
		}
	}

	return claims, nil
}

func (t *TokenIssuer) TTL() time.Duration {
	return t.ttl
}
