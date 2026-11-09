package database

import (
	"context"
	"fmt"
	"net"
	"net/url"
	"time"

	"github.com/ardhisparahita/epl-prediction-system/backend/config"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func ConnectPostgres(cfg config.Config) (*gorm.DB, error) {
	dsn := url.URL{
		Scheme: "postgres",
		User:   url.UserPassword(cfg.DBUser, cfg.DBPassword),
		Host:   net.JoinHostPort(cfg.DBHost, cfg.DBPort),
		Path:   cfg.DBName,
	}

	query := url.Values{}
	query.Set("sslmode", cfg.DBSSLMode)

	rawQuery := query.Encode()
	if cfg.DBTimezone != "" {
		if rawQuery != "" {
			rawQuery += "&"
		}
		rawQuery += fmt.Sprintf("TimeZone=%s", cfg.DBTimezone)
	}
	dsn.RawQuery = rawQuery

	db, err := gorm.Open(
		postgres.Open(dsn.String()),
		&gorm.Config{},
	)

	if err != nil {
		return nil, fmt.Errorf(
			"failed to initialize postgres: %w", err,
		)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf(
			"failed to get database instance: %w", err,
		)
	}

	sqlDB.SetMaxIdleConns(5)
	sqlDB.SetMaxOpenConns(20)
	sqlDB.SetConnMaxLifetime(time.Hour)
	sqlDB.SetConnMaxIdleTime(5 * time.Minute)

	ctx, cancel := context.WithTimeout(
		context.Background(),
		5*time.Second,
	)
	defer cancel()

	if err := sqlDB.PingContext(ctx); err != nil {
		sqlDB.Close()

		return nil, fmt.Errorf(
			"failed to ping postgres: %w", err,
		)
	}

	return db, nil
}
