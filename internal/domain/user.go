package domain

import (
	"errors"
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"
)

var ErrInvalidUser = errors.New("invalid user")
var emailRegex = regexp.MustCompile(`^[^\s@]+@[^\s@]+\.[^\s@]+$`)
var phoneRegex = regexp.MustCompile(`^\+?[0-9]{8,15}$`)

type User struct {
	ID        uuid.UUID `json:"id"`
	Name      string    `json:"name"`
	Phone     string    `json:"phone"`
	Email     string    `json:"email"`
	Password  string    `json:"-"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (u User) ValidateForRegister() error {
	name := strings.TrimSpace(u.Name)
	phone := strings.TrimSpace(u.Phone)
	email := strings.TrimSpace(strings.ToLower(u.Email))
	password := strings.TrimSpace(u.Password)

	if len(name) < 2 || len(name) > 120 {
		return ErrInvalidUser
	}
	if !phoneRegex.MatchString(phone) {
		return ErrInvalidUser
	}
	if !emailRegex.MatchString(email) {
		return ErrInvalidUser
	}
	if len(password) < 8 || len(password) > 72 {
		return ErrInvalidUser
	}
	return nil
}

func ValidateUserProfile(name string, phone string) error {
	name = strings.TrimSpace(name)
	phone = strings.TrimSpace(phone)
	if len(name) < 2 || len(name) > 120 {
		return ErrInvalidUser
	}
	if !phoneRegex.MatchString(phone) {
		return ErrInvalidUser
	}
	return nil
}
