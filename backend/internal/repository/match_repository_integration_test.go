//go:build integration

package repository

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/ardhisparahita/epl-prediction-system/backend/config"
	"github.com/ardhisparahita/epl-prediction-system/backend/database"
	"github.com/ardhisparahita/epl-prediction-system/backend/internal/models"
	"gorm.io/gorm"
)

// setupMatchRepositoryIntegrationTest initializes a PostgreSQL
// connection and transaction for repository integration tests.
func setupMatchRepositoryIntegrationTest(
	t *testing.T,
) (*MatchRepository, *gorm.DB) {
	t.Helper()

	// Go runs tests from the package directory.
	// Move to the backend directory so config.Load()
	// can correctly locate the project root .env file.
	originalWD, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get working directory: %v", err)
	}

	backendDir := filepath.Clean(
		filepath.Join(originalWD, "..", ".."),
	)

	if err := os.Chdir(backendDir); err != nil {
		t.Fatalf("failed to change working directory: %v", err)
	}

	// Restore the original working directory after the test.
	t.Cleanup(func() {
		if err := os.Chdir(originalWD); err != nil {
			t.Errorf("failed to restore working directory: %v", err)
		}
	})

	// Load project configuration.
	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("failed to load config: %v", err)
	}

	// Connect to PostgreSQL.
	db, err := database.ConnectPostgres(cfg)
	if err != nil {
		t.Fatalf("failed to connect PostgreSQL: %v", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		t.Fatalf("failed to get SQL database: %v", err)
	}

	// Close database connection after transaction rollback.
	t.Cleanup(func() {
		if err := sqlDB.Close(); err != nil {
			t.Errorf("failed to close database connection: %v", err)
		}
	})

	// Begin transaction.
	tx := db.Begin()
	if tx.Error != nil {
		t.Fatalf("failed to begin transaction: %v", tx.Error)
	}

	// Rollback all test data after each test.
	t.Cleanup(func() {
		_ = tx.Rollback().Error
	})

	// Initialize repository using the transaction.
	repo, err := NewMatchRepository(tx)
	if err != nil {
		t.Fatalf("failed to create repository: %v", err)
	}

	return repo, tx
}

// parseTestDate converts a date string into time.Time.
func parseTestDate(
	t *testing.T,
	value string,
) time.Time {
	t.Helper()

	date, err := time.Parse(time.DateOnly, value)
	if err != nil {
		t.Fatalf("invalid test date %s: %v", value, err)
	}

	return date
}

// TestFindRecentByTeamRespectsCutoffAndLimit verifies that
// the repository retrieves home and away matches,
// excludes matches on or after the cutoff date,
// and respects the requested limit.
func TestFindRecentByTeamRespectsCutoffAndLimit(t *testing.T) {
	repo, tx := setupMatchRepositoryIntegrationTest(t)

	// Generate unique team names to avoid conflicts
	// with existing historical matches.
	suffix := fmt.Sprintf("%d", time.Now().UnixNano())
	team := "RepositoryTestHome" + suffix

	// Prepare test fixtures.
	fixtures := []models.Match{
		{
			Date:     parseTestDate(t, "2026-01-10"),
			HomeTeam: team,
			AwayTeam: "OpponentA" + suffix,
			FTHG:     2,
			FTAG:     1,
		},
		{
			Date:     parseTestDate(t, "2026-01-20"),
			HomeTeam: "OpponentB" + suffix,
			AwayTeam: team,
			FTHG:     1,
			FTAG:     2,
		},
		{
			Date:     parseTestDate(t, "2026-02-01"),
			HomeTeam: team,
			AwayTeam: "OpponentC" + suffix,
			FTHG:     3,
			FTAG:     0,
		},
		{
			Date:     parseTestDate(t, "2026-02-10"),
			HomeTeam: "OpponentD" + suffix,
			AwayTeam: team,
			FTHG:     0,
			FTAG:     1,
		},
		{
			Date:     parseTestDate(t, "2026-03-01"),
			HomeTeam: team,
			AwayTeam: "OpponentE" + suffix,
			FTHG:     2,
			FTAG:     0,
		},
		{
			// Same date as kickoff: must be excluded.
			Date:     parseTestDate(t, "2026-04-01"),
			HomeTeam: team,
			AwayTeam: "OpponentF" + suffix,
			FTHG:     1,
			FTAG:     1,
		},
		{
			// Future match: must be excluded.
			Date:     parseTestDate(t, "2026-05-01"),
			HomeTeam: team,
			AwayTeam: "OpponentG" + suffix,
			FTHG:     2,
			FTAG:     0,
		},
	}

	// Insert fixtures inside the transaction.
	// All inserted records will be rolled back after the test.
	if err := tx.Create(&fixtures).Error; err != nil {
		t.Fatalf("failed to insert test fixtures: %v", err)
	}

	// Define kickoff date.
	before := parseTestDate(t, "2026-04-01")

	// Retrieve the three most recent matches.
	matches, err := repo.FindRecentByTeam(
		context.Background(),
		team,
		before,
		3,
	)
	if err != nil {
		t.Fatalf("failed to retrieve matches: %v", err)
	}

	// Expected results in descending date order.
	expectedDates := []string{
		"2026-03-01",
		"2026-02-10",
		"2026-02-01",
	}

	if len(matches) != len(expectedDates) {
		t.Fatalf(
			"expected %d matches, got %d",
			len(expectedDates),
			len(matches),
		)
	}

	// Verify date ordering and expected records.
	for i, match := range matches {
		actualDate := match.Date.Format(time.DateOnly)

		if actualDate != expectedDates[i] {
			t.Errorf(
				"match %d: expected date %s, got %s",
				i,
				expectedDates[i],
				actualDate,
			)
		}
	}

	// Verify that both home and away matches are included.
	hasHomeMatch := false
	hasAwayMatch := false

	for _, match := range matches {
		if match.HomeTeam == team {
			hasHomeMatch = true
		}

		if match.AwayTeam == team {
			hasAwayMatch = true
		}

		// Verify cutoff date.
		if !match.Date.Before(before) {
			t.Errorf(
				"match on or after cutoff included: %s",
				match.Date.Format(time.DateOnly),
			)
		}
	}

	if !hasHomeMatch {
		t.Error("expected at least one home match")
	}

	if !hasAwayMatch {
		t.Error("expected at least one away match")
	}
}

