package service

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"

	"github.com/storex/go-crud/internal/domain"
)

type BrandRepository interface {
	Create(ctx context.Context, brand domain.Brand) (domain.Brand, error)
	List(ctx context.Context) ([]domain.Brand, error)
}

type BrandService struct {
	repo BrandRepository
}

func NewBrandService(repo BrandRepository) *BrandService {
	return &BrandService{repo: repo}
}

func (s *BrandService) Create(ctx context.Context, name string, slug string) (domain.Brand, error) {
	brand := domain.Brand{
		ID:   uuid.New(),
		Name: strings.TrimSpace(name),
		Slug: strings.TrimSpace(slug),
	}
	if err := brand.ValidateForCreate(); err != nil {
		return domain.Brand{}, err
	}
	created, err := s.repo.Create(ctx, brand)
	if err != nil {
		return domain.Brand{}, fmt.Errorf("create brand: %w", err)
	}
	return created, nil
}

func (s *BrandService) List(ctx context.Context) ([]domain.Brand, error) {
	return s.repo.List(ctx)
}
