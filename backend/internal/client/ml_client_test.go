package client

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/ardhisparahita/epl-prediction-system/backend/internal/dto"
)

func float64Ptr(value float64) *float64 {
	return &value
}

func TestPredictSuccess(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			if r.Method != http.MethodPost {
				t.Errorf("expected POST, got %s", r.Method)
			}

			if r.URL.Path != predictionEndpoint {
				t.Errorf(
					"expected path %s, got %s",
					predictionEndpoint,
					r.URL.Path,
				)
			}

			if r.Header.Get("Content-Type") != "application/json" {
				t.Errorf("unexpected Content-Type: %s", r.Header.Get("Content-Type"))
			}

			if r.Header.Get("Accept") != "application/json" {
				t.Errorf("unexpected Accept: %s", r.Header.Get("Accept"))
			}

			var received dto.PredictionRequest

			if err := json.NewDecoder(r.Body).Decode(&received); err != nil {
				t.Errorf("failed to decode request: %v", err)
				http.Error(w, "invalid request", http.StatusBadRequest)
				return
			}

			if received.HomeTeam != "Arsenal" {
				t.Errorf("expected Arsenal, got %s", received.HomeTeam)
			}

			if received.AwayTeam != "Chelsea" {
				t.Errorf("expected Chelsea, got %s", received.AwayTeam)
			}

			if received.KickoffDate != "2026-11-15" {
				t.Errorf("unexpected kickoff date: %s", received.KickoffDate)
			}

			if len(received.HistoricalMatches) != 1 {
				t.Errorf(
					"expected 1 historical match, got %d",
					len(received.HistoricalMatches),
				)
			}

			match := received.HistoricalMatches[0]

			if match.HomeShotsOnTarget == nil ||
				*match.HomeShotsOnTarget != 5 {
				t.Errorf("home shots on target was not serialized correctly")
			}

			if match.AwayShotsOnTarget != nil {
				t.Errorf("expected nil away shots on target")
			}

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)

			_, _ = fmt.Fprint(w, `{
				"home_team": "Arsenal",
				"away_team": "Chelsea",
				"kickoff_date": "2026-11-15",
				"feature_count": 12,
				"history_window": 5,
				"minimum_history_matches": 3,
				"probabilities": {
					"logistic_regression": {
						"H": 0.40889,
						"D": 0.334658,
						"A": 0.256452
					},
					"random_forest": {
						"H": 0.386105,
						"D": 0.313836,
						"A": 0.300059
					}
				},
				"model_version": "epl-result-live-2026-2027-v1"
			}`)
		},
	))
	defer server.Close()

	mlClient, err := NewMLClient(
		server.URL+"/",
		5*time.Second,
	)
	if err != nil {
		t.Fatalf("failed to create ML client: %v", err)
	}

	payload := dto.PredictionRequest{
		HomeTeam:    "Arsenal",
		AwayTeam:    "Chelsea",
		KickoffDate: "2026-11-15",
		HistoricalMatches: []dto.HistoricalMatch{
			{
				Date:              "2026-10-01",
				HomeTeam:          "Arsenal",
				AwayTeam:          "Opponent",
				HomeGoals:         2,
				AwayGoals:         0,
				HomeShotsOnTarget: float64Ptr(5),
				AwayShotsOnTarget: nil,
			},
		},
	}

	result, err := mlClient.Predict(
		context.Background(),
		payload,
	)
	if err != nil {
		t.Fatalf("prediction request failed: %v", err)
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

	if result.Probabilities.LogisticRegression.H != 0.40889 {
		t.Errorf(
			"unexpected LR home probability: %f",
			result.Probabilities.LogisticRegression.H,
		)
	}

	if result.Probabilities.RandomForest.A != 0.300059 {
		t.Errorf(
			"unexpected RF away probability: %f",
			result.Probabilities.RandomForest.A,
		)
	}

	if result.ModelVersion != "epl-result-live-2026-2027-v1" {
		t.Errorf("unexpected model version: %s", result.ModelVersion)
	}
}

