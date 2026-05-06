package handler

import (
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"

	"github.com/storex/go-crud/internal/domain"
	"github.com/storex/go-crud/internal/service"
	"github.com/storex/go-crud/internal/transport/http/middleware"
	"github.com/storex/go-crud/internal/transport/response"
)

type ItemHandler struct {
	itemService *service.ItemService
}

type createItemRequest struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	BrandID     string `json:"brand_id"`
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

	var brandID *uuid.UUID
	if req.BrandID != "" {
		parsedBrandID, err := uuid.Parse(req.BrandID)
		if err != nil {
			return fiber.NewError(fiber.StatusBadRequest, "invalid brand_id")
		}
		brandID = &parsedBrandID
	}

	item, err := h.itemService.Create(c.Context(), req.Title, req.Description, claims.Email, brandID)
	if err != nil {
		if err == domain.ErrInvalidItem {
			return fiber.NewError(fiber.StatusBadRequest, "invalid item")
		}
		return err
	}
	return response.Single(c, fiber.StatusCreated, item)
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
	return response.Single(c, fiber.StatusOK, item)
}

func (h *ItemHandler) List(c *fiber.Ctx) error {
	limit, _ := strconv.Atoi(c.Query("limit", "20"))
	offset, _ := strconv.Atoi(c.Query("offset", "0"))

	items, total, page, pageLimit, err := h.itemService.List(c.Context(), limit, offset)
	if err != nil {
		return err
	}
	return response.List(c, fiber.StatusOK, items, response.Pagination{
		Page:  page,
		Limit: pageLimit,
		Total: total,
	})
}
