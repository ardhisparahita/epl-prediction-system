package models

import "time"

// Match represents a completed EPL match stored in PostgreSQL.
type Match struct {
	ID uint64 `gorm:"primaryKey;autoIncrement"`

	Date time.Time `gorm:"type:date;not null;uniqueIndex:idx_matches_fixture,priority:1;index:idx_matches_date"`

	HomeTeam string `gorm:"type:varchar(100);not null;uniqueIndex:idx_matches_fixture,priority:2;index:idx_matches_home_date,priority:1"`

	AwayTeam string `gorm:"type:varchar(100);not null;uniqueIndex:idx_matches_fixture,priority:3;index:idx_matches_away_date,priority:1"`

	FTHG int `gorm:"not null"`
	FTAG int `gorm:"not null"`

	HST *int `gorm:"default:null"`
	AST *int `gorm:"default:null"`

	CreatedAt time.Time
	UpdatedAt time.Time
}

// TableName explicitly defines the PostgreSQL table name.
func (Match) TableName() string {
	return "matches"
}
