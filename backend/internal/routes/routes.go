package routes

import (
	"github.com/ardhisparahita/epl-prediction-system/backend/internal/handler"

	"github.com/gofiber/fiber/v3"
	"gorm.io/gorm"
)

func SetupRoutes(app *fiber.App, db *gorm.DB) {
	api := app.Group("/api/v1")

	api.Get("/health", handler.HealthCheck)
	api.Get("/health/db", handler.DatabaseHealth(db))
}
