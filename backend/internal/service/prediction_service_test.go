package service

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/ardhisparahita/epl-prediction-system/backend/internal/dto"
	"github.com/ardhisparahita/epl-prediction-system/backend/internal/models"
)

// Mock ML client.
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

// Mock match history repository.
type mockMatchHistoryRepository struct {
	matchesByTeam map[string][]models.Match
	errByTeam     map[string]error
	calls         []historyQuery
}

type historyQuery struct {
	team   string
	before time.Time
	limit  int
}

func (m *mockMatchHistoryRepository) FindRecentByTeam(
	ctx context.Context,
	team string,
	before time.Time,
	limit int,
) ([]models.Match, error) {
	m.calls = append(m.calls, historyQuery{
		team:   team,
		before: before,
		limit:  limit,
	})

	if err := m.errByTeam[team]; err != nil {
		return nil, err
	}

	return m.matchesByTeam[team], nil
}

// Test helpers.
func intPointer(value int) *int {
	return &value
}

func testMatch(
	date string,
	home string,
	away string,
	homeGoals int,
	awayGoals int,
	homeSOT *int,
	awaySOT *int,
) models.Match {
	parsedDate, err := time.Parse(time.DateOnly, date)
	if err != nil {
		panic(err)
	}

	return models.Match{
		Date:     parsedDate,
		HomeTeam: home,
		AwayTeam: away,
		FTHG:     homeGoals,
		FTAG:     awayGoals,
		HST:      homeSOT,
		AST:      awaySOT,
	}
}

func newTestRepository() *mockMatchHistoryRepository {
	commonMatch := testMatch(
		"2026-09-20",
		"Chelsea",
		"Arsenal",
		1,
		1,
		intPointer(4),
		intPointer(6),
	)

	return &mockMatchHistoryRepository{
		matchesByTeam: map[string][]models.Match{
			"Arsenal": {
				testMatch(
					"2026-10-20",
					"Arsenal",
					"Liverpool",
					2,
					1,
					intPointer(5),
					intPointer(3),
				),
				testMatch(
					"2026-10-05",
					"Man United",
					"Arsenal",
					1,
					2,
					intPointer(4),
					intPointer(7),
				),
				commonMatch,
				testMatch(
					"2026-11-15",
					"Arsenal",
					"Everton",
					2,
					0,
					intPointer(5),
					intPointer(2),
				),
				testMatch(
					"2026-11-16",
					"Arsenal",
					"Spurs",
					1,
					0,
					intPointer(4),
					intPointer(1),
				),
			},
			"Chelsea": {
				testMatch(
					"2026-10-22",
					"Chelsea",
					"Wolves",
					3,
					0,
					intPointer(8),
					intPointer(2),
				),
				testMatch(
					"2026-10-10",
					"Spurs",
					"Chelsea",
					1,
					2,
					nil,
					intPointer(5),
				),
				commonMatch,
				testMatch(
					"2026-11-15",
					"Chelsea",
					"Everton",
					1,
					0,
					intPointer(4),
					intPointer(1),
				),
			},
		},
		errByTeam: make(map[string]error),
	}
}

func newTestService(
	t *testing.T,
	repository *mockMatchHistoryRepository,
	client *mockPredictionClient,
) *PredictionService {
	t.Helper()

	svc, err := NewPredictionService(client, repository)
	if err != nil {
		t.Fatalf("failed to create prediction service: %v", err)
	}

	return svc
}

func validInput() dto.PredictionInput {
	return dto.PredictionInput{
		HomeTeam:    "Arsenal",
		AwayTeam:    "Chelsea",
		KickoffDate: "2026-11-15",
	}
}

// Test constructor dependencies.
func TestNewPredictionServiceRejectsNilDependencies(t *testing.T) {
	client := &mockPredictionClient{}
	repository := newTestRepository()

	tests := []struct {
		name       string
		client     PredictionClient
		repository MatchHistoryRepository
	}{
		{
			name:       "nil ML client",
			client:     nil,
			repository: repository,
		},
		{
			name:       "nil match repository",
			client:     client,
			repository: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewPredictionService(
				tt.client,
				tt.repository,
			)

			if err == nil {
				t.Fatal("expected constructor error")
			}
		})
	}
}

