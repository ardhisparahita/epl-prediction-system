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
