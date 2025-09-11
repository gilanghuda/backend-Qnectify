package routes

import (
	"github.com/gilanghuda/backend-Quizzo/app/controllers"
	"github.com/gilanghuda/backend-Quizzo/pkg/middleware"
	"github.com/gofiber/fiber/v2"
)

func RegisterQuizRoutes(app *fiber.App) {
	app.Post("/quizzes", controllers.UploadAndGenerateQuiz)

	quiz := app.Group("/quiz", middleware.JWTProtected())
	quiz.Post("/upload", controllers.UploadAndGenerateQuiz)
}
