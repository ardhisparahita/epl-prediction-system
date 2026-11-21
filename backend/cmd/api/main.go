package main

import (
	"fmt"
	"log"
	"time"

	"github.com/ardhisparahita/epl-prediction-system/backend/config"
	"github.com/ardhisparahita/epl-prediction-system/backend/database"
	"github.com/ardhisparahita/epl-prediction-system/backend/internal/client"
	"github.com/ardhisparahita/epl-prediction-system/backend/internal/repository"
	"github.com/ardhisparahita/epl-prediction-system/backend/internal/routes"
	"github.com/ardhisparahita/epl-prediction-system/backend/internal/service"

	"github.com/gofiber/fiber/v3"
)

const mlServiceTimeout = 15 * time.Second

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	// Load application configuration.
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("load configuration: %w", err)
	}

	// Connect to PostgreSQL.
	db, err := database.ConnectPostgres(cfg)
	if err != nil {
		return err
	}

	sqlDB, err := db.DB()
	if err != nil {
		return fmt.Errorf("get database instance: %w", err)
	}
	defer sqlDB.Close()

	// Initialize match history repository.
	matchRepository, err := repository.NewMatchRepository(db)
	if err != nil {
		return fmt.Errorf(
			"initialize match repository: %w",
			err,
		)
	}

	// Initialize Python ML Service client.
	mlClient, err := client.NewMLClient(
		cfg.MLServiceURL,
		mlServiceTimeout,
	)
	if err != nil {
		return fmt.Errorf("initialize ML client: %w", err)
	}

	// Initialize prediction service.
	predictionService, err := service.NewPredictionService(
		mlClient,
		matchRepository,
	)
	if err != nil {
		return fmt.Errorf("initialize prediction service: %w", err)
	}

	matchService, err := service.NewMatchService(matchRepository)
	if err != nil {
		return fmt.Errorf("initialize match service: %w", err)
	}

	// Initialize Fiber application.
	app := fiber.New(fiber.Config{
		AppName: "EPL Prediction API",
	})

	// Register application routes.
	routes.SetupRoutes(
		app,
		db,
		predictionService,
		matchService,
	)

	log.Printf(
		"EPL Prediction API starting on port %s",
		cfg.AppPort,
	)

	log.Printf(
		"ML Service configured at %s",
		cfg.MLServiceURL,
	)

	return app.Listen(":" + cfg.AppPort)
}
