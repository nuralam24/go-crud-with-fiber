package postgres

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/storex/go-crud/internal/domain"
	sqlcgen "github.com/storex/go-crud/internal/repository/postgres/sqlc"
)

type BrandRepository struct {
	queries *sqlcgen.Queries
}

func NewBrandRepository(pool *pgxpool.Pool) *BrandRepository {
	return &BrandRepository{queries: sqlcgen.New(pool)}
}

func (r *BrandRepository) Create(ctx context.Context, brand domain.Brand) (domain.Brand, error) {
	created, err := r.queries.CreateBrand(ctx, sqlcgen.CreateBrandParams{
		ID:   brand.ID,
		Name: brand.Name,
		Slug: brand.Slug,
	})
	if err != nil {
		return domain.Brand{}, fmt.Errorf("create brand: %w", err)
	}
	return toDomainBrand(created), nil
}

func (r *BrandRepository) List(ctx context.Context) ([]domain.Brand, error) {
	rows, err := r.queries.ListBrands(ctx)
	if err != nil {
		return nil, fmt.Errorf("list brands: %w", err)
	}
	brands := make([]domain.Brand, 0, len(rows))
	for _, row := range rows {
		brands = append(brands, toDomainBrand(row))
	}
	return brands, nil
}

func toDomainBrand(row sqlcgen.Brand) domain.Brand {
	return domain.Brand{
		ID:        row.ID,
		Name:      row.Name,
		Slug:      row.Slug,
		CreatedAt: row.CreatedAt.Time,
		UpdatedAt: row.UpdatedAt.Time,
	}
}
