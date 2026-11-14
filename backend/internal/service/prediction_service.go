package service

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/ardhisparahita/epl-prediction-system/backend/internal/dto"
	"github.com/ardhisparahita/epl-prediction-system/backend/internal/models"
)

// PredictionClient defines the contract for communicating
// with the Python ML Service.
type PredictionClient interface {
	Predict(
		ctx context.Context,
		payload dto.PredictionRequest,
	) (dto.PredictionResponse, error)
}

// MatchHistoryRepository defines the contract for retrieving
// historical matches from PostgreSQL.
type MatchHistoryRepository interface {
	FindRecentByTeam(
		ctx context.Context,
		team string,
		before time.Time,
		limit int,
	) ([]models.Match, error)
}

const (
	historyWindow         = 5
	minimumHistoryMatches = 3
)

// PredictionService handles match prediction operations.
type PredictionService struct {
	mlClient  PredictionClient
	matchRepo MatchHistoryRepository
}

// NewPredictionService creates a new PredictionService.
func NewPredictionService(
	mlClient PredictionClient,
	matchRepo MatchHistoryRepository,
) (*PredictionService, error) {
	if mlClient == nil {
		return nil, errors.New("ML client is required")
	}

	if matchRepo == nil {
		return nil, errors.New("match history repository is required")
	}

	return &PredictionService{
		mlClient:  mlClient,
		matchRepo: matchRepo,
	}, nil
}

// Predict retrieves historical matches and sends them to the ML Service.
func (s *PredictionService) Predict(
	ctx context.Context,
	input dto.PredictionInput,
) (dto.PredictionResponse, error) {
	var result dto.PredictionResponse

	if s == nil || s.mlClient == nil || s.matchRepo == nil {
		return result, errors.New(
			"prediction service is not initialized",
		)
	}

	homeTeam := strings.TrimSpace(input.HomeTeam)
	awayTeam := strings.TrimSpace(input.AwayTeam)
	kickoffDateText := strings.TrimSpace(input.KickoffDate)

	if homeTeam == "" {
		return result, &ValidationError{
			Message: "home_team is required",
		}
	}

	if awayTeam == "" {
		return result, errors.New("away_team is required")
	}

	if strings.EqualFold(homeTeam, awayTeam) {
		return result, errors.New(
			"home_team and away_team must be different",
		)
	}

	kickoffDate, err := time.Parse(
		time.DateOnly,
		kickoffDateText,
	)
	if err != nil {
		return result, &ValidationError{
			Message: "kickoff_date must use YYYY-MM-DD format",
		}
	}

	// Retrieve the latest five matches for the home team.
	homeMatches, err := s.matchRepo.FindRecentByTeam(
		ctx,
		homeTeam,
		kickoffDate,
		historyWindow,
	)
	if err != nil {
		return result, &RepositoryError{
			Err: fmt.Errorf(
				"retrieve history for home team %q: %w",
				homeTeam,
				err,
			),
		}
	}

	// Retrieve the latest five matches for the away team.
	awayMatches, err := s.matchRepo.FindRecentByTeam(
		ctx,
		awayTeam,
		kickoffDate,
		historyWindow,
	)
	if err != nil {
		return result, &RepositoryError{
			Err: fmt.Errorf(
				"retrieve history for away team %q: %w",
				awayTeam,
				err,
			),
		}
	}

	// Apply the cutoff date and history window.
	homeMatches = filterRecentMatches(homeMatches, kickoffDate)
	awayMatches = filterRecentMatches(awayMatches, kickoffDate)

	// Each team must have at least three previous matches.
	if len(homeMatches) < minimumHistoryMatches {
		return result, &ValidationError{
			Message: fmt.Sprintf(
				"insufficient match history for home team %q: found %d, minimum required %d",
				homeTeam,
				len(homeMatches),
				minimumHistoryMatches,
			),
		}
	}

	if len(awayMatches) < minimumHistoryMatches {
		return result, fmt.Errorf(
			"insufficient match history for away team %q: found %d, minimum required %d",
			awayTeam,
			len(awayMatches),
			minimumHistoryMatches,
		)
	}

	// Merge both histories and remove duplicated fixtures.
	historicalMatches := mergeMatchHistory(
		homeMatches,
		awayMatches,
	)

	// Build the payload expected by the Python ML Service.
	payload := dto.PredictionRequest{
		HomeTeam:          homeTeam,
		AwayTeam:          awayTeam,
		KickoffDate:       kickoffDate.Format(time.DateOnly),
		HistoricalMatches: convertHistoricalMatches(historicalMatches),
	}

	return s.mlClient.Predict(ctx, payload)
}

