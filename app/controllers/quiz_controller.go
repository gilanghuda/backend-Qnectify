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
	"github.com/google/uuid"
)

func UploadAndGenerateQuiz(c *fiber.Ctx) error {
	numQuestionsStr := c.FormValue("num_questions")
	difficulty := c.FormValue("difficulty")
	description := c.FormValue("description")
	timeLimitStr := c.FormValue("time_limit")

	if numQuestionsStr == "" || difficulty == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "num_questions and difficulty are required",
		})
	}

	numQuestions, _ := strconv.Atoi(numQuestionsStr)

	var timeLimit *int
	if timeLimitStr != "" {
		if tl, err := strconv.Atoi(timeLimitStr); err == nil {
			timeLimit = &tl
		} else {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid time_limit"})
		}
	}

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
			println("Type assertion to *models.Quiz failed or is nil", qptr)
			println("quizIntf type:", fmt.Sprintf("%T", quizIntf))
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
	quizID, err := quizQueries.InsertQuiz(quiz, userID, description, timeLimit)
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

func GetFeed(c *fiber.Ctx) error {
	userID, err := utils.ExtractUserID(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": err.Error()})
	}

	q := queries.QuizQueries{DB: database.DB}
	feed, err := q.GetQuizzesFromFollowing(userID.String())
	if err != nil {
		log.Printf("GetQuizzesFromFollowing error: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to get feed"})
	}

	return c.JSON(fiber.Map{"feed": feed})
}

func AttemptQuiz(c *fiber.Ctx) error {
	claims := c.Locals("user")
	var mapClaims map[string]interface{}

	switch v := claims.(type) {
	case map[string]interface{}:
		mapClaims = v
	case jwt.MapClaims:
		mapClaims = map[string]interface{}(v)
	default:
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Invalid token claims"})
	}

	userID, ok := mapClaims["user_id"].(string)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Invalid user id in token"})
	}

	var req struct {
		QuizID  string            `json:"quiz_id"`
		Answers map[string]string `json:"answers"`
	}
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request body"})
	}

	q := queries.QuizQueries{DB: database.DB}
	attempted, err := q.HasUserAttemptedQuiz(req.QuizID, userID)
	if err != nil {
		log.Printf("HasUserAttemptedQuiz error: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to check previous attempts"})
	}
	if attempted {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "user already attempted this quiz"})
	}
	score, totalQuestions, err := q.EvaluateQuizAttempt(req.QuizID, req.Answers)
	if err != nil {
		log.Printf("EvaluateQuizAttempt error: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to evaluate attempt"})
	}

	attemptID, err := q.InsertQuizAttempt(req.QuizID, userID, score, totalQuestions, true, req.Answers)
	if err != nil {
		log.Printf("InsertQuizAttempt error: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to save attempt"})
	}

	return c.JSON(fiber.Map{"attempt_id": attemptID, "score": score, "total_questions": totalQuestions})
}

func GetAttemptHistory(c *fiber.Ctx) error {
	userID, err := utils.ExtractUserID(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": err.Error()})
	}

	quizIDParam := c.Query("quiz_id")
	var quizID *uuid.UUID
	if quizIDParam != "" {
		parsedID, err := uuid.Parse(quizIDParam)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid quiz_id"})
		}
		quizID = &parsedID
	}

	limit := 50
	if l := c.Query("limit"); l != "" {
		if v, err := strconv.Atoi(l); err == nil {
			limit = v
		}
	}

	q := queries.QuizQueries{DB: database.DB}
	attempts, err := q.GetAttemptsForUser(userID.String(), quizID, limit)
	if err != nil {
		log.Printf("GetAttemptsForUser error: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to get attempts"})
	}

	return c.JSON(fiber.Map{"attempts": attempts})
}

func GetQuizDetail(c *fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "quiz id is required"})
	}
	q := queries.QuizQueries{DB: database.DB}
	quiz, err := q.GetQuizByID(id)
	if err != nil {
		log.Printf("GetQuizByID error: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to get quiz"})
	}
	if quiz == nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "quiz not found"})
	}
	return c.JSON(fiber.Map{"quiz": quiz})
}

func GetUserLeaderboard(c *fiber.Ctx) error {
	limit := 10
	if l := c.Query("limit"); l != "" {
		if v, err := strconv.Atoi(l); err == nil {
			limit = v
		}
	}
	q := queries.QuizQueries{DB: database.DB}
	res, err := q.GetUserLeaderboard(limit)
	if err != nil {
		log.Printf("GetUserLeaderboard error: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to get leaderboard"})
	}
	return c.JSON(fiber.Map{"leaderboard": res})
}

func GetStudyGroupLeaderboard(c *fiber.Ctx) error {
	limit := 10
	if l := c.Query("limit"); l != "" {
		if v, err := strconv.Atoi(l); err == nil {
			limit = v
		}
	}
	q := queries.QuizQueries{DB: database.DB}
	res, err := q.GetStudyGroupLeaderboard(limit)
	if err != nil {
		log.Printf("GetStudyGroupLeaderboard error: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to get leaderboard"})
	}
	return c.JSON(fiber.Map{"leaderboard": res})
}

func GetAttemptDetail(c *fiber.Ctx) error {
	userID, err := utils.ExtractUserID(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": err.Error()})
	}

	attemptID := c.Params("id")
	if attemptID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "attempt id is required"})
	}

	q := queries.QuizQueries{DB: database.DB}
	detail, err := q.GetAttemptDetail(attemptID)
	if err != nil {
		log.Printf("GetAttemptDetail error: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to get attempt detail"})
	}
	if detail == nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "attempt not found"})
	}

	if detail.Attempt.UserID != userID {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "forbidden"})
	}

	answerMap := map[string]string{}
	for _, a := range detail.Answers {
		answerMap[a.QuestionID.String()] = a.SelectedOptionID.String()
	}

	questionsOut := make([]map[string]string, 0, len(detail.Quiz.Questions))
	for _, qn := range detail.Quiz.Questions {
		qID := qn.ID.String()
		myAnsID := ""
		if v, ok := answerMap[qID]; ok {
			myAnsID = v
		}

		optContent := map[string]string{}
		correctAnsContent := ""
		for _, opt := range qn.Options {
			optContent[opt.ID.String()] = opt.Content
			if opt.IsCorrect {
				correctAnsContent = opt.Content
			}
		}

		myAnsContent := ""
		if myAnsID != "" {
			if c, ok := optContent[myAnsID]; ok {
				myAnsContent = c
			}
		}

		questionsOut = append(questionsOut, map[string]string{
			"id":             qID,
			"quiz_id":        qn.QuizID.String(),
			"question_text":  qn.Question,
			"my_answer":      myAnsContent,
			"correct_answer": correctAnsContent,
			"explanation":    qn.Explanation,
		})
	}

	resp := fiber.Map{
		"attempt": detail.Attempt,
		"quiz_id": detail.Quiz.ID.String(),
		"title":   detail.Quiz.Title,
		"time_limit": func() interface{} {
			if detail.Quiz.TimeLimit.Valid {
				return detail.Quiz.TimeLimit.Int64
			}
			return nil
		}(),
		"total_correct": detail.TotalCorrect,
		"total_questions": func() int {
			if detail.Attempt.TotalQuestions > 0 {
				return detail.Attempt.TotalQuestions
			}
			return len(detail.Quiz.Questions)
		}(),
		"questions": questionsOut,
	}

	return c.JSON(resp)
}
