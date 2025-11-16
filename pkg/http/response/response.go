package response

import (
	"net/http"

	"github.com/gofiber/fiber/v2"
	"github.com/shngxx/point/pkg/http/errors"
)

// OK sends a 200 OK response with data
func OK(c *fiber.Ctx, data any) error {
	return c.Status(http.StatusOK).JSON(errors.SuccessResponse{
		Success: true,
		Data:    data,
	})
}

// Created sends a 201 Created response with data
func Created(c *fiber.Ctx, data any) error {
	return c.Status(http.StatusCreated).JSON(errors.SuccessResponse{
		Success: true,
		Data:    data,
	})
}

// BadRequest sends a 400 Bad Request response
func BadRequest(c *fiber.Ctx, err error) error {
	return c.Status(http.StatusBadRequest).JSON(errors.ErrorResponse{
		Success: false,
		Error:   err.Error(),
		Code:    errors.CodeBadRequest,
	})
}

// NotFound sends a 404 Not Found response
func NotFound(c *fiber.Ctx, msg string) error {
	return c.Status(http.StatusNotFound).JSON(errors.ErrorResponse{
		Success: false,
		Error:   msg,
		Code:    errors.CodeNotFound,
	})
}

// InternalError sends a 500 Internal Server Error response
func InternalError(c *fiber.Ctx, err error) error {
	return c.Status(http.StatusInternalServerError).JSON(errors.ErrorResponse{
		Success: false,
		Error:   err.Error(),
		Code:    errors.CodeInternalError,
	})
}
