package service

import (
	"context"
	"errors"
	"testing"

	"github.com/ardhisparahita/epl-prediction-system/backend/internal/dto"
)

type mockPredictionClient struct {
	request  dto.PredictionRequest
	response dto.PredictionResponse
	err      error
	calls    int
}

func (m *mockPredictionClient) Predict(
	ctx context.Context,
	payload dto.PredictionRequest,
) (dto.PredictionResponse, error) {
	m.calls++
	m.request = payload

	return m.response, m.err
}

func TestNewPredictionServiceRejectsNilClient(t *testing.T) {
	_, err := NewPredictionService(nil)

	if err == nil {
		t.Fatal("expected error for nil ML client")
	}
}

func TestPredictForwardsRequestAndReturnsResponse(t *testing.T) {
	mockClient := &mockPredictionClient{
		response: dto.PredictionResponse{
			HomeTeam:     "Arsenal",
			AwayTeam:     "Chelsea",
			FeatureCount: 12,
			ModelVersion: "epl-result-live-2026-2027-v1",
		},
	}

	predictionService, err := NewPredictionService(mockClient)
	if err != nil {
		t.Fatalf("failed to create prediction service: %v", err)
	}

	payload := dto.PredictionRequest{
		HomeTeam:    "Arsenal",
		AwayTeam:    "Chelsea",
		KickoffDate: "2026-11-15",
		HistoricalMatches: []dto.HistoricalMatch{
			{
				Date:      "2026-10-01",
				HomeTeam:  "Arsenal",
				AwayTeam:  "Opponent",
				HomeGoals: 2,
				AwayGoals: 0,
			},
		},
	}

	result, err := predictionService.Predict(
		context.Background(),
		payload,
	)

	if err != nil {
		t.Fatalf("prediction failed: %v", err)
	}

	if mockClient.calls != 1 {
		t.Fatalf(
			"expected 1 client call, got %d",
			mockClient.calls,
		)
	}

	if mockClient.request.HomeTeam != payload.HomeTeam {
		t.Errorf("home team was not forwarded correctly")
	}

	if mockClient.request.AwayTeam != payload.AwayTeam {
		t.Errorf("away team was not forwarded correctly")
	}

	if result.HomeTeam != "Arsenal" {
		t.Errorf("expected Arsenal, got %s", result.HomeTeam)
	}

	if result.FeatureCount != 12 {
		t.Errorf("expected 12 features, got %d", result.FeatureCount)
	}

	if result.ModelVersion != "epl-result-live-2026-2027-v1" {
		t.Errorf("unexpected model version: %s", result.ModelVersion)
	}
}

func TestPredictPropagatesClientError(t *testing.T) {
	expectedErr := errors.New("ML service unavailable")

	mockClient := &mockPredictionClient{
		err: expectedErr,
	}

	predictionService, err := NewPredictionService(mockClient)
	if err != nil {
		t.Fatalf("failed to create prediction service: %v", err)
	}

	_, err = predictionService.Predict(
		context.Background(),
		dto.PredictionRequest{},
	)

	if !errors.Is(err, expectedErr) {
		t.Fatalf(
			"expected original client error, got %v",
			err,
		)
	}
}

func TestPredictRejectsUninitializedService(t *testing.T) {
	var predictionService *PredictionService

	_, err := predictionService.Predict(
		context.Background(),
		dto.PredictionRequest{},
	)

	if err == nil {
		t.Fatal("expected error for uninitialized prediction service")
	}
}
