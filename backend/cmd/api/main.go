package main

import (
	"log"

	"github.com/ardhisparahita/epl-prediction-system/backend/internal/routes"
	"github.com/gofiber/fiber/v3"
)

func main() {
	app := fiber.New(fiber.Config{
		AppName: "EPL Prediction API",
	})

	routes.SetupRoutes(app)

	log.Println("EPL Prediction API starting on :8080")

	if err := app.Listen(":8080"); err != nil {
		log.Fatal(err)
	}
}
