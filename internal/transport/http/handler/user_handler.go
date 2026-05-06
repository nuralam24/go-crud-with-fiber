package handler

import (
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"

	"github.com/storex/go-crud/internal/domain"
	"github.com/storex/go-crud/internal/service"
	"github.com/storex/go-crud/internal/transport/apierror"
	"github.com/storex/go-crud/internal/transport/http/middleware"
	"github.com/storex/go-crud/internal/transport/response"
)

type UserHandler struct {
	userService *service.UserService
	authService *service.AuthService
}

type userLoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type userRegisterRequest struct {
	Name     string `json:"name"`
	Phone    string `json:"phone"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

type userRefreshRequest struct {
	RefreshToken string `json:"refresh_token"`
}

type updateProfileRequest struct {
	Name  string `json:"name"`
	Phone string `json:"phone"`
}

func NewUserHandler(userService *service.UserService, authService *service.AuthService) *UserHandler {
	return &UserHandler{
		userService: userService,
		authService: authService,
	}
}

func (h *UserHandler) Login(c *fiber.Ctx) error {
	var req userLoginRequest
	if err := c.BodyParser(&req); err != nil {
		return apierror.New(fiber.StatusBadRequest, "invalid_payload", "invalid payload")
	}
	req.Email = strings.TrimSpace(req.Email)
	req.Password = strings.TrimSpace(req.Password)
	if req.Email == "" || req.Password == "" {
		return apierror.New(fiber.StatusBadRequest, "validation_error", "email and password are required")
	}
	if strings.EqualFold(req.Email, service.AdminEmail) {
		return apierror.New(fiber.StatusUnauthorized, "invalid_credentials", "admin login is not allowed in user module")
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

func (h *UserHandler) Register(c *fiber.Ctx) error {
	var req userRegisterRequest
	if err := c.BodyParser(&req); err != nil {
		return apierror.New(fiber.StatusBadRequest, "invalid_payload", "invalid payload")
	}
	if strings.EqualFold(strings.TrimSpace(req.Email), service.AdminEmail) {
		return apierror.New(fiber.StatusBadRequest, "validation_error", "admin email is reserved")
	}

	user, err := h.userService.Register(c.Context(), req.Name, req.Phone, req.Email, req.Password)
	if err != nil {
		if err == domain.ErrInvalidUser {
			return apierror.New(fiber.StatusBadRequest, "validation_error", "invalid user data")
		}
		if err == service.ErrUserExists {
			return apierror.New(fiber.StatusConflict, "email_exists", "email already exists")
		}
		return err
	}
	return response.Single(c, fiber.StatusCreated, user)
}

func (h *UserHandler) Refresh(c *fiber.Ctx) error {
	var req userRefreshRequest
	if err := c.BodyParser(&req); err != nil {
		return apierror.New(fiber.StatusBadRequest, "invalid_payload", "invalid payload")
	}

	tokens, err := h.authService.Refresh(c.Context(), req.RefreshToken)
	if err != nil {
		return apierror.New(fiber.StatusUnauthorized, "invalid_refresh_token", "invalid refresh token")
	}
	if tokens.Role != service.RoleUser {
		return apierror.New(fiber.StatusUnauthorized, "invalid_refresh_token", "invalid user refresh token")
	}

	return response.Single(c, fiber.StatusOK, fiber.Map{
		"access_token":  tokens.AccessToken,
		"refresh_token": tokens.RefreshToken,
		"role":          tokens.Role,
	})
}

func (h *UserHandler) GetMyProfile(c *fiber.Ctx) error {
	claims := middleware.ClaimsFromCtx(c)
	if claims == nil || claims.UserID == "" {
		return fiber.NewError(fiber.StatusForbidden, "access denied")
	}
	userID, err := uuid.Parse(claims.UserID)
	if err != nil {
		return fiber.NewError(fiber.StatusUnauthorized, "invalid token")
	}

	user, err := h.userService.GetMyProfile(c.Context(), userID)
	if err != nil {
		if err == service.ErrNotFound {
			return fiber.NewError(fiber.StatusNotFound, "user not found")
		}
		return err
	}
	return response.Single(c, fiber.StatusOK, user)
}

func (h *UserHandler) UpdateMyProfile(c *fiber.Ctx) error {
	claims := middleware.ClaimsFromCtx(c)
	if claims == nil || claims.UserID == "" {
		return fiber.NewError(fiber.StatusForbidden, "access denied")
	}
	userID, err := uuid.Parse(claims.UserID)
	if err != nil {
		return fiber.NewError(fiber.StatusUnauthorized, "invalid token")
	}

	var req updateProfileRequest
	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid payload")
	}

	user, err := h.userService.UpdateMyProfile(c.Context(), userID, req.Name, req.Phone)
	if err != nil {
		if err == domain.ErrInvalidUser {
			return fiber.NewError(fiber.StatusBadRequest, "invalid user data")
		}
		if err == service.ErrNotFound {
			return fiber.NewError(fiber.StatusNotFound, "user not found")
		}
		return err
	}

	return response.Single(c, fiber.StatusOK, user)
}
