package http

import (
	"context"
	"fmt"
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/shngxx/point/internal/usecase"
)

// GetPointService defines the interface for getting point information
type GetPointService interface {
	GetPoint(ctx context.Context, id int) (*usecase.PointInfo, error)
}

// NewGetPointHandler creates a handler for getting point information
func NewGetPointHandler(service GetPointService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		ctx := c.UserContext()
		if ctx == nil {
			ctx = context.Background()
		}

		id := c.Params("id")
		if id == "" {
			id = "1"
		}

		pointID, err := strconv.Atoi(id)
		if err != nil || pointID <= 0 {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": fmt.Sprintf("Invalid point ID: %s", id),
			})
		}

		pointInfo, err := service.GetPoint(ctx, pointID)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": fmt.Sprintf("Error getting point information: %v", err),
			})
		}

		return c.JSON(pointInfo)
	}
}
