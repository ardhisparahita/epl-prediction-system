package routes

import (
	"github.com/ardhisparahita/epl-prediction-system/backend/internal/handler"
	"github.com/gofiber/fiber/v3"
)

func SetupRoutes(app *fiber.App) {
	api := app.Group("/api/v1")

	api.Get("/health", handler.HealthCheck)
}
