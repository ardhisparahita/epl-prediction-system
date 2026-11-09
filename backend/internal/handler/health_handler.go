package handler

import (
	"context"
	"time"

	"github.com/gofiber/fiber/v3"
	"gorm.io/gorm"
)

func HealthCheck(c fiber.Ctx) error {
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"status":  "success",
		"message": "EPL Prediction API is running",
	})
}

func DatabaseHealth(db *gorm.DB) fiber.Handler {
	return func(c fiber.Ctx) error {
		sqlDB, err := db.DB()
		if err != nil {
			return c.Status(
				fiber.StatusServiceUnavailable,
			).JSON(fiber.Map{
				"status":   "error",
				"database": "unavailable",
			})
		}

		ctx, cancel := context.WithTimeout(
			context.Background(),
			3*time.Second,
		)
		defer cancel()

		if err := sqlDB.PingContext(ctx); err != nil {
			return c.Status(
				fiber.StatusServiceUnavailable,
			).JSON(fiber.Map{
				"status":   "error",
				"database": "unavailable",
			})
		}

		return c.Status(fiber.StatusOK).JSON(fiber.Map{
			"status":   "success",
			"database": "connected",
		})
	}
}