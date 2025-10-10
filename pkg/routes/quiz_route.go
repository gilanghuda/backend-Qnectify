package routes

import (
	"github.com/gilanghuda/backend-Quizzo/app/controllers"
	"github.com/gilanghuda/backend-Quizzo/pkg/middleware"
	"github.com/gofiber/fiber/v2"
)

func RegisterQuizRoutes(app *fiber.App) {
	app.Post("/quizzes", controllers.UploadAndGenerateQuiz)
	app.Get("/quiz/leaderboard/users", controllers.GetUserLeaderboard)
	app.Get("/quiz/leaderboard/study-groups", controllers.GetStudyGroupLeaderboard)
	app.Get("/quizes/:id", controllers.GetQuizDetail)

	quiz := app.Group("/quiz", middleware.JWTProtected())
	quiz.Post("/upload", controllers.UploadAndGenerateQuiz)
	quiz.Get("/getMyQuiz", controllers.GetQuizByUser)
	quiz.Get("/feed", controllers.GetFeed)
	quiz.Post("/attempt", controllers.AttemptQuiz)
	quiz.Get("/attempts", controllers.GetAttemptHistory)
	quiz.Get("/attempt/:id", controllers.GetAttemptDetail)
}
