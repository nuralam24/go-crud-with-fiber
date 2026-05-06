package domain

import (
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
)

var ErrInvalidBrand = errors.New("invalid brand")

type Brand struct {
	ID        uuid.UUID `json:"id"`
	Name      string    `json:"name"`
	Slug      string    `json:"slug"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (b Brand) ValidateForCreate() error {
	if strings.TrimSpace(b.Name) == "" || strings.TrimSpace(b.Slug) == "" {
		return ErrInvalidBrand
	}
	return nil
}
