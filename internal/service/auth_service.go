package service

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"strings"
	"sync"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

var ErrUnauthorized = errors.New("invalid credentials")
var ErrTooManyAttempts = errors.New("too many failed login attempts")
var ErrInvalidRefreshToken = errors.New("invalid refresh token")

type Role string

const (
	RoleAdmin Role = "admin"
	RoleUser  Role = "user"
)

type Claims struct {
	Role   Role   `json:"role"`
	Email  string `json:"email"`
	UserID string `json:"user_id"`
	jwt.RegisteredClaims
}

type AuthService struct {
	jwtSecret        []byte
	repo             AuthRepository
	accessTTL        time.Duration
	refreshTTL       time.Duration
	maxLoginAttempts int
	lockoutDuration  time.Duration
	mu               sync.Mutex
	loginAttempts    map[string]loginAttempt
}

type AuthRepository interface {
	GetAdminByEmail(ctx context.Context, email string) (AdminAuthRecord, error)
	GetUserByEmail(ctx context.Context, email string) (UserAuthRecord, error)
	CreateRefreshToken(ctx context.Context, token RefreshTokenRecord) (RefreshTokenRecord, error)
	GetRefreshTokenByHash(ctx context.Context, tokenHash string) (RefreshTokenRecord, error)
	RevokeRefreshToken(ctx context.Context, id uuid.UUID) error
	IsNotFound(err error) bool
	Wrap(action string, err error) error
}

type AdminAuthRecord struct {
	ID           uuid.UUID
	Email        string
	PasswordHash string
}

type UserAuthRecord struct {
	ID           uuid.UUID
	Email        string
	PasswordHash string
}

const AdminEmail = "admin@gmail.com"

type RefreshTokenRecord struct {
	ID        uuid.UUID
	UserID    uuid.UUID
	Email     string
	Role      Role
	TokenHash string
	ExpiresAt time.Time
}

type AuthTokens struct {
	AccessToken  string
	RefreshToken string
	Role         Role
}

type loginAttempt struct {
	Failures    int
	LockedUntil time.Time
}

func NewAuthService(secret string, repo AuthRepository, accessTTL time.Duration, refreshTTL time.Duration, maxLoginAttempts int, lockoutDuration time.Duration) *AuthService {
	return &AuthService{
		jwtSecret:        []byte(secret),
		repo:             repo,
		accessTTL:        accessTTL,
		refreshTTL:       refreshTTL,
		maxLoginAttempts: maxLoginAttempts,
		lockoutDuration:  lockoutDuration,
		loginAttempts:    make(map[string]loginAttempt),
	}
}

func (a *AuthService) Login(ctx context.Context, email string, password string) (AuthTokens, error) {
	email = strings.ToLower(strings.TrimSpace(email))
	password = strings.TrimSpace(password)
	if email == "" || password == "" {
		return AuthTokens{}, ErrUnauthorized
	}
	if a.isLocked(email) {
		return AuthTokens{}, ErrTooManyAttempts
	}

	userID, role, ok, err := a.match(ctx, email, password)
	if err != nil {
		return AuthTokens{}, err
	}
	if !ok {
		a.registerFailure(email)
		return AuthTokens{}, ErrUnauthorized
	}
	a.clearFailures(email)

	now := time.Now()
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, Claims{
		Role:   role,
		Email:  email,
		UserID: userID.String(),
		RegisteredClaims: jwt.RegisteredClaims{
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(a.accessTTL)),
		},
	})

	accessToken, err := token.SignedString(a.jwtSecret)
	if err != nil {
		return AuthTokens{}, err
	}

	refreshToken, refreshHash := generateRefreshToken()
	_, err = a.repo.CreateRefreshToken(ctx, RefreshTokenRecord{
		ID:        uuid.New(),
		UserID:    userID,
		Email:     email,
		Role:      role,
		TokenHash: refreshHash,
		ExpiresAt: now.Add(a.refreshTTL),
	})
	if err != nil {
		return AuthTokens{}, a.repo.Wrap("create refresh token", err)
	}

	return AuthTokens{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		Role:         role,
	}, nil
}

