package controllers

import (
	"github.com/gilanghuda/backend-Quizzo/app/models"
	"github.com/gilanghuda/backend-Quizzo/app/queries"
	"github.com/gilanghuda/backend-Quizzo/pkg/database"
	"github.com/gilanghuda/backend-Quizzo/pkg/utils"
	"github.com/gofiber/fiber/v2"
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

	return c.JSON(fiber.Map{"study_group": created})
}
