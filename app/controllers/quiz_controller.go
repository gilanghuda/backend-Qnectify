package controllers

import (
	"bytes"
	"fmt"
	"io"
	"strconv"

	"github.com/gilanghuda/backend-Quizzo/app/models"
	"github.com/gilanghuda/backend-Quizzo/app/queries"
	"github.com/gilanghuda/backend-Quizzo/pkg/database"
	"github.com/gilanghuda/backend-Quizzo/pkg/utils"
	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
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

	// Read the uploaded file into memory so we can both upload the original PDF
	// to the external service and pass a reader to GenerateQuiz.
	fileBytes, err := io.ReadAll(file)
	if err != nil {
		log.Error().Err(err).Msg("failed to read uploaded file")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to read uploaded file"})
	}

	// First generate quiz from the uploaded file
	quizIntf, err := utils.GenerateQuiz(bytes.NewReader(fileBytes), numQuestions, difficulty)
	if err != nil {
		log.Error().Err(err).Msg("error generating quiz from file")
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
		log.Error().Err(err).Msg("failed to begin tx")
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
		log.Error().Err(err).Msg("InsertQuiz error")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to insert quiz"})
	}
	log.Info().Str("quiz_id", quizID).Msg("quiz inserted")

	// Use quizID as id_file when saving original PDF
	if err := utils.SaveFile(quizID, fileHeader.Filename, fileBytes); err != nil {
		log.Error().Err(err).Msg("failed to upload original file")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to upload original file"})
	}
	log.Info().Str("quiz_id", quizID).Str("filename", fileHeader.Filename).Msg("original file saved")

	questionIDs, err := quizQueries.InsertQuestionsBulk(quizID, quiz.Questions)
	if err != nil {
		log.Error().Err(err).Msg("InsertQuestionsBulk error")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to insert questions"})
	}
	log.Info().Int("questions_count", len(questionIDs)).Msg("questions inserted")

	if err := quizQueries.InsertOptionsBulk(questionIDs, quiz.Questions); err != nil {
		log.Error().Err(err).Msg("InsertOptionsBulk error")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to insert options"})
	}
	log.Info().Msg("options inserted")

	if err := tx.Commit(); err != nil {
		log.Error().Err(err).Msg("tx commit error")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to commit transaction"})
	}
	committed = true
	log.Info().Str("quiz_id", quizID).Msg("transaction committed, quiz generation complete")

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
		log.Error().Err(err).Msg("GetQuizByUserId error")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to get quiz"})
	}
	log.Info().Str("user_id", userID).Msg("retrieved quizzes for user")

	return c.JSON(fiber.Map{"quiz": quiz})
}

func GetFeed(c *fiber.Ctx) error {
	userID, err := utils.ExtractUserID(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": err.Error()})
	}

	q := queries.QuizQueries{DB: database.DB}
	feed, err := q.GetFeedWithLikes(userID.String())
	if err != nil {
		log.Error().Err(err).Msg("GetFeedWithLikes error")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to get feed"})
	}
	log.Info().Str("user_id", userID.String()).Int("feed_count", len(feed)).Msg("feed retrieved")

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
		log.Error().Err(err).Msg("HasUserAttemptedQuiz error")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to check previous attempts"})
	}
	if attempted {
		log.Info().Str("user_id", userID).Str("quiz_id", req.QuizID).Msg("user already attempted quiz")
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "user already attempted this quiz"})
	}
	score, totalQuestions, err := q.EvaluateQuizAttempt(req.QuizID, req.Answers)
	if err != nil {
		log.Error().Err(err).Msg("EvaluateQuizAttempt error")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to evaluate attempt"})
	}

	attemptID, err := q.InsertQuizAttempt(req.QuizID, userID, score, totalQuestions, true, req.Answers)
	if err != nil {
		log.Error().Err(err).Msg("InsertQuizAttempt error")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to save attempt"})
	}
	log.Info().Str("attempt_id", attemptID).Int("score", score).Int("total_questions", totalQuestions).Msg("quiz attempt recorded")

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
		log.Error().Err(err).Msg("GetAttemptsForUser error")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to get attempts"})
	}
	log.Info().Str("user_id", userID.String()).Int("count", len(attempts)).Msg("attempt history retrieved")

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
		log.Error().Err(err).Msg("GetQuizByID error")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to get quiz"})
	}
	if quiz == nil {
		log.Info().Str("quiz_id", id).Msg("quiz not found")
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "quiz not found"})
	}
	log.Info().Str("quiz_id", id).Msg("quiz detail retrieved")
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
		log.Error().Err(err).Msg("GetUserLeaderboard error")
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
		log.Error().Err(err).Msg("GetStudyGroupLeaderboard error")
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

func GetQuizFile(c *fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "file id is required"})
	}

	data, contentType, err := utils.GetFile(id)
	if err != nil {
		log.Error().Err(err).Msg("GetFile error")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to fetch file"})
	}
	log.Info().Str("file_id", id).Msg("file fetched")

	if contentType == "" {
		contentType = "application/pdf"
	}

	c.Set("Content-Type", contentType)
	c.Set("Content-Disposition", "inline; filename=\""+id+".pdf\"")
	return c.SendStream(bytes.NewReader(data))
}

func AddQuizToStudyGroup(c *fiber.Ctx) error {
	userID, err := utils.ExtractUserID(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": err.Error()})
	}

	var req struct {
		QuizID       string `json:"quiz_id"`
		StudyGroupID string `json:"study_group_id"`
	}
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request body"})
	}
	if req.QuizID == "" || req.StudyGroupID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "quiz_id and study_group_id are required"})
	}

	q := queries.QuizQueries{DB: database.DB}
	quiz, err := q.GetQuizByID(req.QuizID)
	if err != nil {
		log.Error().Err(err).Msg("GetQuizByID error")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to fetch quiz"})
	}
	if quiz == nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "quiz not found"})
	}

	if quiz.CreatedBy != userID.String() {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "forbidden"})
	}

	if err := q.AssignQuizToStudyGroup(req.QuizID, req.StudyGroupID); err != nil {
		log.Error().Err(err).Msg("AssignQuizToStudyGroup error")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to assign quiz to study group"})
	}
	log.Info().Str("quiz_id", req.QuizID).Str("study_group_id", req.StudyGroupID).Msg("quiz assigned to study group")

	return c.JSON(fiber.Map{"message": "quiz assigned to study group successfully"})
}

func GetQuizzesByStudyGroup(c *fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "study group id is required"})
	}

	limit := 50
	if l := c.Query("limit"); l != "" {
		if v, err := strconv.Atoi(l); err == nil {
			limit = v
		}
	}

	q := queries.QuizQueries{DB: database.DB}
	quizzes, err := q.GetQuizzesByStudyGroup(id, limit)
	if err != nil {
		log.Error().Err(err).Msg("GetQuizzesByStudyGroup error")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to get quizzes for study group"})
	}
	log.Info().Str("study_group_id", id).Int("count", len(quizzes)).Msg("quizzes retrieved for study group")

	return c.JSON(fiber.Map{"quizzes": quizzes})
}
