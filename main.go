package main

import (
	"log"

	"github.com/gilanghuda/backend-Quizzo/pkg/database"
	"github.com/gilanghuda/backend-Quizzo/pkg/routes"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/joho/godotenv"
)

func main() {
	_ = godotenv.Load()

	app := fiber.New(fiber.Config{
		BodyLimit: 20 * 1024 * 1024,
	})

	app.Use(cors.New(cors.Config{
		AllowOrigins:     "http://localhost:3000",
		AllowHeaders:     "Origin, Content-Type, Accept, Authorization",
		AllowMethods:     "GET,POST,PUT,DELETE,OPTIONS",
		AllowCredentials: true,
	}))

	app.Get("/", func(c *fiber.Ctx) error {
		return c.SendString("versi 2 ")
	})

	_, err := database.InitDB()
	if err != nil {
		log.Fatalf("Failed to connect to the database: %v", err)
	}

	routes.RegisterUserRoutes(app)
	routes.RegisterQuizRoutes(app)
	routes.RegisterStudyGroupRoutes(app)

	log.Fatal(app.Listen(":3000"))
}
