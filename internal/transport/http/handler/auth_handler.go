package handler

import (
	"strings"

	"github.com/gofiber/fiber/v2"

	"github.com/storex/go-crud/internal/service"
	"github.com/storex/go-crud/internal/transport/apierror"
	"github.com/storex/go-crud/internal/transport/response"
)

type AuthHandler struct {
	authService *service.AuthService
}

type loginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type refreshRequest struct {
	RefreshToken string `json:"refresh_token"`
}

func NewAuthHandler(authService *service.AuthService) *AuthHandler {
	return &AuthHandler{authService: authService}
}

func (h *AuthHandler) AdminLogin(c *fiber.Ctx) error {
	var req loginRequest
	if err := c.BodyParser(&req); err != nil {
		return apierror.New(fiber.StatusBadRequest, "invalid_payload", "invalid payload")
	}
	req.Email = strings.TrimSpace(req.Email)
	req.Password = strings.TrimSpace(req.Password)
	if req.Email == "" || req.Password == "" {
		return apierror.New(fiber.StatusBadRequest, "validation_error", "email and password are required")
	}
	if !strings.EqualFold(req.Email, service.AdminEmail) {
		return apierror.New(fiber.StatusUnauthorized, "invalid_credentials", "invalid admin credentials")
	}

	tokens, err := h.authService.Login(c.Context(), req.Email, req.Password)
	if err != nil {
		if err == service.ErrTooManyAttempts {
			return apierror.New(fiber.StatusTooManyRequests, "too_many_attempts", "account temporarily locked due to failed login attempts")
		}
		return apierror.New(fiber.StatusUnauthorized, "invalid_credentials", "invalid credentials")
	}

	return response.Single(c, fiber.StatusOK, fiber.Map{
		"access_token":  tokens.AccessToken,
		"refresh_token": tokens.RefreshToken,
		"role":          tokens.Role,
	})
}

func (h *AuthHandler) AdminRefresh(c *fiber.Ctx) error {
	var req refreshRequest
	if err := c.BodyParser(&req); err != nil {
		return apierror.New(fiber.StatusBadRequest, "invalid_payload", "invalid payload")
	}

	tokens, err := h.authService.Refresh(c.Context(), req.RefreshToken)
	if err != nil {
		return apierror.New(fiber.StatusUnauthorized, "invalid_refresh_token", "invalid refresh token")
	}
	if tokens.Role != service.RoleAdmin {
		return apierror.New(fiber.StatusUnauthorized, "invalid_refresh_token", "invalid admin refresh token")
	}

	return response.Single(c, fiber.StatusOK, fiber.Map{
		"access_token":  tokens.AccessToken,
		"refresh_token": tokens.RefreshToken,
		"role":          tokens.Role,
	})
}
