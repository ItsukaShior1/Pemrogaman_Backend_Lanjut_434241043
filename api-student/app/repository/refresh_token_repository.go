package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"api-student/app/model"
)

type RefreshTokenRepository interface {
	Create(ctx context.Context, rt model.RefreshToken) (model.RefreshToken, error)
	FindByHash(ctx context.Context, hash string) (model.RefreshToken, error)
	Revoke(ctx context.Context, id int, replacedBy *int) error
	RevokeAllForUser(ctx context.Context, userID int) error
}

type refreshTokenPostgresRepository struct {
	pool *pgxpool.Pool
}

func NewRefreshTokenRepository(pool *pgxpool.Pool) RefreshTokenRepository {
	return &refreshTokenPostgresRepository{pool: pool}
}

func (r *refreshTokenPostgresRepository) Create(ctx context.Context, rt model.RefreshToken) (model.RefreshToken, error) {
	err := r.pool.QueryRow(
		ctx,
		`INSERT INTO refresh_tokens (user_id, token_hash, expires_at)
		 VALUES ($1, $2, $3)
		 RETURNING id, created_at`,
		rt.UserID, rt.TokenHash, rt.ExpiresAt,
	).Scan(&rt.ID, &rt.CreatedAt)

	if err != nil {
		return model.RefreshToken{}, fmt.Errorf("menyimpan refresh token: %w", err)
	}

	return rt, nil
}

func (r *refreshTokenPostgresRepository) FindByHash(ctx context.Context, hash string) (model.RefreshToken, error) {
	var rt model.RefreshToken
	var revokedAt *time.Time
	err := r.pool.QueryRow(
		ctx,
		`SELECT id, user_id, token_hash, expires_at, revoked_at, replaced_by, created_at
		 FROM refresh_tokens WHERE token_hash = $1`,
		hash,
	).Scan(&rt.ID, &rt.UserID, &rt.TokenHash, &rt.ExpiresAt, &revokedAt, &rt.ReplacedBy, &rt.CreatedAt)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.RefreshToken{}, ErrNotFound
		}
		return model.RefreshToken{}, fmt.Errorf("mengambil refresh token: %w", err)
	}

	rt.RevokedAt = revokedAt
	return rt, nil
}

func (r *refreshTokenPostgresRepository) Revoke(ctx context.Context, id int, replacedBy *int) error {
	tag, err := r.pool.Exec(
		ctx,
		`UPDATE refresh_tokens
		 SET revoked_at = NOW(), replaced_by = $1
		 WHERE id = $2 AND revoked_at IS NULL`,
		replacedBy, id,
	)
	if err != nil {
		return fmt.Errorf("mencabut refresh token: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *refreshTokenPostgresRepository) RevokeAllForUser(ctx context.Context, userID int) error {
	_, err := r.pool.Exec(
		ctx,
		`UPDATE refresh_tokens
		 SET revoked_at = NOW()
		 WHERE user_id = $1 AND revoked_at IS NULL`,
		userID,
	)
	if err != nil {
		return fmt.Errorf("mencabut semua refresh token: %w", err)
	}
	return nil
}
