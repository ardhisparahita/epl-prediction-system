package handler

import "github.com/gofiber/fiber/v3"

func HealthCheck(c fiber.Ctx) error {
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"status":  "success",
		"message": "EPL Prediction API is running",
	})
}