package handler

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/ardhisparahita/epl-prediction-system/backend/internal/client"
	"github.com/ardhisparahita/epl-prediction-system/backend/internal/dto"
	"github.com/gofiber/fiber/v3"
)

type mockPredictionService struct {
	response dto.PredictionResponse
	err      error
	called   bool
	request  dto.PredictionRequest
}

func (m *mockPredictionService) Predict(
	ctx context.Context,
	payload dto.PredictionRequest,
) (dto.PredictionResponse, error) {
	m.called = true
	m.request = payload

	return m.response, m.err
}

func newPredictionTestApp(
	predictionService PredictionService,
) *fiber.App {
	app := fiber.New()

	handler := NewPredictionHandler(predictionService)

	app.Post("/api/v1/predictions", handler.Predict)

	return app
}

func TestPredictionHandlerSuccess(t *testing.T) {
	mockService := &mockPredictionService{
		response: dto.PredictionResponse{
			HomeTeam:              "Arsenal",
			AwayTeam:              "Chelsea",
			KickoffDate:           "2026-11-15",
			FeatureCount:          12,
			HistoryWindow:         5,
			MinimumHistoryMatches: 3,
			Probabilities: dto.PredictionProbabilities{
				LogisticRegression: dto.ClassProbabilities{
					H: 0.408890,
					D: 0.334658,
					A: 0.256452,
				},
				RandomForest: dto.ClassProbabilities{
					H: 0.386105,
					D: 0.313836,
					A: 0.300059,
				},
			},
			ModelVersion: "epl-result-live-2026-2027-v1",
		},
	}

	app := newPredictionTestApp(mockService)

	body := `{
		"home_team": "Arsenal",
		"away_team": "Chelsea",
		"kickoff_date": "2026-11-15",
		"historical_matches": []
	}`

	req := httptest.NewRequest(
		http.MethodPost,
		"/api/v1/predictions",
		strings.NewReader(body),
	)
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected HTTP 200, got %d", resp.StatusCode)
	}

	var result dto.PredictionResponse

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if result.HomeTeam != "Arsenal" {
		t.Errorf("expected Arsenal, got %s", result.HomeTeam)
	}

	if result.AwayTeam != "Chelsea" {
		t.Errorf("expected Chelsea, got %s", result.AwayTeam)
	}

	if result.FeatureCount != 12 {
		t.Errorf("expected 12 features, got %d", result.FeatureCount)
	}

	if result.Probabilities.LogisticRegression.H != 0.408890 {
		t.Errorf("unexpected LR probability: %f",
			result.Probabilities.LogisticRegression.H)
	}

	if result.ModelVersion != "epl-result-live-2026-2027-v1" {
		t.Errorf("unexpected model version: %s", result.ModelVersion)
	}

	if !mockService.called {
		t.Error("prediction service was not called")
	}

	if mockService.request.HomeTeam != "Arsenal" {
		t.Errorf("home team was not forwarded correctly")
	}
}

func TestPredictionHandlerInvalidJSON(t *testing.T) {
	mockService := &mockPredictionService{}
	app := newPredictionTestApp(mockService)

	req := httptest.NewRequest(
		http.MethodPost,
		"/api/v1/predictions",
		strings.NewReader(`{"home_team":`),
	)
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected HTTP 400, got %d", resp.StatusCode)
	}

	if mockService.called {
		t.Error("prediction service should not be called for invalid JSON")
	}
}

func TestPredictionHandlerUpstreamValidationError(t *testing.T) {
	mockService := &mockPredictionService{
		err: &client.HTTPError{
			StatusCode: http.StatusUnprocessableEntity,
			Body:       "Riwayat Arsenal hanya 2 pertandingan",
		},
	}

	app := newPredictionTestApp(mockService)

	body := `{
		"home_team": "Arsenal",
		"away_team": "Chelsea",
		"kickoff_date": "2026-11-15",
		"historical_matches": []
	}`

	req := httptest.NewRequest(
		http.MethodPost,
		"/api/v1/predictions",
		strings.NewReader(body),
	)
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("expected HTTP 422, got %d", resp.StatusCode)
	}

	var result map[string]string

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		t.Fatalf("failed to decode error response: %v", err)
	}

	if result["message"] != mockService.err.Error() &&
		result["message"] != "Riwayat Arsenal hanya 2 pertandingan" {
		t.Errorf("unexpected error message: %s", result["message"])
	}
}

func TestPredictionHandlerMLServiceUnavailable(t *testing.T) {
	mockService := &mockPredictionService{
		err: errors.New("connection refused"),
	}

	app := newPredictionTestApp(mockService)

	body := `{
		"home_team": "Arsenal",
		"away_team": "Chelsea",
		"kickoff_date": "2026-11-15",
		"historical_matches": []
	}`

	req := httptest.NewRequest(
		http.MethodPost,
		"/api/v1/predictions",
		strings.NewReader(body),
	)
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadGateway {
		t.Fatalf("expected HTTP 502, got %d", resp.StatusCode)
	}

	var result map[string]string

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		t.Fatalf("failed to decode error response: %v", err)
	}

	if result["message"] != "ML prediction service is unavailable" {
		t.Errorf("unexpected error message: %s", result["message"])
	}
}

func TestPredictionHandlerServiceNotInitialized(t *testing.T) {
	app := newPredictionTestApp(nil)

	req := httptest.NewRequest(
		http.MethodPost,
		"/api/v1/predictions",
		strings.NewReader(`{}`),
	)
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusServiceUnavailable {
		t.Fatalf("expected HTTP 503, got %d", resp.StatusCode)
	}
}
