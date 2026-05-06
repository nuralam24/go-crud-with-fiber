package domain

import (
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
)

var ErrInvalidItem = errors.New("invalid item")

type Item struct {
	ID          uuid.UUID `json:"id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Brand       *Brand    `json:"brand,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

func (i Item) ValidateForCreate() error {
	title := strings.TrimSpace(i.Title)
	if title == "" || len(title) > 120 {
		return ErrInvalidItem
	}
	if len(i.Description) > 4000 {
		return ErrInvalidItem
	}
	return nil
}
