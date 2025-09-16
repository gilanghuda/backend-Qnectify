package utils

import (
	"errors"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

func ExtractUserID(c *fiber.Ctx) (uuid.UUID, error) {
	claims := c.Locals("user")
	var mapClaims map[string]interface{}

	switch v := claims.(type) {
	case map[string]interface{}:
		mapClaims = v
	case jwt.MapClaims:
		mapClaims = map[string]interface{}(v)
	default:
		return uuid.Nil, errors.New("invalid token claims")
	}

	userIDStr, ok := mapClaims["user_id"].(string)
	if !ok {
		return uuid.Nil, errors.New("invalid user id in token")
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return uuid.Nil, errors.New("invalid user id format")
	}

	return userID, nil
}
