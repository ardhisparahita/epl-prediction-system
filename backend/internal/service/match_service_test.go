package service

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/ardhisparahita/epl-prediction-system/backend/internal/dto"
	"github.com/ardhisparahita/epl-prediction-system/backend/internal/models"
	"github.com/ardhisparahita/epl-prediction-system/backend/internal/repository"
)

type matchServiceTestRepository struct {
	matches []models.Match
	total   int64
	err     error

	filters repository.MatchFilters
	limit   int
	offset  int
	called  bool
}

func (r *matchServiceTestRepository) FindMatches(
	_ context.Context,
	filters repository.MatchFilters,
	limit int,
	offset int,
) ([]models.Match, int64, error) {
	r.called = true
	r.filters = filters
	r.limit = limit
	r.offset = offset

	return r.matches, r.total, r.err
}

func TestMatchServiceListDefaultsAndMapsMatches(t *testing.T) {
	date, _ := time.Parse(time.DateOnly, "2026-11-08")

	repo := &matchServiceTestRepository{
		matches: []models.Match{
			{
				ID:       3140,
				Date:     date,
				HomeTeam: "Arsenal",
				AwayTeam: "Chelsea",
				FTHG:     2,
				FTAG:     1,
			},
		},
		total: 21,
	}

	svc, err := NewMatchService(repo)
	if err != nil {
		t.Fatalf("failed to create service: %v", err)
	}

	result, err := svc.List(context.Background(), dto.MatchListInput{})
	if err != nil {
		t.Fatalf("failed to list matches: %v", err)
	}

	if repo.limit != 20 || repo.offset != 0 {
		t.Errorf(
			"expected limit 20 and offset 0, got %d and %d",
			repo.limit,
			repo.offset,
		)
	}

	if result.Pagination.Page != 1 {
		t.Errorf("expected page 1, got %d", result.Pagination.Page)
	}

	if result.Pagination.TotalPages != 2 {
		t.Errorf("expected 2 total pages, got %d", result.Pagination.TotalPages)
	}

	if len(result.Data) != 1 {
		t.Fatalf("expected 1 match, got %d", len(result.Data))
	}

	match := result.Data[0]

	if match.ID != 3140 ||
		match.HomeTeam != "Arsenal" ||
		match.AwayTeam != "Chelsea" ||
		match.Date != "2026-11-08" {
		t.Errorf("unexpected match data: %+v", match)
	}
}

func TestMatchServiceListAppliesFiltersAndPagination(t *testing.T) {
	repo := &matchServiceTestRepository{
		total: 35,
	}

	svc, _ := NewMatchService(repo)

	result, err := svc.List(
		context.Background(),
		dto.MatchListInput{
			Team:  " Arsenal ",
			From:  "2026-01-01",
			To:    "2026-06-30",
			Page:  2,
			Limit: 10,
		},
	)
	if err != nil {
		t.Fatalf("failed to list matches: %v", err)
	}

	if repo.filters.Team != "Arsenal" {
		t.Errorf("expected team Arsenal, got %s", repo.filters.Team)
	}

	if repo.limit != 10 || repo.offset != 10 {
		t.Errorf(
			"expected limit 10 and offset 10, got %d and %d",
			repo.limit,
			repo.offset,
		)
	}

	if repo.filters.From == nil || repo.filters.To == nil {
		t.Fatal("expected date filters to be set")
	}

	if result.Pagination.Page != 2 ||
		result.Pagination.TotalPages != 4 {
		t.Errorf("unexpected pagination: %+v", result.Pagination)
	}
}

func TestMatchServiceListRejectsInvalidInput(t *testing.T) {
	tests := []struct {
		name  string
		input dto.MatchListInput
	}{
		{
			name:  "negative page",
			input: dto.MatchListInput{Page: -1},
		},
		{
			name:  "limit exceeds maximum",
			input: dto.MatchListInput{Limit: 101},
		},
		{
			name:  "invalid from date",
			input: dto.MatchListInput{From: "01-01-2026"},
		},
		{
			name:  "invalid to date",
			input: dto.MatchListInput{To: "invalid"},
		},
		{
			name: "reversed date range",
			input: dto.MatchListInput{
				From: "2026-06-01",
				To:   "2026-01-01",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &matchServiceTestRepository{}
			svc, _ := NewMatchService(repo)

			_, err := svc.List(context.Background(), tt.input)

			if err == nil {
				t.Fatal("expected validation error, got nil")
			}

			if repo.called {
				t.Error("repository should not be called for invalid input")
			}

			var validationErr *ValidationError
			if !errors.As(err, &validationErr) {
				t.Errorf("expected ValidationError, got %T", err)
			}
		})
	}
}

func TestMatchServiceListWrapsRepositoryError(t *testing.T) {
	repo := &matchServiceTestRepository{
		err: errors.New("database connection failed"),
	}

	svc, _ := NewMatchService(repo)

	_, err := svc.List(context.Background(), dto.MatchListInput{})
	if err == nil {
		t.Fatal("expected repository error, got nil")
	}

	var repositoryErr *RepositoryError
	if !errors.As(err, &repositoryErr) {
		t.Fatalf("expected RepositoryError, got %T", err)
	}

	if !strings.Contains(err.Error(), "database connection failed") {
		t.Errorf("unexpected error message: %v", err)
	}
}
