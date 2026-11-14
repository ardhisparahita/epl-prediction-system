package service

import (
	"context"
	"errors"

	"github.com/ardhisparahita/epl-prediction-system/backend/internal/dto"
)

// PredictionClient defines the contract for communicating
// with the Python ML Service.
type PredictionClient interface {
	Predict(
		ctx context.Context,
		payload dto.PredictionRequest,
	) (dto.PredictionResponse, error)
}

// PredictionService handles match prediction operations.
type PredictionService struct {
	mlClient PredictionClient
}

// NewPredictionService creates a new PredictionService.
func NewPredictionService(
	mlClient PredictionClient,
) (*PredictionService, error) {
	if mlClient == nil {
		return nil, errors.New("ML client is required")
	}

	return &PredictionService{
		mlClient: mlClient,
	}, nil
}

// Predict forwards the prediction request to the ML Service.
func (s *PredictionService) Predict(
	ctx context.Context,
	payload dto.PredictionRequest,
) (dto.PredictionResponse, error) {
	var result dto.PredictionResponse

	if s == nil || s.mlClient == nil {
		return result, errors.New(
			"prediction service is not initialized",
		)
	}

	return s.mlClient.Predict(ctx, payload)
}
