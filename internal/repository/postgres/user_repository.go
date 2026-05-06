package postgres

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/storex/go-crud/internal/domain"
	sqlcgen "github.com/storex/go-crud/internal/repository/postgres/sqlc"
)

type UserRepository struct {
	queries *sqlcgen.Queries
}

func NewUserRepository(pool *pgxpool.Pool) *UserRepository {
	return &UserRepository{queries: sqlcgen.New(pool)}
}

func (r *UserRepository) Create(ctx context.Context, user domain.User) (domain.User, error) {
	created, err := r.queries.CreateUser(ctx, sqlcgen.CreateUserParams{
		ID:           user.ID,
		Name:         user.Name,
		Phone:        user.Phone,
		Email:        user.Email,
		PasswordHash: user.Password,
	})
	if err != nil {
		return domain.User{}, fmt.Errorf("create user: %w", err)
	}
	return toDomainUser(created), nil
}

func (r *UserRepository) GetByID(ctx context.Context, id uuid.UUID) (domain.User, error) {
	user, err := r.queries.GetUserByID(ctx, id)
	if err != nil {
		return domain.User{}, err
	}
	return toDomainUser(user), nil
}

func (r *UserRepository) UpdateProfile(ctx context.Context, id uuid.UUID, name string, phone string) (domain.User, error) {
	updated, err := r.queries.UpdateUserProfile(ctx, sqlcgen.UpdateUserProfileParams{
		ID:    id,
		Name:  name,
		Phone: phone,
	})
	if err != nil {
		return domain.User{}, fmt.Errorf("update user profile: %w", err)
	}
	return toDomainUser(updated), nil
}

func toDomainUser(row sqlcgen.User) domain.User {
	return domain.User{
		ID:        row.ID,
		Name:      row.Name,
		Phone:     row.Phone,
		Email:     row.Email,
		Password:  row.PasswordHash,
		CreatedAt: row.CreatedAt.Time,
		UpdatedAt: row.UpdatedAt.Time,
	}
}
