package user_module

import "time"

type UserProfileResponse struct {
	ID              uint       `json:"id"`
	Username        string     `json:"username"`
	Email           string     `json:"email"`
	VerifiedEmailAt *time.Time `json:"verified_email_at"`
	DisplayName     *string    `json:"display_name"`
	Bio             *string    `json:"bio"`
	AvatarURL       *string    `json:"avatar_url"`
	DateOfBirth     *time.Time `json:"date_of_birth"`
	Gender          GenderType `json:"gender"`
	FavoriteGenres  *string    `json:"favorite_genres"`
	ReviewCount     int        `json:"review_count"`
	FollowerCount   int        `json:"follower_count"`
	FollowingCount  int        `json:"following_count"`
	IsPrivate       bool       `json:"is_private"`
	Level           int        `json:"level"`
	Role            string     `json:"role"`
	CreatedAt       time.Time  `json:"created_at"`
}

type UpdateProfileRequest struct {
	DisplayName    *string    `json:"display_name"`
	Bio            *string    `json:"bio"`
	AvatarURL      *string    `json:"avatar_url"`
	DateOfBirth    *time.Time `json:"date_of_birth"`
	Gender         GenderType `json:"gender"`
	GenderOther    *string    `json:"gender_other"`
	FavoriteGenres *string    `json:"favorite_genres"`
	IsPrivate      *bool      `json:"is_private"`
}

type UpdateFavoriteGenresRequest struct {
	FavoriteGenres []int `json:"favorite_genres"`
}

type UserSummaryResponse struct {
	ID          uint    `json:"id"`
	Username    string  `json:"username"`
	DisplayName *string `json:"display_name"`
	AvatarURL   *string `json:"avatar_url"`
}

type RequestEmailChangeRequest struct {
	NewEmail string `json:"new_email"`
}

type VerifyEmailChangeRequest struct {
	OTP         string `json:"otp"`           // code mailed to the current address
	NewEmailOTP string `json:"new_email_otp"` // code mailed to the new address
}

type ChangePasswordRequest struct {
	OldPassword     string `json:"old_password"`
	NewPassword     string `json:"new_password"`
	ConfirmPassword string `json:"confirm_password"`
}

type ForgotPasswordRequest struct {
	Email string `json:"email"`
}

type ResetPasswordRequest struct {
	Token           string `json:"token"`
	UserID          uint   `json:"user_id"`
	NewPassword     string `json:"new_password"`
	ConfirmPassword string `json:"confirm_password"`
}
