package postgres

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/storex/go-crud/internal/domain"
	sqlcgen "github.com/storex/go-crud/internal/repository/postgres/sqlc"
)

type ItemRepository struct {
	queries *sqlcgen.Queries
}

func NewItemRepository(pool *pgxpool.Pool) *ItemRepository {
	return &ItemRepository{
		queries: sqlcgen.New(pool),
	}
}

func (r *ItemRepository) Create(ctx context.Context, item domain.Item, brandID *uuid.UUID) (domain.Item, error) {
	params := sqlcgen.CreateItemParams{
		ID:          item.ID,
		Title:       item.Title,
		Description: item.Description,
	}
	if brandID != nil {
		params.BrandID = pgtype.UUID{Bytes: *brandID, Valid: true}
	}

	created, err := r.queries.CreateItem(ctx, params)
	if err != nil {
		return domain.Item{}, fmt.Errorf("insert item: %w", err)
	}
	return toDomainItemFromModel(created), nil
}

func (r *ItemRepository) GetByID(ctx context.Context, id uuid.UUID) (domain.Item, error) {
	item, err := r.queries.GetItemByID(ctx, id)
	if err != nil {
		return domain.Item{}, err
	}
	return toDomainItem(item), nil
}

func (r *ItemRepository) List(ctx context.Context, limit int, offset int) ([]domain.Item, error) {
	rows, err := r.queries.ListItems(ctx, sqlcgen.ListItemsParams{
		Limit:  int32(limit),
		Offset: int32(offset),
	})
	if err != nil {
		return nil, fmt.Errorf("list items: %w", err)
	}
	items := make([]domain.Item, 0, len(rows))
	for _, row := range rows {
		items = append(items, toDomainItemFromList(row))
	}
	return items, nil
}

func (r *ItemRepository) Count(ctx context.Context) (int, error) {
	total, err := r.queries.CountItems(ctx)
	if err != nil {
		return 0, fmt.Errorf("count items: %w", err)
	}
	return int(total), nil
}

func toDomainItemFromModel(item sqlcgen.Item) domain.Item {
	domainItem := domain.Item{
		ID:          item.ID,
		Title:       item.Title,
		Description: item.Description,
		CreatedAt:   item.CreatedAt.Time,
		UpdatedAt:   item.UpdatedAt.Time,
	}
	if item.BrandID.Valid {
		domainItem.Brand = &domain.Brand{
			ID: item.BrandID.Bytes,
		}
	}
	return domainItem
}

func toDomainItem(row sqlcgen.GetItemByIDRow) domain.Item {
	domainItem := domain.Item{
		ID:          row.ID,
		Title:       row.Title,
		Description: row.Description,
		CreatedAt:   row.CreatedAt.Time,
		UpdatedAt:   row.UpdatedAt.Time,
	}
	if row.BrandRefID.Valid {
		domainItem.Brand = &domain.Brand{
			ID:   row.BrandRefID.Bytes,
			Name: row.BrandName.String,
			Slug: row.BrandSlug.String,
		}
	}
	return domainItem
}

func toDomainItemFromList(row sqlcgen.ListItemsRow) domain.Item {
	domainItem := domain.Item{
		ID:          row.ID,
		Title:       row.Title,
		Description: row.Description,
		CreatedAt:   row.CreatedAt.Time,
		UpdatedAt:   row.UpdatedAt.Time,
	}
	if row.BrandRefID.Valid {
		domainItem.Brand = &domain.Brand{
			ID:   row.BrandRefID.Bytes,
			Name: row.BrandName.String,
			Slug: row.BrandSlug.String,
		}
	}
	return domainItem
}
