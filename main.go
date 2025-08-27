package main

import (
	"log"

	"github.com/gilanghuda/backend-Quizzo/pkg/database"
	"github.com/gilanghuda/backend-Quizzo/pkg/routes"
	"github.com/gofiber/fiber/v2"
	"github.com/joho/godotenv"
)

func main() {
	// Try to load .env, but don't fail if not found (for production/CI)
	_ = godotenv.Load()

	app := fiber.New()

	app.Get("/", func(c *fiber.Ctx) error {
		return c.SendString("Hello, World!")
	})

	_, err := database.InitDB()
	if err != nil {
		log.Fatalf("Failed to connect to the database: %v", err)
	}

	routes.RegisterUserRoutes(app)

	log.Fatal(app.Listen(":3000"))
}
