package controllers

import (
	"fmt"
	"log"

	"github.com/gilanghuda/backend-Quizzo/pkg/utils"
	"github.com/gofiber/fiber/v2"
)

func UploadAndGenerateQuiz(c *fiber.Ctx) error {
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

	content, err := utils.ExtractContent(file, fileHeader)
	if err != nil {
		log.Printf("Error extracting content: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": fmt.Sprintf("failed to extract content: %v", err),
		})
	}

	quiz, err := utils.GenerateQuizFromContent([]byte(content), 5, "medium")
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to generate quiz",
		})
	}

	return c.JSON(fiber.Map{
		"quiz": quiz,
	})
}
