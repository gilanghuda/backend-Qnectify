package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"time"

	"github.com/gilanghuda/backend-Quizzo/pkg/database"
	"github.com/gilanghuda/backend-Quizzo/pkg/routes"
	"github.com/gilanghuda/backend-Quizzo/pkg/utils"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/joho/godotenv"
)

func main() {
	if err := run(); err != nil {
		log.Fatalln(err)
	}
}

func run() error {
	_ = godotenv.Load()

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	app := fiber.New(fiber.Config{
		BodyLimit: 20 * 1024 * 1024,
	})

	app.Use(cors.New(cors.Config{
		AllowOrigins:     "https://qnectify-cyan.vercel.app/,http://localhost:3000,http://localhost:3003",
		AllowHeaders:     "Origin, Content-Type, Accept, Authorization",
		AllowMethods:     "GET,POST,PUT,DELETE,OPTIONS",
		AllowCredentials: true,
	}))

	app.Get("/", func(c *fiber.Ctx) error {
		return c.SendString("versi 1.2.3 ")
	})

	otelShutdown, err := utils.SetupTracer()
	if err != nil {
		log.Printf("failed to init otel: %v", err)
	} else {
		log.Printf("otel initialized successfully")
		defer func() {
			_ = otelShutdown(ctx)
		}()
	}

	_, err = database.InitDB()
	if err != nil {
		return err
	}

	routes.RegisterUserRoutes(app)
	routes.RegisterQuizRoutes(app)
	routes.RegisterStudyGroupRoutes(app)
	routes.RegisterSocialsRoutes(app)

	errCh := make(chan error, 1)
	go func() {
		errCh <- app.Listen(":3000")
	}()

	select {
	case <-ctx.Done():
		_, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		if err := app.Shutdown(); err != nil {
			log.Printf("failed to shutdown Fiber app: %v", err)
		}

		return nil
	case err := <-errCh:
		return err
	}
}
