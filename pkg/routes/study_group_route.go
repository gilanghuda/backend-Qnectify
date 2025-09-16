package routes

import (
	"github.com/gilanghuda/backend-Quizzo/app/controllers"
	"github.com/gilanghuda/backend-Quizzo/pkg/middleware"
	"github.com/gofiber/fiber/v2"
)

func RegisterStudyGroupRoutes(app *fiber.App) {
	studyGroup := app.Group("/study-group", middleware.JWTProtected())
	studyGroup.Post("/create", controllers.CreateStudyGroup)
	studyGroup.Get("/:id", controllers.GetStudyGroup)
	studyGroup.Put("/:id", controllers.UpdateStudyGroup)
	studyGroup.Delete("/:id", controllers.DeleteStudyGroup)
}
