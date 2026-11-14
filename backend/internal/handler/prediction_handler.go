package handler

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/ardhisparahita/epl-prediction-system/backend/internal/client"
	"github.com/ardhisparahita/epl-prediction-system/backend/internal/dto"
	"github.com/gofiber/fiber/v3"
)

// PredictionService defines the contract required by the handler.
type PredictionService interface {
	Predict(
		ctx context.Context,
		payload dto.PredictionRequest,
	) (dto.PredictionResponse, error)
}

// PredictionHandler handles match prediction requests.
type PredictionHandler struct {
	predictionService PredictionService
}

// NewPredictionHandler creates a new prediction handler.
func NewPredictionHandler(
	predictionService PredictionService,
) *PredictionHandler {
	return &PredictionHandler{
		predictionService: predictionService,
	}
}

// Predict handles POST /api/v1/predictions.
func (h *PredictionHandler) Predict(c fiber.Ctx) error {
	if h == nil || h.predictionService == nil {
		return c.Status(fiber.StatusServiceUnavailable).JSON(
			fiber.Map{
				"status":  "error",
				"message": "Prediction service is not initialized",
			},
		)
	}

	var payload dto.PredictionRequest

	if err := json.Unmarshal(c.Body(), &payload); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(
			fiber.Map{
				"status":  "error",
				"message": "Invalid JSON request body",
			},
		)
	}

	// Fiber v3 Ctx implements context.Context.
	result, err := h.predictionService.Predict(c, payload)

	if err != nil {
		var upstreamErr *client.HTTPError

		if errors.As(err, &upstreamErr) {
			if upstreamErr.StatusCode == fiber.StatusUnprocessableEntity {
				return c.Status(
					fiber.StatusUnprocessableEntity,
				).JSON(fiber.Map{
					"status":  "error",
					"message": upstreamErr.Body,
				})
			}
		}

		return c.Status(fiber.StatusBadGateway).JSON(
			fiber.Map{
				"status":  "error",
				"message": "ML prediction service is unavailable",
			},
		)
	}

	return c.Status(fiber.StatusOK).JSON(result)
}
