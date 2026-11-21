package service

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/ardhisparahita/epl-prediction-system/backend/internal/dto"
	"github.com/ardhisparahita/epl-prediction-system/backend/internal/models"
	"github.com/ardhisparahita/epl-prediction-system/backend/internal/repository"
)

const (
	defaultMatchPage  = 1
	defaultMatchLimit = 20
	maxMatchLimit     = 100
)

// MatchListRepository defines the contract for retrieving matches.
type MatchListRepository interface {
	FindMatches(
		ctx context.Context,
		filters repository.MatchFilters,
		limit int,
		offset int,
	) ([]models.Match, int64, error)
}

// MatchService handles match history operations.
type MatchService struct {
	matchRepo MatchListRepository
}

// NewMatchService creates a new MatchService.
func NewMatchService(
	matchRepo MatchListRepository,
) (*MatchService, error) {
	if matchRepo == nil {
		return nil, errors.New("match repository is required")
	}

	return &MatchService{
		matchRepo: matchRepo,
	}, nil
}

// List retrieves matches with filters and pagination.
func (s *MatchService) List(
	ctx context.Context,
	input dto.MatchListInput,
) (dto.MatchListResponse, error) {
	var result dto.MatchListResponse

	if s == nil || s.matchRepo == nil {
		return result, errors.New("match service is not initialized")
	}

	team := strings.TrimSpace(input.Team)
	fromText := strings.TrimSpace(input.From)
	toText := strings.TrimSpace(input.To)

	page := input.Page
	if page == 0 {
		page = defaultMatchPage
	}
	if page < 1 {
		return result, &ValidationError{
			Message: "page must be greater than zero",
		}
	}

	limit := input.Limit
	if limit == 0 {
		limit = defaultMatchLimit
	}
	if limit < 1 || limit > maxMatchLimit {
		return result, &ValidationError{
			Message: fmt.Sprintf(
				"limit must be between 1 and %d",
				maxMatchLimit,
			),
		}
	}

	// Prevent integer overflow when calculating the offset.
	maxInt := int(^uint(0) >> 1)
	if page-1 > maxInt/limit {
		return result, &ValidationError{
			Message: "page is too large",
		}
	}

	var from *time.Time
	if fromText != "" {
		parsed, err := time.Parse(time.DateOnly, fromText)
		if err != nil {
			return result, &ValidationError{
				Message: "from must use YYYY-MM-DD format",
			}
		}
		from = &parsed
	}

	var to *time.Time
	if toText != "" {
		parsed, err := time.Parse(time.DateOnly, toText)
		if err != nil {
			return result, &ValidationError{
				Message: "to must use YYYY-MM-DD format",
			}
		}
		to = &parsed
	}

	if from != nil && to != nil && from.After(*to) {
		return result, &ValidationError{
			Message: "from date must not be after to date",
		}
	}

	offset := (page - 1) * limit

	matches, total, err := s.matchRepo.FindMatches(
		ctx,
		repository.MatchFilters{
			Team: team,
			From: from,
			To:   to,
		},
		limit,
		offset,
	)
	if err != nil {
		return result, &RepositoryError{
			Err: fmt.Errorf("retrieve match history: %w", err),
		}
	}

	items := make([]dto.MatchListItem, 0, len(matches))

	for _, match := range matches {
		items = append(items, dto.MatchListItem{
			ID:                match.ID,
			Date:              match.Date.Format(time.DateOnly),
			HomeTeam:          match.HomeTeam,
			AwayTeam:          match.AwayTeam,
			HomeGoals:         match.FTHG,
			AwayGoals:         match.FTAG,
			HomeShotsOnTarget: match.HST,
			AwayShotsOnTarget: match.AST,
		})
	}

	var totalPages int64
	if total > 0 {
		totalPages = (total-1)/int64(limit) + 1
	}

	return dto.MatchListResponse{
		Data: items,
		Pagination: dto.MatchPagination{
			Page:       page,
			Limit:      limit,
			Total:      total,
			TotalPages: totalPages,
		},
	}, nil
}
