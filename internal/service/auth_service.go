package service

import (
	"crypto/subtle"
	"errors"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var ErrUnauthorized = errors.New("invalid credentials")

type Role string

const (
	RoleAdmin Role = "admin"
	RoleUser  Role = "user"
)

type Claims struct {
	Role  Role   `json:"role"`
	Email string `json:"email"`
	jwt.RegisteredClaims
}

type AuthService struct {
	jwtSecret []byte
	adminCred userCred
	userCred  userCred
}

type userCred struct {
	email    string
	password string
	role     Role
}

func NewAuthService(secret string, adminEmail string, adminPassword string, userEmail string, userPassword string) *AuthService {
	return &AuthService{
		jwtSecret: []byte(secret),
		adminCred: userCred{email: adminEmail, password: adminPassword, role: RoleAdmin},
		userCred:  userCred{email: userEmail, password: userPassword, role: RoleUser},
	}
}

func (a *AuthService) Login(email string, password string) (string, Role, error) {
	email = strings.TrimSpace(email)
	password = strings.TrimSpace(password)
	if email == "" || password == "" {
		return "", "", ErrUnauthorized
	}

	cred, ok := a.match(email, password)
	if !ok {
		return "", "", ErrUnauthorized
	}

	now := time.Now()
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, Claims{
		Role:  cred.role,
		Email: cred.email,
		RegisteredClaims: jwt.RegisteredClaims{
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(30 * time.Minute)),
		},
	})

	signed, err := token.SignedString(a.jwtSecret)
	if err != nil {
		return "", "", err
	}
	return signed, cred.role, nil
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

func (a *AuthService) match(email string, password string) (userCred, bool) {
	if secureEqual(email, a.adminCred.email) && secureEqual(password, a.adminCred.password) {
		return a.adminCred, true
	}
	if secureEqual(email, a.userCred.email) && secureEqual(password, a.userCred.password) {
		return a.userCred, true
	}
	return userCred{}, false
}

func secureEqual(left string, right string) bool {
	if len(left) != len(right) {
		return false
	}
	return subtle.ConstantTimeCompare([]byte(left), []byte(right)) == 1
}
