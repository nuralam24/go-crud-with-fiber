package service

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"golang.org/x/crypto/bcrypt"

	"github.com/storex/go-crud/internal/domain"
)

var ErrUserExists = errors.New("user already exists")

type UserRepository interface {
	Create(ctx context.Context, user domain.User) (domain.User, error)
	GetByID(ctx context.Context, id uuid.UUID) (domain.User, error)
	UpdateProfile(ctx context.Context, id uuid.UUID, name string, phone string) (domain.User, error)
}

type UserService struct {
	repo UserRepository
}

func NewUserService(repo UserRepository) *UserService {
	return &UserService{repo: repo}
}

func (s *UserService) Register(ctx context.Context, name string, phone string, email string, password string) (domain.User, error) {
	email = strings.ToLower(strings.TrimSpace(email))
	password = strings.TrimSpace(password)

	user := domain.User{
		ID:       uuid.New(),
		Name:     strings.TrimSpace(name),
		Phone:    strings.TrimSpace(phone),
		Email:    email,
		Password: password,
	}
	if err := user.ValidateForRegister(); err != nil {
		return domain.User{}, err
	}

	passwordHash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return domain.User{}, fmt.Errorf("hash password: %w", err)
	}
	user.Password = string(passwordHash)

	created, err := s.repo.Create(ctx, user)
	if err != nil {
		if strings.Contains(err.Error(), "duplicate key") {
			return domain.User{}, ErrUserExists
		}
		return domain.User{}, fmt.Errorf("register user: %w", err)
	}
	return created, nil
}

func (s *UserService) GetMyProfile(ctx context.Context, id uuid.UUID) (domain.User, error) {
	user, err := s.repo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.User{}, ErrNotFound
		}
		return domain.User{}, err
	}
	return user, nil
}

func (s *UserService) UpdateMyProfile(ctx context.Context, id uuid.UUID, name string, phone string) (domain.User, error) {
	name = strings.TrimSpace(name)
	phone = strings.TrimSpace(phone)
	if err := domain.ValidateUserProfile(name, phone); err != nil {
		return domain.User{}, err
	}
	updated, err := s.repo.UpdateProfile(ctx, id, name, phone)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.User{}, ErrNotFound
		}
		return domain.User{}, err
	}
	return updated, nil
}