func TestPredictHandlesHTTPError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusUnprocessableEntity)

			_, _ = fmt.Fprint(
				w,
				`{"detail":"Riwayat Arsenal hanya 2 pertandingan"}`,
			)
		},
	))
	defer server.Close()

	mlClient, err := NewMLClient(server.URL, 5*time.Second)
	if err != nil {
		t.Fatalf("failed to create ML client: %v", err)
	}

	result, err := mlClient.Predict(
		context.Background(),
		dto.PredictionRequest{},
	)

	if err == nil {
		t.Fatal("expected HTTP error, got nil")
	}

	if result.ModelVersion != "" {
		t.Error("expected empty result on HTTP error")
	}

	var httpErr *HTTPError

	if !errors.As(err, &httpErr) {
		t.Fatalf("expected HTTPError, got %T", err)
	}

	if httpErr.StatusCode != http.StatusUnprocessableEntity {
		t.Errorf(
			"expected HTTP 422, got %d",
			httpErr.StatusCode,
		)
	}

	if !strings.Contains(
		httpErr.Body,
		"Riwayat Arsenal hanya 2 pertandingan",
	) {
		t.Errorf("unexpected error body: %s", httpErr.Body)
	}
}

func TestNewMLClientRejectsInvalidConfiguration(t *testing.T) {
	tests := []struct {
		name        string
		baseURL     string
		timeout     time.Duration
		expectedErr string
	}{
		{
			name:        "empty URL",
			baseURL:     "",
			timeout:     time.Second,
			expectedErr: "ML service URL is required",
		},
		{
			name:        "invalid scheme",
			baseURL:     "ftp://localhost:8000",
			timeout:     time.Second,
			expectedErr: "must use HTTP or HTTPS",
		},
		{
			name:        "missing host",
			baseURL:     "http:///predictions",
			timeout:     time.Second,
			expectedErr: "must contain a host",
		},
		{
			name:        "URL contains query",
			baseURL:     "http://localhost:8000?key=value",
			timeout:     time.Second,
			expectedErr: "must not contain query parameters",
		},
		{
			name:        "URL contains fragment",
			baseURL:     "http://localhost:8000#prediction",
			timeout:     time.Second,
			expectedErr: "must not contain query parameters",
		},
		{
			name:        "zero timeout",
			baseURL:     "http://localhost:8000",
			timeout:     0,
			expectedErr: "timeout must be positive",
		},
		{
			name:        "negative timeout",
			baseURL:     "http://localhost:8000",
			timeout:     -time.Second,
			expectedErr: "timeout must be positive",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewMLClient(tt.baseURL, tt.timeout)

			if err == nil {
				t.Fatal("expected error, got nil")
			}

			if !strings.Contains(err.Error(), tt.expectedErr) {
				t.Errorf(
					"expected error containing %q, got %q",
					tt.expectedErr,
					err.Error(),
				)
			}
		})
	}
}

func TestPredictHandlesTimeout(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			// Simulate a slow ML Service response.
			time.Sleep(200 * time.Millisecond)

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)

			_, _ = fmt.Fprint(w, `{}`)
		},
	))
	defer server.Close()

	mlClient, err := NewMLClient(
		server.URL,
		20*time.Millisecond,
	)
	if err != nil {
		t.Fatalf("failed to create ML client: %v", err)
	}

	_, err = mlClient.Predict(
		context.Background(),
		dto.PredictionRequest{},
	)

	if err == nil {
		t.Fatal("expected timeout error, got nil")
	}

	if !strings.Contains(err.Error(), "failed to communicate") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestPredictRejectsInvalidJSONResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			_, _ = fmt.Fprint(w, `{"home_team":`)
		},
	))
	defer server.Close()

	mlClient, err := NewMLClient(server.URL, 5*time.Second)
	if err != nil {
		t.Fatalf("failed to create ML client: %v", err)
	}

	_, err = mlClient.Predict(
		context.Background(),
		dto.PredictionRequest{},
	)

	if err == nil {
		t.Fatal("expected JSON decoding error, got nil")
	}

	if !strings.Contains(err.Error(), "failed to decode") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestPredictRejectsOversizedResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			_, _ = fmt.Fprint(
				w,
				strings.Repeat("a", maxResponseSize+1),
			)
		},
	))
	defer server.Close()

	mlClient, err := NewMLClient(server.URL, 5*time.Second)
	if err != nil {
		t.Fatalf("failed to create ML client: %v", err)
	}

	_, err = mlClient.Predict(
		context.Background(),
		dto.PredictionRequest{},
	)

	if err == nil {
		t.Fatal("expected oversized response error, got nil")
	}

	if !strings.Contains(err.Error(), "exceeds maximum allowed size") {
		t.Errorf("unexpected error: %v", err)
	}
}
