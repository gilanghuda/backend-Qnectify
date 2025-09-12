package controllers

import (
	"fmt"
	"log"
	"strconv"

	"github.com/gilanghuda/backend-Quizzo/app/models"
	"github.com/gilanghuda/backend-Quizzo/app/queries"
	"github.com/gilanghuda/backend-Quizzo/pkg/database"
	"github.com/gilanghuda/backend-Quizzo/pkg/utils"
	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
)

func UploadAndGenerateQuiz(c *fiber.Ctx) error {
	numQuestionsStr := c.FormValue("num_questions")
	difficulty := c.FormValue("difficulty")
	description := c.FormValue("description")

	if numQuestionsStr == "" || difficulty == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "num_questions and difficulty are required",
		})
	}

	numQuestions, _ := strconv.Atoi(numQuestionsStr)

	claims := c.Locals("user")
	var mapClaims map[string]interface{}

	switch v := claims.(type) {
	case map[string]interface{}:
		mapClaims = v
	case jwt.MapClaims:
		mapClaims = map[string]interface{}(v)
	default:
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Invalid token claims",
		})
	}

	userID, ok := mapClaims["user_id"].(string)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Invalid user id in token",
		})
	}

	fileHeader, err := c.FormFile("file")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "file is required",
		})
	}

	file, err := fileHeader.Open()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to open uploaded file",
		})
	}
	defer file.Close()

	quizIntf, err := utils.GenerateQuiz(file, numQuestions, difficulty)
	if err != nil {
		log.Printf("Error generating quiz from file: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": fmt.Sprintf("failed to generate quiz: %v", err),
		})
	}

	quiz, ok := quizIntf.(models.Quiz)
	if !ok {
		qptr, ok2 := quizIntf.(*models.Quiz)
		if !ok2 || qptr == nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "invalid quiz format"})
		}
		quiz = *qptr
	}

	db := database.DB
	if db == nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "database not initialized"})
	}

	tx, err := db.Begin()
	if err != nil {
		log.Printf("failed to begin tx: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to begin transaction"})
	}

	committed := false
	defer func() {
		if !committed {
			_ = tx.Rollback()
		}
	}()
	quizQueries := queries.QuizQueries{DB: database.DB}
	quizID, err := quizQueries.InsertQuiz(quiz, userID, description)
	if err != nil {
		log.Printf("InsertQuiz error: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to insert quiz"})
	}

	questionIDs, err := quizQueries.InsertQuestionsBulk(quizID, quiz.Questions)
	if err != nil {
		log.Printf("InsertQuestionsBulk error: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to insert questions"})
	}

	if err := quizQueries.InsertOptionsBulk(questionIDs, quiz.Questions); err != nil {
		log.Printf("InsertOptionsBulk error: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to insert options"})
	}

	if err := tx.Commit(); err != nil {
		log.Printf("tx commit error: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to commit transaction"})
	}
	committed = true

	return c.JSON(fiber.Map{"quiz_id": quizID, "message": "quiz generated and saved successfully"})
}

func GetQuizByUser(c *fiber.Ctx) error {
	claims := c.Locals("user")
	var mapClaims map[string]interface{}

	switch v := claims.(type) {
	case map[string]interface{}:
		mapClaims = v
	case jwt.MapClaims:
		mapClaims = map[string]interface{}(v)
	default:
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Invalid token claims",
		})
	}

	userID, ok := mapClaims["user_id"].(string)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Invalid user id in token",
		})
	}
	quizQueries := queries.QuizQueries{DB: database.DB}
	quiz, err := quizQueries.GetQuizByUserId(userID)
	if err != nil {
		log.Printf("GetQuizByUserId error: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to get quiz"})
	}

	return c.JSON(fiber.Map{"quiz": quiz})
}
