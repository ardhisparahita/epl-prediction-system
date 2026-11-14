package main

import (
	"encoding/csv"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/ardhisparahita/epl-prediction-system/backend/config"
	"github.com/ardhisparahita/epl-prediction-system/backend/database"
	"github.com/ardhisparahita/epl-prediction-system/backend/internal/models"

	"gorm.io/gorm/clause"
)

func main() {
	filePath := flag.String(
		"file",
		"../ml-service/data/processed/historical_matches.csv",
		"Path to historical matches CSV",
	)
	flag.Parse()

	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	db, err := database.ConnectPostgres(cfg)
	if err != nil {
		log.Fatalf("Failed to connect to PostgreSQL: %v", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		log.Fatalf("Failed to get database connection: %v", err)
	}
	defer sqlDB.Close()

	matches, err := readMatches(*filePath)
	if err != nil {
		log.Fatalf("Failed to read CSV: %v", err)
	}

	if len(matches) == 0 {
		log.Println("No matches found in CSV")
		return
	}

	result := db.Clauses(clause.OnConflict{
		Columns: []clause.Column{
			{Name: "date"},
			{Name: "home_team"},
			{Name: "away_team"},
		},
		DoNothing: true,
	}).CreateInBatches(&matches, 500)

	if result.Error != nil {
		log.Fatalf("Failed to import matches: %v", result.Error)
	}

	skipped := int64(len(matches)) - result.RowsAffected

	log.Println("Historical matches import completed")
	log.Printf("Total CSV records : %d", len(matches))
	log.Printf("Inserted records  : %d", result.RowsAffected)
	log.Printf("Skipped duplicates: %d", skipped)
}

func readMatches(filePath string) ([]models.Match, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("open file: %w", err)
	}
	defer file.Close()

	reader := csv.NewReader(file)
	reader.TrimLeadingSpace = true

	header, err := reader.Read()
	if err != nil {
		return nil, fmt.Errorf("read CSV header: %w", err)
	}

	columns := make(map[string]int)

	for i, name := range header {
		name = strings.TrimSpace(name)
		name = strings.TrimPrefix(name, "\uFEFF")
		columns[name] = i
	}

	required := []string{
		"Date",
		"HomeTeam",
		"AwayTeam",
		"FTHG",
		"FTAG",
		"HST",
		"AST",
	}

	for _, name := range required {
		if _, exists := columns[name]; !exists {
			return nil, fmt.Errorf("required CSV column %q not found", name)
		}
	}

	var matches []models.Match

	for line := 2; ; line++ {
		record, err := reader.Read()

		if err == io.EOF {
			break
		}

		if err != nil {
			return nil, fmt.Errorf("CSV line %d: %w", line, err)
		}

		get := func(column string) string {
			return strings.TrimSpace(record[columns[column]])
		}

		matchDate, err := time.Parse("2006-01-02", get("Date"))
		if err != nil {
			return nil, fmt.Errorf(
				"CSV line %d: invalid date: %w",
				line,
				err,
			)
		}

		homeTeam := get("HomeTeam")
		awayTeam := get("AwayTeam")

		if homeTeam == "" || awayTeam == "" {
			return nil, fmt.Errorf(
				"CSV line %d: home team or away team is empty",
				line,
			)
		}

		fthg, err := strconv.Atoi(get("FTHG"))
		if err != nil || fthg < 0 {
			return nil, fmt.Errorf("CSV line %d: invalid FTHG", line)
		}

		ftag, err := strconv.Atoi(get("FTAG"))
		if err != nil || ftag < 0 {
			return nil, fmt.Errorf("CSV line %d: invalid FTAG", line)
		}

		hst, err := parseNullableInt(get("HST"))
		if err != nil {
			return nil, fmt.Errorf("CSV line %d: invalid HST: %w", line, err)
		}

		ast, err := parseNullableInt(get("AST"))
		if err != nil {
			return nil, fmt.Errorf("CSV line %d: invalid AST: %w", line, err)
		}

		matches = append(matches, models.Match{
			Date:     matchDate,
			HomeTeam: homeTeam,
			AwayTeam: awayTeam,
			FTHG:     fthg,
			FTAG:     ftag,
			HST:      hst,
			AST:      ast,
		})
	}

	return matches, nil
}

func parseNullableInt(value string) (*int, error) {
	if value == "" || strings.EqualFold(value, "NA") ||
		strings.EqualFold(value, "NaN") ||
		strings.EqualFold(value, "null") {
		return nil, nil
	}

	number, err := strconv.Atoi(value)
	if err != nil {
		return nil, err
	}

	if number < 0 {
		return nil, fmt.Errorf("value cannot be negative")
	}

	return &number, nil
}
