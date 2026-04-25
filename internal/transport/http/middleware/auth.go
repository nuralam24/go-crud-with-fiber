package middleware

import (
	"strings"

	"github.com/gofiber/fiber/v2"

	"github.com/storex/go-crud/internal/service"
)

const claimsContextKey = "claims"

func Authenticate(authService *service.AuthService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		authHeader := c.Get("Authorization")
		if authHeader == "" {
			return fiber.NewError(fiber.StatusUnauthorized, "missing authorization header")
		}

		token := strings.TrimPrefix(authHeader, "Bearer ")
		if token == authHeader || strings.TrimSpace(token) == "" {
			return fiber.NewError(fiber.StatusUnauthorized, "invalid authorization header")
		}

		claims, err := authService.ParseToken(token)
		if err != nil {
			return fiber.NewError(fiber.StatusUnauthorized, "invalid token")
		}
		c.Locals(claimsContextKey, claims)
		return c.Next()
	}
}

func Authorize(allowedRoles ...service.Role) fiber.Handler {
	roleSet := make(map[service.Role]struct{}, len(allowedRoles))
	for _, role := range allowedRoles {
		roleSet[role] = struct{}{}
	}

	return func(c *fiber.Ctx) error {
		claims, ok := c.Locals(claimsContextKey).(*service.Claims)
		if !ok {
			return fiber.NewError(fiber.StatusForbidden, "access denied")
		}

		if _, exists := roleSet[claims.Role]; !exists {
			return fiber.NewError(fiber.StatusForbidden, "insufficient permissions")
		}

		return c.Next()
	}
}

func ClaimsFromCtx(c *fiber.Ctx) *service.Claims {
	claims, _ := c.Locals(claimsContextKey).(*service.Claims)
	return claims
}
