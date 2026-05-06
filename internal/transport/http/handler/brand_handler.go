package handler

import (
	"github.com/gofiber/fiber/v2"

	"github.com/storex/go-crud/internal/domain"
	"github.com/storex/go-crud/internal/service"
	"github.com/storex/go-crud/internal/transport/response"
)

type BrandHandler struct {
	brandService *service.BrandService
}

type createBrandRequest struct {
	Name string `json:"name"`
	Slug string `json:"slug"`
}

func NewBrandHandler(brandService *service.BrandService) *BrandHandler {
	return &BrandHandler{brandService: brandService}
}

func (h *BrandHandler) Create(c *fiber.Ctx) error {
	var req createBrandRequest
	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid payload")
	}

	brand, err := h.brandService.Create(c.Context(), req.Name, req.Slug)
	if err != nil {
		if err == domain.ErrInvalidBrand {
			return fiber.NewError(fiber.StatusBadRequest, "invalid brand")
		}
		return err
	}
	return response.Single(c, fiber.StatusCreated, brand)
}

func (h *BrandHandler) List(c *fiber.Ctx) error {
	brands, err := h.brandService.List(c.Context())
	if err != nil {
		return err
	}
	return response.List(c, fiber.StatusOK, brands, response.Pagination{
		Page:  1,
		Limit: len(brands),
		Total: len(brands),
	})
}
