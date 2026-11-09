package auth_module

import "time"

type RegisterRequest struct {
	Username        string `json:"username" validate:"required,min=3,max=50,alphanum"`
	Email           string `json:"email"    validate:"required,email"`
	Password        string `json:"password" validate:"required,min=8"`
	ConfirmPassword string `json:"confirm_password" validate:"required,min=8,eqfield=Password"`
}

type LoginRequest struct {
	Identifier string `json:"identifier" validate:"required"`
	Password   string `json:"password"   validate:"required"`
}

type UserResponse struct {
	ID             uint       `json:"id"`
	Username       string     `json:"username"`
	Email          string     `json:"email"`
	DisplayName    *string    `json:"display_name"`
	Bio            *string    `json:"bio"`
	AvatarURL      *string    `json:"avatar_url"`
	DateOfBirth    *time.Time `json:"date_of_birth"`
	Gender         string     `json:"gender"`
	IsVerified     bool       `json:"is_verified"`
	Role           string     `json:"role"`
	FavoriteGenres *string    `json:"favorite_genres"`
	IsPrivate      bool       `json:"is_private"`
	Level          int        `json:"level"`
	OnboardedAt    *time.Time `json:"onboarded_at"`
}

type AuthResponse struct {
	User         UserResponse `json:"user"`
	AccessToken  string       `json:"access_token,omitempty"`
	RefreshToken string       `json:"refresh_token,omitempty"`
}

type MessageResponse struct {
	Message string `json:"message"`
}
