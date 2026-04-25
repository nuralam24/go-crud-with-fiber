package handler

import (
	"strings"

	"github.com/gofiber/fiber/v2"

	"github.com/storex/go-crud/internal/service"
)

type AuthHandler struct {
	authService *service.AuthService
}

type loginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func NewAuthHandler(authService *service.AuthService) *AuthHandler {
	return &AuthHandler{authService: authService}
}

func (h *AuthHandler) Login(c *fiber.Ctx) error {
	var req loginRequest
	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid payload")
	}
	req.Email = strings.TrimSpace(req.Email)
	req.Password = strings.TrimSpace(req.Password)
	if req.Email == "" || req.Password == "" {
		return fiber.NewError(fiber.StatusBadRequest, "email and password are required")
	}

	token, role, err := h.authService.Login(req.Email, req.Password)
	if err != nil {
		return fiber.NewError(fiber.StatusUnauthorized, "invalid credentials")
	}

	return c.JSON(fiber.Map{
		"access_token": token,
		"role":         role,
	})
}
