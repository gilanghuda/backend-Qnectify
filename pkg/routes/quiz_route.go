package routes

import (
	"github.com/gilanghuda/backend-Quizzo/app/controllers"
	"github.com/gofiber/fiber/v2"
)

func RegisterQuizRoutes(app *fiber.App) {
	app.Post("/quizzes", controllers.UploadAndGenerateQuiz)
}
