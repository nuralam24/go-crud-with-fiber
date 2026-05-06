package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"

	sqlcgen "github.com/storex/go-crud/internal/repository/postgres/sqlc"
	"github.com/storex/go-crud/internal/service"
)

type AuthRepository struct {
	queries *sqlcgen.Queries
}

func NewAuthRepository(pool *pgxpool.Pool) *AuthRepository {
	return &AuthRepository{queries: sqlcgen.New(pool)}
}

func (r *AuthRepository) GetAdminByEmail(ctx context.Context, email string) (service.AdminAuthRecord, error) {
	admin, err := r.queries.GetAdminByEmail(ctx, email)
	if err != nil {
		return service.AdminAuthRecord{}, err
	}
	return service.AdminAuthRecord{
		ID:           admin.ID,
		Email:        admin.Email,
		PasswordHash: admin.PasswordHash,
	}, nil
}

func (r *AuthRepository) GetUserByEmail(ctx context.Context, email string) (service.UserAuthRecord, error) {
	user, err := r.queries.GetUserByEmail(ctx, email)
	if err != nil {
		return service.UserAuthRecord{}, err
	}
	return service.UserAuthRecord{
		ID:           user.ID,
		Email:        user.Email,
		PasswordHash: user.PasswordHash,
	}, nil
}

func (r *AuthRepository) CreateRefreshToken(ctx context.Context, token service.RefreshTokenRecord) (service.RefreshTokenRecord, error) {
	created, err := r.queries.CreateRefreshToken(ctx, sqlcgen.CreateRefreshTokenParams{
		ID:        token.ID,
		UserID:    token.UserID,
		Email:     token.Email,
		Role:      string(token.Role),
		TokenHash: token.TokenHash,
		ExpiresAt: pgtype.Timestamptz{Time: token.ExpiresAt, Valid: true},
	})
	if err != nil {
		return service.RefreshTokenRecord{}, err
	}
	return service.RefreshTokenRecord{
		ID:        created.ID,
		UserID:    created.UserID,
		Email:     created.Email,
		Role:      service.Role(created.Role),
		TokenHash: created.TokenHash,
		ExpiresAt: created.ExpiresAt.Time,
	}, nil
}

func (r *AuthRepository) GetRefreshTokenByHash(ctx context.Context, tokenHash string) (service.RefreshTokenRecord, error) {
	row, err := r.queries.GetRefreshTokenByHash(ctx, tokenHash)
	if err != nil {
		return service.RefreshTokenRecord{}, err
	}
	return service.RefreshTokenRecord{
		ID:        row.ID,
		UserID:    row.UserID,
		Email:     row.Email,
		Role:      service.Role(row.Role),
		TokenHash: row.TokenHash,
		ExpiresAt: row.ExpiresAt.Time,
	}, nil
}

func (r *AuthRepository) RevokeRefreshToken(ctx context.Context, id uuid.UUID) error {
	return r.queries.RevokeRefreshToken(ctx, id)
}

func (r *AuthRepository) IsNotFound(err error) bool {
	return errors.Is(err, pgx.ErrNoRows)
}

func (r *AuthRepository) Wrap(action string, err error) error {
	return fmt.Errorf("%s: %w", action, err)
}