// TestFindRecentByTeamRejectsInvalidArguments verifies
// that empty team names and invalid limits are rejected.
func TestFindRecentByTeamRejectsInvalidArguments(t *testing.T) {
	repo, _ := setupMatchRepositoryIntegrationTest(t)

	before := parseTestDate(t, "2026-11-15")

	tests := []struct {
		name  string
		team  string
		limit int
	}{
		{
			name:  "empty team",
			team:  "",
			limit: 5,
		},
		{
			name:  "zero limit",
			team:  "Arsenal",
			limit: 0,
		},
		{
			name:  "negative limit",
			team:  "Arsenal",
			limit: -1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := repo.FindRecentByTeam(
				context.Background(),
				tt.team,
				before,
				tt.limit,
			)

			if err == nil {
				t.Fatal("expected validation error")
			}
		})
	}
}

func TestFindRecentByTeamReturnsEmptyWhenNoHistory(t *testing.T) {
	repo, _ := setupMatchRepositoryIntegrationTest(t)

	suffix := fmt.Sprintf("%d", time.Now().UnixNano())
	team := "RepositoryNoHistory" + suffix

	before := parseTestDate(t, "2026-11-15")

	matches, err := repo.FindRecentByTeam(
		context.Background(),
		team,
		before,
		5,
	)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(matches) != 0 {
		t.Fatalf(
			"expected 0 matches, got %d",
			len(matches),
		)
	}
}

func TestFindMatchesFiltersAndPaginates(t *testing.T) {
	repo, tx := setupMatchRepositoryIntegrationTest(t)

	suffix := fmt.Sprintf("%d", time.Now().UnixNano())
	team := "FindMatchesTeam" + suffix

	fixtures := []models.Match{
		{
			Date:     parseTestDate(t, "2026-01-10"),
			HomeTeam: team,
			AwayTeam: "OpponentA" + suffix,
			FTHG:     2,
			FTAG:     1,
		},
		{
			Date:     parseTestDate(t, "2026-01-20"),
			HomeTeam: "OpponentB" + suffix,
			AwayTeam: team,
			FTHG:     1,
			FTAG:     2,
		},
		{
			Date:     parseTestDate(t, "2026-02-01"),
			HomeTeam: team,
			AwayTeam: "OpponentC" + suffix,
			FTHG:     3,
			FTAG:     0,
		},
		{
			Date:     parseTestDate(t, "2026-02-10"),
			HomeTeam: "UnrelatedHome" + suffix,
			AwayTeam: "UnrelatedAway" + suffix,
			FTHG:     1,
			FTAG:     1,
		},
		{
			Date:     parseTestDate(t, "2026-03-01"),
			HomeTeam: team,
			AwayTeam: "OpponentD" + suffix,
			FTHG:     2,
			FTAG:     0,
		},
	}

	if err := tx.Create(&fixtures).Error; err != nil {
		t.Fatalf("failed to insert fixtures: %v", err)
	}

	from := parseTestDate(t, "2026-01-15")
	to := parseTestDate(t, "2026-02-28")

	matches, total, err := repo.FindMatches(
		context.Background(),
		MatchFilters{
			Team: team,
			From: &from,
			To:   &to,
		},
		1,
		1,
	)

	if err != nil {
		t.Fatalf("failed to find matches: %v", err)
	}

	if total != 2 {
		t.Fatalf("expected total 2, got %d", total)
	}

	if len(matches) != 1 {
		t.Fatalf("expected 1 match on current page, got %d", len(matches))
	}

	if actual := matches[0].Date.Format(time.DateOnly); actual != "2026-01-20" {
		t.Errorf("expected match date 2026-01-20, got %s", actual)
	}

	if matches[0].AwayTeam != team {
		t.Errorf("expected away team %s, got %s", team, matches[0].AwayTeam)
	}
}

func TestFindMatchesRejectsInvalidArguments(t *testing.T) {
	repo, _ := setupMatchRepositoryIntegrationTest(t)

	from := parseTestDate(t, "2026-03-01")
	to := parseTestDate(t, "2026-01-01")

	tests := []struct {
		name    string
		limit   int
		offset  int
		filters MatchFilters
	}{
		{
			name:   "zero limit",
			limit:  0,
			offset: 0,
		},
		{
			name:   "negative limit",
			limit:  -1,
			offset: 0,
		},
		{
			name:   "negative offset",
			limit:  10,
			offset: -1,
		},
		{
			name:   "invalid date range",
			limit:  10,
			offset: 0,
			filters: MatchFilters{
				From: &from,
				To:   &to,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, _, err := repo.FindMatches(
				context.Background(),
				tt.filters,
				tt.limit,
				tt.offset,
			)

			if err == nil {
				t.Fatal("expected error, got nil")
			}
		})
	}
}
