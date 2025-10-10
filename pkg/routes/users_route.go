package routes

import (
	"github.com/gilanghuda/backend-Quizzo/app/controllers"
	"github.com/gilanghuda/backend-Quizzo/pkg/middleware"
	"github.com/gofiber/fiber/v2"
)

func RegisterUserRoutes(app *fiber.App) {
	app.Post("/signup", controllers.UserSignUp)
	app.Post("/signin", controllers.UserSignIn)
	app.Post("/signin/google", controllers.UserSignInGoogle)
	app.Post("/logout", controllers.UserLogout)

	user := app.Group("/user", middleware.JWTProtected())
	user.Get("/profile", controllers.UserProfile)
	user.Post("/follow/:id", controllers.FollowUser)
	user.Post("/unfollow/:id", controllers.UnfollowUser)
	user.Get("/recommendations", controllers.RecommendUsers)

	app.Get("/users/:id", controllers.GetUserByID)
}
