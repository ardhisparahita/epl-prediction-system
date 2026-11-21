package repository

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/ardhisparahita/epl-prediction-system/backend/internal/models"
	"gorm.io/gorm"
)

// MatchRepository handles historical match queries.
type MatchRepository struct {
	db *gorm.DB
}

type MatchFilters struct {
	Team string
	From *time.Time
	To   *time.Time
}

// NewMatchRepository creates a new match repository.
func NewMatchRepository(db *gorm.DB) (*MatchRepository, error) {
	if db == nil {
		return nil, errors.New("database connection is required")
	}

	return &MatchRepository{
		db: db,
	}, nil
}

// FindRecentByTeam retrieves the most recent matches involving a team
// before the specified fixture date.
func (r *MatchRepository) FindRecentByTeam(
	ctx context.Context,
	team string,
	before time.Time,
	limit int,
) ([]models.Match, error) {
	if r == nil || r.db == nil {
		return nil, errors.New("match repository is not initialized")
	}

	team = strings.TrimSpace(team)
	if team == "" {
		return nil, errors.New("team name is required")
	}

	if limit <= 0 {
		return nil, errors.New("limit must be greater than zero")
	}

	var matches []models.Match

	err := r.db.WithContext(ctx).
		Where(
			"(home_team = ? OR away_team = ?) AND date < ?",
			team,
			team,
			before.Format(time.DateOnly),
		).
		Order("date DESC").
		Order("id DESC").
		Limit(limit).
		Find(&matches).
		Error

	if err != nil {
		return nil, fmt.Errorf(
			"find recent matches for team %q: %w",
			team,
			err,
		)
	}

	return matches, nil
}

// FindMatches retrieves completed matches with filters and pagination.
func (r *MatchRepository) FindMatches(
	ctx context.Context,
	filters MatchFilters,
	limit int,
	offset int,
) ([]models.Match, int64, error) {
	if r == nil || r.db == nil {
		return nil, 0, errors.New("match repository is not initialized")
	}

	if limit <= 0 {
		return nil, 0, errors.New("limit must be greater than zero")
	}

	if offset < 0 {
		return nil, 0, errors.New("offset must not be negative")
	}

	filters.Team = strings.TrimSpace(filters.Team)

	if filters.From != nil && filters.To != nil &&
		filters.From.After(*filters.To) {
		return nil, 0, errors.New("start date must not be after end date")
	}

	applyFilters := func(query *gorm.DB) *gorm.DB {
		if filters.Team != "" {
			query = query.Where(
				"(LOWER(home_team) = LOWER(?) OR LOWER(away_team) = LOWER(?))",
				filters.Team,
				filters.Team,
			)
		}

		if filters.From != nil {
			query = query.Where(
				"date >= ?",
				filters.From.Format(time.DateOnly),
			)
		}

		if filters.To != nil {
			query = query.Where(
				"date <= ?",
				filters.To.Format(time.DateOnly),
			)
		}

		return query
	}

	var total int64

	countQuery := applyFilters(
		r.db.WithContext(ctx).Model(&models.Match{}),
	)

	if err := countQuery.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("count matches: %w", err)
	}

	var matches []models.Match

	listQuery := applyFilters(
		r.db.WithContext(ctx).Model(&models.Match{}),
	)

	err := listQuery.
		Order("date DESC").
		Order("id DESC").
		Limit(limit).
		Offset(offset).
		Find(&matches).
		Error

	if err != nil {
		return nil, 0, fmt.Errorf("find matches: %w", err)
	}

	return matches, total, nil
}
