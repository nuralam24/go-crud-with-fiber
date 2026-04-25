package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/storex/go-crud/internal/domain"
	"github.com/storex/go-crud/internal/platform/async"
)

var ErrNotFound = errors.New("resource not found")

type ItemRepository interface {
	Create(ctx context.Context, item domain.Item) (domain.Item, error)
	GetByID(ctx context.Context, id uuid.UUID) (domain.Item, error)
	List(ctx context.Context, limit int, offset int) ([]domain.Item, error)
}

type ItemService struct {
	repo        ItemRepository
	auditLogger *async.AuditLogger
}

func NewItemService(repo ItemRepository, auditLogger *async.AuditLogger) *ItemService {
	return &ItemService{
		repo:        repo,
		auditLogger: auditLogger,
	}
}

func (s *ItemService) Create(ctx context.Context, title string, description string, actor string) (domain.Item, error) {
	item := domain.Item{
		ID:          uuid.New(),
		Title:       title,
		Description: description,
	}
	if err := item.ValidateForCreate(); err != nil {
		return domain.Item{}, err
	}

	created, err := s.repo.Create(ctx, item)
	if err != nil {
		return domain.Item{}, fmt.Errorf("create item: %w", err)
	}

	s.auditLogger.Publish(async.AuditEvent{
		Action: "create-item",
		Actor:  actor,
		Target: created.ID.String(),
	})

	return created, nil
}

func (s *ItemService) GetByID(ctx context.Context, id uuid.UUID) (domain.Item, error) {
	item, err := s.repo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.Item{}, ErrNotFound
		}
		return domain.Item{}, err
	}
	return item, nil
}

func (s *ItemService) List(ctx context.Context, limit int, offset int) ([]domain.Item, error) {
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}
	if offset < 0 {
		offset = 0
	}
	return s.repo.List(ctx, limit, offset)
}
