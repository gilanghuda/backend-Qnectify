package routes

import (
	"github.com/gilanghuda/backend-Quizzo/app/controllers"
	"github.com/gilanghuda/backend-Quizzo/pkg/middleware"
	"github.com/gofiber/fiber/v2"
)

func RegisterSocialsRoutes(app *fiber.App) {
	appGroup := app.Group("/socials")
	appGroup.Get("/likes/:id", controllers.GetLikesCount)

	s := app.Group("/socials", middleware.JWTProtected())
	s.Post("/like/:id", controllers.ToggleLike)
	s.Post("/comment/:id", controllers.AddComment)
	s.Delete("/comment/:id", controllers.DeleteComment)
	s.Get("/comments/:id", controllers.GetComments)
}
