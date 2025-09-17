package controllers

import (
	"strconv"

	"github.com/gilanghuda/backend-Quizzo/app/models"
	"github.com/gilanghuda/backend-Quizzo/app/queries"
	"github.com/gilanghuda/backend-Quizzo/pkg/database"
	"github.com/gilanghuda/backend-Quizzo/pkg/utils"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
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
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to create study group",
		})
	}

	if err := studyGroupQueries.JoinStudyGroup(created.ID, userID); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to add creator as member"})
	}

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
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to get study group"})
	}
	if sg == nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "study group not found"})
	}
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
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to update study group"})
	}
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
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to delete study group"})
	}
	return c.JSON(fiber.Map{"message": "deleted"})
}

func JoinStudyGroup(c *fiber.Ctx) error {
	userID, err := utils.ExtractUserID(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": err.Error()})
	}
	idStr := c.Params("id")
	groupID, err := uuid.Parse(idStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid group id"})
	}
	var req struct {
		InvitationCode string `json:"invitation_code"`
	}
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid body"})
	}
	q := queries.StudyGroupQueries{DB: database.DB}
	group, err := q.GetStudyGroup(groupID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to get study group"})
	}
	if group == nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "study group not found"})
	}
	if group.IsPrivate {
		if group.InviteCode == nil || req.InvitationCode == "" || req.InvitationCode != *group.InviteCode {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "invalid invitation code"})
		}
	}
	if err := q.JoinStudyGroup(groupID, userID); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to join study group"})
	}
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
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to get study groups"})
	}
	return c.JSON(fiber.Map{"study_groups": res})
}
