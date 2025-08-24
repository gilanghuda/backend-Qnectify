package models

type SignUp struct {
	Email    string `json:"email" validate:"required,email,lte=255"`
	Username string `json:"username" validate:"required,lte=255"`
	Password string `json:"password" validate:"required,lte=255"`
	UserRole string `json:"user_role,omitempty"`
}

type SignIn struct {
	Email    string `json:"email" validate:"required,email,lte=255"`
	Password string `json:"password" validate:"required,lte=255"`
}
