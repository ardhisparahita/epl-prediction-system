package handler

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ardhisparahita/epl-prediction-system/backend/internal/dto"
	"github.com/ardhisparahita/epl-prediction-system/backend/internal/service"
	"github.com/gofiber/fiber/v3"
)

type mockMatchService struct {
	request  dto.MatchListInput
	response dto.MatchListResponse
	err      error
	called   bool
}

func (m *mockMatchService) List(
	ctx context.Context,
	input dto.MatchListInput,
) (dto.MatchListResponse, error) {
	m.called = true
	m.request = input

	return m.response, m.err
}

func newMatchTestApp(matchService MatchService) *fiber.App {
	app := fiber.New()

	handler := NewMatchHandler(matchService)
	app.Get("/api/v1/matches", handler.List)

	return app
}

func TestMatchHandlerSuccess(t *testing.T) {
	mockService := &mockMatchService{
		response: dto.MatchListResponse{
			Data: []dto.MatchListItem{
				{
					ID:        6253,
					Date:      "2026-11-07",
					HomeTeam:  "Arsenal",
					AwayTeam:  "Hull",
					HomeGoals: 1,
					AwayGoals: 0,
				},
			},
			Pagination: dto.MatchPagination{
				Page:       1,
				Limit:      20,
				Total:      314,
				TotalPages: 16,
			},
		},
	}

	app := newMatchTestApp(mockService)

	req := httptest.NewRequest(
		http.MethodGet,
		"/api/v1/matches",
		nil,
	)

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected HTTP 200, got %d", resp.StatusCode)
	}

	var result dto.MatchListResponse

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if len(result.Data) != 1 {
		t.Fatalf("expected 1 match, got %d", len(result.Data))
	}

	if result.Data[0].HomeTeam != "Arsenal" {
		t.Errorf("expected Arsenal, got %s", result.Data[0].HomeTeam)
	}

	if result.Pagination.Total != 314 {
		t.Errorf("expected total 314, got %d", result.Pagination.Total)
	}

	if !mockService.called {
		t.Error("match service was not called")
	}
}

func TestMatchHandlerForwardsQueryParameters(t *testing.T) {
	mockService := &mockMatchService{
		response: dto.MatchListResponse{
			Data: []dto.MatchListItem{},
		},
	}

	app := newMatchTestApp(mockService)

	req := httptest.NewRequest(
		http.MethodGet,
		"/api/v1/matches?team=Arsenal&from=2026-08-01&to=2026-11-08&page=2&limit=10",
		nil,
	)

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected HTTP 200, got %d", resp.StatusCode)
	}

	if !mockService.called {
		t.Fatal("match service was not called")
	}

	if mockService.request.Team != "Arsenal" {
		t.Errorf("expected team Arsenal, got %s", mockService.request.Team)
	}

	if mockService.request.From != "2026-08-01" {
		t.Errorf("unexpected from date: %s", mockService.request.From)
	}

	if mockService.request.To != "2026-11-08" {
		t.Errorf("unexpected to date: %s", mockService.request.To)
	}

	if mockService.request.Page != 2 {
		t.Errorf("expected page 2, got %d", mockService.request.Page)
	}

	if mockService.request.Limit != 10 {
		t.Errorf("expected limit 10, got %d", mockService.request.Limit)
	}
}

func TestMatchHandlerInvalidIntegerQuery(t *testing.T) {
	tests := []struct {
		name string
		path string
	}{
		{
			name: "invalid page",
			path: "/api/v1/matches?page=abc",
		},
		{
			name: "invalid limit",
			path: "/api/v1/matches?limit=abc",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := &mockMatchService{}
			app := newMatchTestApp(mockService)

			req := httptest.NewRequest(http.MethodGet, tt.path, nil)

			resp, err := app.Test(req)
			if err != nil {
				t.Fatalf("request failed: %v", err)
			}
			defer resp.Body.Close()

			if resp.StatusCode != http.StatusBadRequest {
				t.Fatalf("expected HTTP 400, got %d", resp.StatusCode)
			}

			if mockService.called {
				t.Error("match service should not be called")
			}
		})
	}
}

func TestMatchHandlerValidationError(t *testing.T) {
	mockService := &mockMatchService{
		err: &service.ValidationError{
			Message: "limit must be between 1 and 100",
		},
	}

	app := newMatchTestApp(mockService)

	req := httptest.NewRequest(
		http.MethodGet,
		"/api/v1/matches?limit=101",
		nil,
	)

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
		t.Fatalf("failed to decode response: %v", err)
	}

	if result["message"] != "limit must be between 1 and 100" {
		t.Errorf("unexpected error message: %s", result["message"])
	}

	if !mockService.called {
		t.Error("match service was not called")
	}
}

func TestMatchHandlerRepositoryError(t *testing.T) {
	mockService := &mockMatchService{
		err: &service.RepositoryError{
			Err: errors.New("database connection failed"),
		},
	}

	app := newMatchTestApp(mockService)

	req := httptest.NewRequest(
		http.MethodGet,
		"/api/v1/matches",
		nil,
	)

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("expected HTTP 500, got %d", resp.StatusCode)
	}

	var result map[string]string

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if result["message"] != "Failed to retrieve match history" {
		t.Errorf("unexpected error message: %s", result["message"])
	}
}

func TestMatchHandlerServiceNotInitialized(t *testing.T) {
	app := newMatchTestApp(nil)

	req := httptest.NewRequest(
		http.MethodGet,
		"/api/v1/matches",
		nil,
	)

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusServiceUnavailable {
		t.Fatalf("expected HTTP 503, got %d", resp.StatusCode)
	}
}