// Test successful prediction.
func TestPredictBuildsHistoryAndCallsMLClient(t *testing.T) {
	repository := newTestRepository()

	client := &mockPredictionClient{
		response: dto.PredictionResponse{
			HomeTeam:     "Arsenal",
			AwayTeam:     "Chelsea",
			FeatureCount: 12,
			ModelVersion: "epl-result-live-2026-2027-v1",
		},
	}

	svc := newTestService(t, repository, client)

	result, err := svc.Predict(
		context.Background(),
		validInput(),
	)
	if err != nil {
		t.Fatalf("prediction failed: %v", err)
	}

	// Repository must be called once for each team.
	if len(repository.calls) != 2 {
		t.Fatalf(
			"expected 2 repository calls, got %d",
			len(repository.calls),
		)
	}

	for _, query := range repository.calls {
		if query.limit != 5 {
			t.Errorf("expected history limit 5, got %d", query.limit)
		}

		if query.before.Format(time.DateOnly) != "2026-11-15" {
			t.Errorf("unexpected cutoff date: %s", query.before)
		}
	}

	// ML client must be called exactly once.
	if client.calls != 1 {
		t.Fatalf(
			"expected 1 ML client call, got %d",
			client.calls,
		)
	}

	payload := client.request

	if payload.HomeTeam != "Arsenal" {
		t.Errorf("unexpected home team: %s", payload.HomeTeam)
	}

	if payload.AwayTeam != "Chelsea" {
		t.Errorf("unexpected away team: %s", payload.AwayTeam)
	}

	if payload.KickoffDate != "2026-11-15" {
		t.Errorf("unexpected kickoff date: %s", payload.KickoffDate)
	}

	// Six valid history entries contain one shared fixture.
	if len(payload.HistoricalMatches) != 5 {
		t.Fatalf(
			"expected 5 unique historical matches, got %d",
			len(payload.HistoricalMatches),
		)
	}

	history := payload.HistoricalMatches

	// History must be chronological.
	if history[0].Date != "2026-09-20" {
		t.Errorf("unexpected first date: %s", history[0].Date)
	}

	if history[4].Date != "2026-10-22" {
		t.Errorf("unexpected last date: %s", history[4].Date)
	}

	// Same-day and future matches must be excluded.
	for _, match := range history {
		if match.Date >= "2026-11-15" {
			t.Errorf(
				"same-day or future match included: %s",
				match.Date,
			)
		}
	}

	// Verify nullable SOT conversion.
	if history[2].HomeShotsOnTarget != nil {
		t.Error("expected missing home SOT to remain nil")
	}

	if history[2].AwayShotsOnTarget == nil {
		t.Fatal("expected away SOT to be present")
	}

	if *history[2].AwayShotsOnTarget != 5 {
		t.Errorf(
			"expected away SOT 5, got %v",
			*history[2].AwayShotsOnTarget,
		)
	}

	if result.FeatureCount != 12 {
		t.Errorf("expected 12 features, got %d", result.FeatureCount)
	}

	if result.ModelVersion != "epl-result-live-2026-2027-v1" {
		t.Errorf("unexpected model version: %s", result.ModelVersion)
	}
}

// Test insufficient home history.
func TestPredictRejectsInsufficientHomeHistory(t *testing.T) {
	repository := newTestRepository()
	repository.matchesByTeam["Arsenal"] =
		repository.matchesByTeam["Arsenal"][:2]

	client := &mockPredictionClient{}
	svc := newTestService(t, repository, client)

	_, err := svc.Predict(
		context.Background(),
		validInput(),
	)

	if err == nil {
		t.Fatal("expected insufficient home history error")
	}

	if !strings.Contains(err.Error(), "Arsenal") {
		t.Errorf("expected error to mention Arsenal, got %v", err)
	}

	if client.calls != 0 {
		t.Error("ML client should not be called")
	}
}

// Test insufficient away history.
func TestPredictRejectsInsufficientAwayHistory(t *testing.T) {
	repository := newTestRepository()
	repository.matchesByTeam["Chelsea"] =
		repository.matchesByTeam["Chelsea"][:2]

	client := &mockPredictionClient{}
	svc := newTestService(t, repository, client)

	_, err := svc.Predict(
		context.Background(),
		validInput(),
	)

	if err == nil {
		t.Fatal("expected insufficient away history error")
	}

	if !strings.Contains(err.Error(), "Chelsea") {
		t.Errorf("expected error to mention Chelsea, got %v", err)
	}

	if client.calls != 0 {
		t.Error("ML client should not be called")
	}
}

// Test repository errors.
func TestPredictPropagatesRepositoryError(t *testing.T) {
	expectedErr := errors.New("database connection failed")

	repository := newTestRepository()
	repository.errByTeam["Arsenal"] = expectedErr

	client := &mockPredictionClient{}
	svc := newTestService(t, repository, client)

	_, err := svc.Predict(
		context.Background(),
		validInput(),
	)

	if !errors.Is(err, expectedErr) {
		t.Fatalf("expected repository error, got %v", err)
	}

	if client.calls != 0 {
		t.Error("ML client should not be called")
	}
}

// Test ML client errors.
func TestPredictPropagatesClientError(t *testing.T) {
	expectedErr := errors.New("ML service unavailable")

	repository := newTestRepository()
	client := &mockPredictionClient{
		err: expectedErr,
	}

	svc := newTestService(t, repository, client)

	_, err := svc.Predict(
		context.Background(),
		validInput(),
	)

	if !errors.Is(err, expectedErr) {
		t.Fatalf("expected client error, got %v", err)
	}
}

// Test invalid user input.
func TestPredictRejectsInvalidInput(t *testing.T) {
	tests := []struct {
		name  string
		input dto.PredictionInput
	}{
		{
			name: "missing home team",
			input: dto.PredictionInput{
				AwayTeam:    "Chelsea",
				KickoffDate: "2026-11-15",
			},
		},
		{
			name: "missing away team",
			input: dto.PredictionInput{
				HomeTeam:    "Arsenal",
				KickoffDate: "2026-11-15",
			},
		},
		{
			name: "same teams",
			input: dto.PredictionInput{
				HomeTeam:    "Arsenal",
				AwayTeam:    "arsenal",
				KickoffDate: "2026-11-15",
			},
		},
		{
			name: "invalid kickoff date",
			input: dto.PredictionInput{
				HomeTeam:    "Arsenal",
				AwayTeam:    "Chelsea",
				KickoffDate: "15-11-2026",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repository := newTestRepository()
			client := &mockPredictionClient{}
			svc := newTestService(t, repository, client)

			_, err := svc.Predict(
				context.Background(),
				tt.input,
			)

			if err == nil {
				t.Fatal("expected validation error")
			}

			if client.calls != 0 {
				t.Error("ML client should not be called")
			}

			if len(repository.calls) != 0 {
				t.Error("repository should not be called for invalid input")
			}
		})
	}
}

// Test uninitialized service.
func TestPredictRejectsUninitializedService(t *testing.T) {
	var svc *PredictionService

	_, err := svc.Predict(
		context.Background(),
		dto.PredictionInput{},
	)

	if err == nil {
		t.Fatal("expected error for uninitialized service")
	}
}
