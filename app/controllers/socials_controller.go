package controllers

import (
	"strconv"

	"github.com/gilanghuda/backend-Quizzo/app/queries"
	"github.com/gilanghuda/backend-Quizzo/pkg/database"
	"github.com/gilanghuda/backend-Quizzo/pkg/utils"
	"github.com/gofiber/fiber/v2"
	"github.com/rs/zerolog/log"
)

func ToggleLike(c *fiber.Ctx) error {
	userID, err := utils.ExtractUserID(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": err.Error()})
	}
	quizID := c.Params("id")
	if quizID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "quiz id required"})
	}
	q := queries.SocialsQueries{DB: database.DB}
	liked, err := q.ToggleLike(quizID, userID.String())
	if err != nil {
		log.Error().Err(err).Str("quiz_id", quizID).Str("user_id", userID.String()).Msg("ToggleLike error")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to toggle like"})
	}
	count, _ := q.CountLikes(quizID)
	log.Info().Str("quiz_id", quizID).Str("user_id", userID.String()).Bool("liked", liked).Int("likes_count", count).Msg("toggle like result")
	return c.JSON(fiber.Map{"liked": liked, "likes_count": count})
}

func GetLikesCount(c *fiber.Ctx) error {
	quizID := c.Params("id")
	if quizID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "quiz id required"})
	}
	q := queries.SocialsQueries{DB: database.DB}
	count, err := q.CountLikes(quizID)
	if err != nil {
		log.Error().Err(err).Str("quiz_id", quizID).Msg("CountLikes error")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to get likes count"})
	}
	log.Info().Str("quiz_id", quizID).Int("likes_count", count).Msg("likes count retrieved")
	return c.JSON(fiber.Map{"likes_count": count})
}

func AddComment(c *fiber.Ctx) error {
	userID, err := utils.ExtractUserID(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": err.Error()})
	}
	quizID := c.Params("id")
	if quizID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "quiz id required"})
	}
	var req struct {
		Content string `json:"content"`
	}
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid body"})
	}
	if req.Content == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "content required"})
	}
	q := queries.SocialsQueries{DB: database.DB}
	commentID, err := q.AddComment(quizID, userID.String(), req.Content)
	if err != nil {
		log.Error().Err(err).Str("quiz_id", quizID).Str("user_id", userID.String()).Msg("AddComment error")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to add comment"})
	}
	log.Info().Str("quiz_id", quizID).Str("user_id", userID.String()).Str("comment_id", commentID).Msg("comment added")
	return c.JSON(fiber.Map{"comment_id": commentID})
}

func DeleteComment(c *fiber.Ctx) error {
	userID, err := utils.ExtractUserID(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": err.Error()})
	}
	commentID := c.Params("id")
	if commentID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "comment id required"})
	}
	q := queries.SocialsQueries{DB: database.DB}
	if err := q.DeleteComment(commentID, userID.String()); err != nil {
		log.Error().Err(err).Str("comment_id", commentID).Str("user_id", userID.String()).Msg("DeleteComment error")
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}
	log.Info().Str("comment_id", commentID).Str("user_id", userID.String()).Msg("comment deleted")
	return c.JSON(fiber.Map{"message": "deleted"})
}

func GetComments(c *fiber.Ctx) error {
	quizID := c.Params("id")
	if quizID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "quiz id required"})
	}
	limit := 20
	if l := c.Query("limit"); l != "" {
		if v, err := strconv.Atoi(l); err == nil {
			limit = v
		}
	}
	q := queries.SocialsQueries{DB: database.DB}
	comments, err := q.GetComments(quizID, limit)
	if err != nil {
		log.Error().Err(err).Str("quiz_id", quizID).Msg("GetComments error")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to get comments"})
	}
	log.Info().Str("quiz_id", quizID).Int("count", len(comments)).Msg("comments retrieved")
	return c.JSON(fiber.Map{"comments": comments})
}
