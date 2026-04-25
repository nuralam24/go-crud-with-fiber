package handler

import (
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"

	"github.com/storex/go-crud/internal/domain"
	"github.com/storex/go-crud/internal/service"
	"github.com/storex/go-crud/internal/transport/http/middleware"
)

type ItemHandler struct {
	itemService *service.ItemService
}

type createItemRequest struct {
	Title       string `json:"title"`
	Description string `json:"description"`
}

func NewItemHandler(itemService *service.ItemService) *ItemHandler {
	return &ItemHandler{itemService: itemService}
}

func (h *ItemHandler) Create(c *fiber.Ctx) error {
	var req createItemRequest
	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid payload")
	}

	claims := middleware.ClaimsFromCtx(c)
	if claims == nil {
		return fiber.NewError(fiber.StatusForbidden, "access denied")
	}

	item, err := h.itemService.Create(c.Context(), req.Title, req.Description, claims.Email)
	if err != nil {
		if err == domain.ErrInvalidItem {
			return fiber.NewError(fiber.StatusBadRequest, "invalid item")
		}
		return err
	}
	return c.Status(fiber.StatusCreated).JSON(item)
}

func (h *ItemHandler) GetByID(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid id")
	}

	item, err := h.itemService.GetByID(c.Context(), id)
	if err != nil {
		if err == service.ErrNotFound {
			return fiber.NewError(fiber.StatusNotFound, "item not found")
		}
		return err
	}
	return c.JSON(item)
}

func (h *ItemHandler) List(c *fiber.Ctx) error {
	limit, _ := strconv.Atoi(c.Query("limit", "20"))
	offset, _ := strconv.Atoi(c.Query("offset", "0"))

	items, err := h.itemService.List(c.Context(), limit, offset)
	if err != nil {
		return err
	}
	return c.JSON(fiber.Map{
		"items":  items,
		"limit":  limit,
		"offset": offset,
	})
}
