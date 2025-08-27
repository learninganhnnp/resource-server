package middleware

import (
	"log"

	"github.com/gofiber/fiber/v2"
)

type ErrorResponse struct {
	Success bool        `json:"success"`
	Data    any         `json:"data"`
	Message string      `json:"message"`
	Error   *ErrorInfo  `json:"error,omitempty"`
}

type ErrorInfo struct {
	Code    string `json:"code"`
	Details string `json:"details,omitempty"`
}

func ErrorHandler(c *fiber.Ctx, err error) error {
	code := fiber.StatusInternalServerError
	message := "Internal Server Error"
	errorCode := "INTERNAL_ERROR"
	details := ""

	if e, ok := err.(*fiber.Error); ok {
		code = e.Code
		message = e.Message
		
		switch code {
		case fiber.StatusBadRequest:
			errorCode = "BAD_REQUEST"
		case fiber.StatusNotFound:
			errorCode = "NOT_FOUND"
		case fiber.StatusUnauthorized:
			errorCode = "UNAUTHORIZED"
		case fiber.StatusForbidden:
			errorCode = "FORBIDDEN"
		case fiber.StatusConflict:
			errorCode = "CONFLICT"
		case fiber.StatusUnprocessableEntity:
			errorCode = "VALIDATION_ERROR"
		default:
			errorCode = "HTTP_ERROR"
		}
		details = e.Message
	}

	log.Printf("Error: %v", err)

	response := ErrorResponse{
		Success: false,
		Data:    nil,
		Message: message,
		Error: &ErrorInfo{
			Code:    errorCode,
			Details: details,
		},
	}

	return c.Status(code).JSON(response)
}