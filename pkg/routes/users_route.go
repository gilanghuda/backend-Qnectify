package routes

import (
	"github.com/gilanghuda/backend-Quizzo/app/controllers"
	"github.com/gilanghuda/backend-Quizzo/pkg/middleware"
	"github.com/gofiber/fiber/v2"
)

func RegisterUserRoutes(app *fiber.App) {
	// Public routes
	app.Post("/signup", controllers.UserSignUp)
	app.Post("/signin", controllers.UserSignIn)

	// Protected routes example
	user := app.Group("/user", middleware.JWTProtected())
	user.Get("/profile", controllers.UserProfile) // You need to implement UserProfile
}
