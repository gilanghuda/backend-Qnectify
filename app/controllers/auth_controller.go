package controllers

import (
	"context"
	"database/sql"
	"log"
	"os"
	"strings"
	"time"

	"github.com/gilanghuda/backend-Quizzo/app/models"
	"github.com/gilanghuda/backend-Quizzo/app/queries"
	"github.com/gilanghuda/backend-Quizzo/pkg/database"
	"github.com/gilanghuda/backend-Quizzo/pkg/utils"
	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

var validate = validator.New()

func UserSignUp(c *fiber.Ctx) error {
	signUp := &models.SignUp{}
	if err := c.BodyParser(signUp); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	if err := validate.Struct(signUp); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	role := signUp.UserRole
	if role == "" {
		role = utils.RoleUser
	}

	valid := false
	for _, r := range utils.ValidRoles {
		if role == r {
			valid = true
			break
		}
	}
	if !valid {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid user role",
		})
	}

	userQueries := queries.UserQueries{DB: database.DB}
	_, err := userQueries.GetUserByEmail(signUp.Email)
	if err == nil {
		log.Println("Attempt to register with existing email:", signUp.Email)
		return c.Status(fiber.StatusConflict).JSON(fiber.Map{
			"error": "Email already registered",
		})
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(signUp.Password), bcrypt.DefaultCost)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to hash password",
		})
	}

	user := &models.User{
		ID:           uuid.New(),
		Email:        signUp.Email,
		Username:     signUp.Username,
		PasswordHash: string(hashedPassword),
		UserRole:     role,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	if err := userQueries.CreateUser(user); err != nil {
		log.Println("Error creating user:", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to create user",
		})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"message": "User registered successfully",
	})
}

func UserSignIn(c *fiber.Ctx) error {
	signIn := &models.SignIn{}
	if err := c.BodyParser(signIn); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	if err := validate.Struct(signIn); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	userQueries := queries.UserQueries{DB: database.DB}
	user, err := userQueries.GetUserByEmail(signIn.Email)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "user didnt exist",
		})
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(signIn.Password)); err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "wrong password",
		})
	}

	secret := os.Getenv("JWT_SECRET")
	claims := jwt.MapClaims{
		"user_id":   user.ID.String(),
		"email":     user.Email,
		"user_role": user.UserRole,
		"exp":       time.Now().Add(time.Hour * 72).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(secret))
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to generate token",
		})
	}

	c.Cookie(&fiber.Cookie{
		Name:     "token",
		Value:    tokenString,
		Expires:  time.Now().Add(time.Hour * 72),
		Path:     "/",
		HTTPOnly: false,
		Secure:   false,
		SameSite: "lax",
	})

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "Sign in successful",
		"user": fiber.Map{
			"id":        user.ID,
			"email":     user.Email,
			"user_role": user.UserRole,
		},
		"token": tokenString,
	})
}

func UserSignInGoogle(c *fiber.Ctx) error {
	var req struct {
		GoogleToken string `json:"id_token"`
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request"})
	}

	if req.GoogleToken == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "google_token is required for this endpoint"})
	}

	email, err := utils.ValidateGoogleIDToken(context.Background(), req.GoogleToken)
	if err != nil {
		log.Printf("google token validation failed: %v", err)
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "invalid google token"})
	}

	username := email
	if parts := strings.Split(email, "@"); len(parts) > 0 {
		username = parts[0]
	}

	uq := queries.UserQueries{DB: database.DB}
	user, err := uq.GetUserByEmail(email)
	if err != nil {
		if err.Error() == "user not found" {
			newUser := models.User{
				ID:           uuid.New(),
				Username:     username,
				Email:        email,
				PasswordHash: "",
				ExpPoints:    "0",
				UserRole:     "user",
				ImageURL:     sql.NullString{Valid: false},
				CreatedAt:    time.Now(),
				UpdatedAt:    time.Now(),
			}
			if err := uq.CreateUser(&newUser); err != nil {
				log.Printf("failed to create user: %v", err)
				return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to create user"})
			}
			user = newUser
		} else {
			log.Printf("GetUserByEmail error: %v", err)
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to get user"})
		}
	}

	user.PasswordHash = ""

	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		secret = "secret"
	}
	claims := jwt.MapClaims{
		"user_id": user.ID.String(),
		"exp":     time.Now().Add(72 * time.Hour).Unix(),
	}
	tok := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenStr, err := tok.SignedString([]byte(secret))
	if err != nil {
		log.Printf("failed to sign token: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to create token"})
	}

	return c.JSON(fiber.Map{"token": tokenStr, "user": user})
}

func UserLogout(c *fiber.Ctx) error {
	c.Cookie(&fiber.Cookie{
		Name:     "token",
		Value:    "",
		Expires:  time.Now().Add(-time.Hour),
		Domain:   "localhost",
		Path:     "/",
		HTTPOnly: true,
		Secure:   false,
		SameSite: "lax",
		MaxAge:   -1,
	})
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "Logout successful",
	})
}
