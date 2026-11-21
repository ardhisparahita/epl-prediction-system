package handler

import (
	"context"
	"errors"
	"strconv"

	"github.com/ardhisparahita/epl-prediction-system/backend/internal/dto"
	"github.com/ardhisparahita/epl-prediction-system/backend/internal/service"
	"github.com/gofiber/fiber/v3"
)

// MatchService defines the contract required by the handler.
type MatchService interface {
	List(
		ctx context.Context,
		input dto.MatchListInput,
	) (dto.MatchListResponse, error)
}

// MatchHandler handles match history requests.
type MatchHandler struct {
	matchService MatchService
}

// NewMatchHandler creates a new match handler.
func NewMatchHandler(matchService MatchService) *MatchHandler {
	return &MatchHandler{
		matchService: matchService,
	}
}

// List handles GET /api/v1/matches.
func (h *MatchHandler) List(c fiber.Ctx) error {
	if h == nil || h.matchService == nil {
		return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{
			"status":  "error",
			"message": "Match service is not initialized",
		})
	}

	page, err := parseQueryInt(c, "page", 0)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  "error",
			"message": "page must be a valid integer",
		})
	}

	limit, err := parseQueryInt(c, "limit", 0)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  "error",
			"message": "limit must be a valid integer",
		})
	}

	input := dto.MatchListInput{
		Team:  c.Query("team"),
		From:  c.Query("from"),
		To:    c.Query("to"),
		Page:  page,
		Limit: limit,
	}

	result, err := h.matchService.List(c, input)
	if err != nil {
		var validationErr *service.ValidationError

		if errors.As(err, &validationErr) {
			return c.Status(fiber.StatusUnprocessableEntity).JSON(
				fiber.Map{
					"status":  "error",
					"message": validationErr.Error(),
				},
			)
		}

		var repositoryErr *service.RepositoryError
		if errors.As(err, &repositoryErr) {
			return c.Status(fiber.StatusInternalServerError).JSON(
				fiber.Map{
					"status":  "error",
					"message": "Failed to retrieve match history",
				},
			)
		}

		return c.Status(fiber.StatusInternalServerError).JSON(
			fiber.Map{
				"status":  "error",
				"message": "Internal server error",
			},
		)
	}

	return c.Status(fiber.StatusOK).JSON(result)
}

func parseQueryInt(c fiber.Ctx, key string, defaultValue int) (int, error) {
	value := c.Query(key)

	if value == "" {
		return defaultValue, nil
	}

	parsed, err := strconv.Atoi(value)
	if err != nil {
		return 0, err
	}

	return parsed, nil
}