func (a *AuthService) Refresh(ctx context.Context, refreshToken string) (AuthTokens, error) {
	refreshToken = strings.TrimSpace(refreshToken)
	if refreshToken == "" {
		return AuthTokens{}, ErrInvalidRefreshToken
	}

	refreshHash := hashToken(refreshToken)
	stored, err := a.repo.GetRefreshTokenByHash(ctx, refreshHash)
	if err != nil {
		if a.repo.IsNotFound(err) {
			return AuthTokens{}, ErrInvalidRefreshToken
		}
		return AuthTokens{}, a.repo.Wrap("get refresh token", err)
	}

	if err := a.repo.RevokeRefreshToken(ctx, stored.ID); err != nil {
		return AuthTokens{}, a.repo.Wrap("revoke refresh token", err)
	}

	now := time.Now()
	accessToken, err := a.signAccessToken(stored.UserID, stored.Email, stored.Role, now)
	if err != nil {
		return AuthTokens{}, err
	}

	nextRefreshToken, nextRefreshHash := generateRefreshToken()
	_, err = a.repo.CreateRefreshToken(ctx, RefreshTokenRecord{
		ID:        uuid.New(),
		UserID:    stored.UserID,
		Email:     stored.Email,
		Role:      stored.Role,
		TokenHash: nextRefreshHash,
		ExpiresAt: now.Add(a.refreshTTL),
	})
	if err != nil {
		return AuthTokens{}, a.repo.Wrap("rotate refresh token", err)
	}

	return AuthTokens{
		AccessToken:  accessToken,
		RefreshToken: nextRefreshToken,
		Role:         stored.Role,
	}, nil
}

func (a *AuthService) ParseToken(rawToken string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(rawToken, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		return a.jwtSecret, nil
	})
	if err != nil {
		return nil, err
	}
	if !token.Valid {
		return nil, ErrUnauthorized
	}

	claims, ok := token.Claims.(*Claims)
	if !ok {
		return nil, ErrUnauthorized
	}
	return claims, nil
}

func (a *AuthService) match(ctx context.Context, email string, password string) (uuid.UUID, Role, bool, error) {
	if email == AdminEmail {
		admin, err := a.repo.GetAdminByEmail(ctx, email)
		if err != nil {
			if a.repo.IsNotFound(err) {
				return uuid.Nil, "", false, nil
			}
			return uuid.Nil, "", false, a.repo.Wrap("get admin", err)
		}
		if err := bcrypt.CompareHashAndPassword([]byte(admin.PasswordHash), []byte(password)); err != nil {
			return uuid.Nil, "", false, nil
		}
		return admin.ID, RoleAdmin, true, nil
	}

	user, err := a.repo.GetUserByEmail(ctx, email)
	if err != nil {
		if a.repo.IsNotFound(err) {
			return uuid.Nil, "", false, nil
		}
		return uuid.Nil, "", false, a.repo.Wrap("get user", err)
	}
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return uuid.Nil, "", false, nil
	}
	return user.ID, RoleUser, true, nil
}

func (a *AuthService) signAccessToken(userID uuid.UUID, email string, role Role, now time.Time) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, Claims{
		Role:   role,
		Email:  email,
		UserID: userID.String(),
		RegisteredClaims: jwt.RegisteredClaims{
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(a.accessTTL)),
		},
	})
	return token.SignedString(a.jwtSecret)
}

func generateRefreshToken() (string, string) {
	token := uuid.NewString() + "." + uuid.NewString()
	return token, hashToken(token)
}

func hashToken(raw string) string {
	sum := sha256.Sum256([]byte(raw))
	return hex.EncodeToString(sum[:])
}

func (a *AuthService) isLocked(key string) bool {
	a.mu.Lock()
	defer a.mu.Unlock()
	state, ok := a.loginAttempts[key]
	if !ok {
		return false
	}
	return time.Now().Before(state.LockedUntil)
}

func (a *AuthService) registerFailure(key string) {
	a.mu.Lock()
	defer a.mu.Unlock()
	now := time.Now()
	state := a.loginAttempts[key]
	if now.Before(state.LockedUntil) {
		return
	}
	state.Failures++
	if state.Failures >= a.maxLoginAttempts {
		state.LockedUntil = now.Add(a.lockoutDuration)
		state.Failures = 0
	}
	a.loginAttempts[key] = state
}

func (a *AuthService) clearFailures(key string) {
	a.mu.Lock()
	defer a.mu.Unlock()
	delete(a.loginAttempts, key)
}
