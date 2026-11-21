package routes

import (
	"github.com/ardhisparahita/epl-prediction-system/backend/internal/handler"
	"github.com/ardhisparahita/epl-prediction-system/backend/internal/service"
	"github.com/gofiber/fiber/v3"
	"gorm.io/gorm"
)

func SetupRoutes(
	app *fiber.App,
	db *gorm.DB,
	predictionService *service.PredictionService,
	matchService *service.MatchService,
) {
	api := app.Group("/api/v1")

	// Health endpoints.
	api.Get("/health", handler.HealthCheck)
	api.Get("/health/db", handler.DatabaseHealth(db))

	// Prediction endpoint.
	predictionHandler := handler.NewPredictionHandler(predictionService)
	api.Post("/predictions", predictionHandler.Predict)

	// Match history endpoint.
	matchHandler := handler.NewMatchHandler(matchService)
	api.Get("/matches", matchHandler.List)
}
