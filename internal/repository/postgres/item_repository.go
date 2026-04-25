package postgres

import (
	"context"
	"fmt"

	"github.com/google/uuid"
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

func (r *ItemRepository) Create(ctx context.Context, item domain.Item) (domain.Item, error) {
	created, err := r.queries.CreateItem(ctx, sqlcgen.CreateItemParams{
		ID:          item.ID,
		Title:       item.Title,
		Description: item.Description,
	})
	if err != nil {
		return domain.Item{}, fmt.Errorf("insert item: %w", err)
	}
	return toDomainItem(created), nil
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
		items = append(items, toDomainItem(row))
	}
	return items, nil
}

func toDomainItem(item sqlcgen.Item) domain.Item {
	return domain.Item{
		ID:          item.ID,
		Title:       item.Title,
		Description: item.Description,
		CreatedAt:   item.CreatedAt.Time,
		UpdatedAt:   item.UpdatedAt.Time,
	}
}