// filterRecentMatches keeps only matches before kickoff,
// sorts them newest first, and limits the result to five matches.
func filterRecentMatches(
	matches []models.Match,
	kickoffDate time.Time,
) []models.Match {
	filtered := make([]models.Match, 0, len(matches))

	kickoffDateText := kickoffDate.Format(time.DateOnly)

	for _, match := range matches {
		if match.Date.Format(time.DateOnly) < kickoffDateText {
			filtered = append(filtered, match)
		}
	}

	sort.Slice(filtered, func(i, j int) bool {
		if !filtered[i].Date.Equal(filtered[j].Date) {
			return filtered[i].Date.After(filtered[j].Date)
		}

		return filtered[i].ID > filtered[j].ID
	})

	if len(filtered) > historyWindow {
		filtered = filtered[:historyWindow]
	}

	return filtered
}

// mergeMatchHistory combines both teams' histories
// without duplicating the same fixture.
func mergeMatchHistory(
	homeMatches []models.Match,
	awayMatches []models.Match,
) []models.Match {
	type matchKey struct {
		Date     string
		HomeTeam string
		AwayTeam string
	}

	seen := make(map[matchKey]struct{})

	merged := make(
		[]models.Match,
		0,
		len(homeMatches)+len(awayMatches),
	)

	allMatches := make([]models.Match, 0, len(homeMatches)+len(awayMatches))
	allMatches = append(allMatches, homeMatches...)
	allMatches = append(allMatches, awayMatches...)

	for _, match := range allMatches {
		key := matchKey{
			Date:     match.Date.Format(time.DateOnly),
			HomeTeam: match.HomeTeam,
			AwayTeam: match.AwayTeam,
		}

		if _, exists := seen[key]; exists {
			continue
		}

		seen[key] = struct{}{}
		merged = append(merged, match)
	}

	// Return history in chronological order.
	sort.Slice(merged, func(i, j int) bool {
		if !merged[i].Date.Equal(merged[j].Date) {
			return merged[i].Date.Before(merged[j].Date)
		}

		if merged[i].HomeTeam != merged[j].HomeTeam {
			return merged[i].HomeTeam < merged[j].HomeTeam
		}

		return merged[i].AwayTeam < merged[j].AwayTeam
	})

	return merged
}

// convertHistoricalMatches converts PostgreSQL models
// into the DTO expected by the Python ML Service.
func convertHistoricalMatches(
	matches []models.Match,
) []dto.HistoricalMatch {
	result := make([]dto.HistoricalMatch, 0, len(matches))

	for _, match := range matches {
		result = append(result, dto.HistoricalMatch{
			Date:              match.Date.Format(time.DateOnly),
			HomeTeam:          match.HomeTeam,
			AwayTeam:          match.AwayTeam,
			HomeGoals:         match.FTHG,
			AwayGoals:         match.FTAG,
			HomeShotsOnTarget: convertNullableInt(match.HST),
			AwayShotsOnTarget: convertNullableInt(match.AST),
		})
	}

	return result
}

// convertNullableInt converts an optional integer into an optional float.
func convertNullableInt(value *int) *float64 {
	if value == nil {
		return nil
	}

	converted := float64(*value)
	return &converted
}
