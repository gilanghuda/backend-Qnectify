package controllers

import (
	"strconv"

	"github.com/gilanghuda/backend-Quizzo/app/models"
	"github.com/gilanghuda/backend-Quizzo/app/queries"
	"github.com/gilanghuda/backend-Quizzo/pkg/database"
	"github.com/gilanghuda/backend-Quizzo/pkg/utils"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
)

func CreateStudyGroup(c *fiber.Ctx) error {
	userID, err := utils.ExtractUserID(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	studyGroup := &models.StudyGroup{}
	if err := c.BodyParser(studyGroup); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	studyGroup.CreatedBy = userID

	if err := validate.Struct(studyGroup); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	studyGroupQueries := queries.StudyGroupQueries{DB: database.DB}
	created, err := studyGroupQueries.CreateStudyGroup(studyGroup)
	if err != nil {
		log.Error().Err(err).Msg("Failed to create study group")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to create study group",
		})
	}

	if err := studyGroupQueries.JoinStudyGroup(created.ID, userID); err != nil {
		log.Error().Err(err).Str("study_group_id", created.ID.String()).Msg("Failed to add creator as member")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to add creator as member"})
	}

	log.Info().Str("study_group_id", created.ID.String()).Str("created_by", userID.String()).Msg("study group created and creator joined")
	return c.JSON(fiber.Map{"study_group": created})
}

func GetStudyGroup(c *fiber.Ctx) error {
	idStr := c.Params("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid id"})
	}
	q := queries.StudyGroupQueries{DB: database.DB}
	sg, err := q.GetStudyGroup(id)
	if err != nil {
		log.Error().Err(err).Str("study_group_id", id.String()).Msg("failed to get study group")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to get study group"})
	}
	if sg == nil {
		log.Info().Str("study_group_id", id.String()).Msg("study group not found")
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "study group not found"})
	}
	log.Info().Str("study_group_id", id.String()).Msg("study group retrieved")
	return c.JSON(fiber.Map{"study_group": sg})
}

func UpdateStudyGroup(c *fiber.Ctx) error {
	idStr := c.Params("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid id"})
	}
	var sg models.StudyGroup
	if err := c.BodyParser(&sg); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid body"})
	}
	sg.ID = id
	q := queries.StudyGroupQueries{DB: database.DB}
	if err := q.UpdateStudyGroup(&sg); err != nil {
		log.Error().Err(err).Str("study_group_id", id.String()).Msg("failed to update study group")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to update study group"})
	}
	log.Info().Str("study_group_id", id.String()).Msg("study group updated")
	return c.JSON(fiber.Map{"message": "updated"})
}

func DeleteStudyGroup(c *fiber.Ctx) error {
	idStr := c.Params("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid id"})
	}
	q := queries.StudyGroupQueries{DB: database.DB}
	if err := q.DeleteStudyGroup(id); err != nil {
		log.Error().Err(err).Str("study_group_id", id.String()).Msg("failed to delete study group")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to delete study group"})
	}
	log.Info().Str("study_group_id", id.String()).Msg("study group deleted")
	return c.JSON(fiber.Map{"message": "deleted"})
}

func JoinStudyGroup(c *fiber.Ctx) error {
	userID, err := utils.ExtractUserID(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": err.Error()})
	}
	var req struct {
		InvitationCode string `json:"invitation_code"`
	}
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid body"})
	}
	if req.InvitationCode == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invitation_code is required"})
	}
	q := queries.StudyGroupQueries{DB: database.DB}
	group, err := q.GetStudyGroupByInviteCode(req.InvitationCode)
	if err != nil {
		log.Error().Err(err).Str("invitation_code", req.InvitationCode).Msg("failed to lookup study group")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to lookup study group"})
	}
	if group == nil {
		log.Info().Str("invitation_code", req.InvitationCode).Msg("study group not found for invite code")
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "study group not found or invalid invitation code"})
	}
	if group.IsPrivate {
		// invite code matched so allow join
	}
	if err := q.JoinStudyGroup(group.ID, userID); err != nil {
		log.Error().Err(err).Str("study_group_id", group.ID.String()).Str("user_id", userID.String()).Msg("failed to join study group")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to join study group"})
	}
	log.Info().Str("study_group_id", group.ID.String()).Str("user_id", userID.String()).Msg("user joined study group")
	return c.JSON(fiber.Map{"message": "joined"})
}

func GetAllStudyGroups(c *fiber.Ctx) error {
	limit := 20
	offset := 0
	if l := c.Query("limit"); l != "" {
		if v, err := strconv.Atoi(l); err == nil {
			limit = v
		}
	}
	if o := c.Query("offset"); o != "" {
		if v, err := strconv.Atoi(o); err == nil {
			offset = v
		}
	}

	q := queries.StudyGroupQueries{DB: database.DB}
	res, err := q.GetAllStudyGroups(limit, offset)
	if err != nil {
		log.Error().Err(err).Msg("failed to get study groups")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to get study groups"})
	}
	log.Info().Int("count", len(res)).Msg("study groups retrieved")
	return c.JSON(fiber.Map{"study_groups": res})
}

func GetUserStudyGroups(c *fiber.Ctx) error {
	userID, err := utils.ExtractUserID(c)
	if err != nil {
		log.Error().Err(err).Msg("error extracting user ID")
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": err.Error()})
	}

	limit := 20
	offset := 0
	if l := c.Query("limit"); l != "" {
		if v, err := strconv.Atoi(l); err == nil {
			limit = v
		}
	}
	if o := c.Query("offset"); o != "" {
		if v, err := strconv.Atoi(o); err == nil {
			offset = v
		}
	}

	q := queries.StudyGroupQueries{DB: database.DB}
	res, err := q.GetStudyGroupsForUser(userID, limit, offset)
	if err != nil {
		log.Error().Err(err).Str("user_id", userID.String()).Msg("failed to get study groups for user")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to get study groups for user"})
	}
	log.Info().Str("user_id", userID.String()).Int("count", len(res)).Msg("study groups for user retrieved")
	return c.JSON(fiber.Map{"study_groups": res})
}

func GetStudyGroupDetail(c *fiber.Ctx) error {
	idStr := c.Params("id")
	if idStr == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "group id required"})
	}
	id, err := uuid.Parse(idStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid group id"})
	}
	q := queries.StudyGroupQueries{DB: database.DB}
	detail, err := q.GetStudyGroupDetail(id)
	if err != nil {
		log.Error().Err(err).Str("study_group_id", id.String()).Msg("failed to get study group detail")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to get study group detail"})
	}
	if detail == nil {
		log.Info().Str("study_group_id", id.String()).Msg("study group detail not found")
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "study group not found"})
	}
	log.Info().Str("study_group_id", id.String()).Msg("study group detail retrieved")
	return c.JSON(fiber.Map{"detail": detail})
}
