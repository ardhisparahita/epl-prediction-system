package main

import (
	"log"

	"github.com/ardhisparahita/epl-prediction-system/backend/config"
	"github.com/ardhisparahita/epl-prediction-system/backend/database"
	"github.com/ardhisparahita/epl-prediction-system/backend/internal/routes"

	"github.com/gofiber/fiber/v3"
)

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}

	db, err := database.ConnectPostgres(cfg)
	if err != nil {
		return err
	}

	sqlDB, err := db.DB()
	if err != nil {
		return err
	}
	defer sqlDB.Close()

	app := fiber.New(fiber.Config{
		AppName: "EPL Prediction API",
	})

	routes.SetupRoutes(app, db)

	log.Printf(
		"EPL Prediction API starting on port %s",
		cfg.AppPort,
	)

	return app.Listen(":" + cfg.AppPort)
}
