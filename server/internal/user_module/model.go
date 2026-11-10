package user_module

import (
	"time"

	"gorm.io/gorm"
)

type RoleName string

const (
	RoleAdmin RoleName = "admin"
	RoleUser  RoleName = "user"
)

type Role struct {
	ID        uint `gorm:"primarykey;autoIncrement"`
	CreatedAt time.Time
	UpdatedAt time.Time
	RoleName  RoleName `gorm:"type:varchar(50);uniqueIndex;not null"`
}

type GenderType string

const (
	GenderMale   GenderType = "male"
	GenderFemale GenderType = "female"
	GenderOther  GenderType = "other"
)

type AuthProvider string

const (
	ProviderGoogle AuthProvider = "google"
)

type UserIdentity struct {
	ID              uint         `gorm:"primarykey;autoIncrement"`
	UserID          uint         `gorm:"not null;index"`
	User            User         `gorm:"foreignKey:UserID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
	Provider        AuthProvider `gorm:"type:varchar(50);not null"`
	ProviderSubject string       `gorm:"type:varchar(255);not null"`
	ProviderEmail   string       `gorm:"type:varchar(255);not null"`
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

type User struct {
	gorm.Model
	Username        string  `gorm:"type:varchar(50);index:idx_users_active_username,unique,where:deleted_at IS NULL;not null"`
	Password        *string `gorm:"type:varchar(255);default:null"`
	Email           string  `gorm:"type:varchar(100);index:idx_users_active_email,unique,where:deleted_at IS NULL;not null"`
	VerifiedEmailAt *time.Time
	DisplayName     *string `gorm:"type:varchar(100)"`
	Bio             *string `gorm:"type:text"`
	AvatarURL       *string `gorm:"type:varchar(255)"`
	DateOfBirth     *time.Time
	Gender          GenderType `gorm:"type:varchar(20)"`
	GenderOther     *string    `gorm:"type:varchar(100)"`
	RoleID          uint       `gorm:"not null;default:2"`
	Role            Role       `gorm:"foreignKey:RoleID"`
	FavoriteGenres  *string    `gorm:"type:text"`
	IsPrivate       bool       `gorm:"default:false"`
	IsActive        bool       `gorm:"default:true"`
	FirstLoginAt    *time.Time `gorm:"default:null"`
	// OnboardedAt is set once, when the user saves or skips favorite genres.
	OnboardedAt *time.Time
	// ErasureDueAt is set when the user asks for permanent erasure: the purge worker removes the
	// account once it passes. -:migration: migration 20261112_users_erasure_due_at adds it.
	ErasureDueAt *time.Time `gorm:"-:migration"`

	// Read-only for GORM: written only through the identity policy (raw SQL).
	// -:migration keeps AutoMigrate from adding them; migration 20261107_account_security_columns does.
	TokenVersion    int        `gorm:"->;-:migration;column:token_version"`
	IdentityResetAt *time.Time `gorm:"->;-:migration;column:identity_reset_at"`
}

type EmailVerification struct {
	ID        uint      `gorm:"primarykey;autoIncrement"`
	UserID    uint      `gorm:"not null;index"`
	User      User      `gorm:"foreignKey:UserID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
	Token     string    `gorm:"type:varchar(64);uniqueIndex;not null"`
	ExpiresAt time.Time `gorm:"not null"`
	CreatedAt time.Time
	// Email is the address the token was mailed to; NULL on rows from before the column existed.
	Email *string `gorm:"-:migration;type:varchar(100)"`
}

func (e *EmailVerification) IsExpired() bool {
	return time.Now().After(e.ExpiresAt)
}

type RefreshToken struct {
	ID          uint      `gorm:"primarykey;autoIncrement"`
	UserID      uint      `gorm:"not null;index"`
	User        User      `gorm:"foreignKey:UserID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
	HashedToken string    `gorm:"type:varchar(255);uniqueIndex;not null"`
	ExpiresAt   time.Time `gorm:"not null"`
	CreatedAt   time.Time
	UserAgent   string `gorm:"type:varchar(255)"`
	IPAddress   string `gorm:"type:varchar(45)"`
}

func (r *RefreshToken) IsExpired() bool {
	return time.Now().After(r.ExpiresAt)
}

type EmailChangeRequest struct {
	ID           uint      `gorm:"primarykey;autoIncrement"`
	UserID       uint      `gorm:"not null;uniqueIndex"`
	User         User      `gorm:"foreignKey:UserID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
	NewEmail     string    `gorm:"type:varchar(100)"`
	OTPHash      string    `gorm:"type:varchar(255);not null"`
	ExpiresAt    time.Time `gorm:"not null"`
	AttemptCount int       `gorm:"default:0"`
	CreatedAt    time.Time
	// NewOTPHash proves the new address; NULL on rows created before two-OTP email change.
	NewOTPHash *string `gorm:"-:migration;type:varchar(255)"`
}

func (e *EmailChangeRequest) IsExpired() bool {
	return time.Now().After(e.ExpiresAt)
}

type PasswordResetToken struct {
	ID          uint      `gorm:"primarykey;autoIncrement"`
	UserID      uint      `gorm:"not null;uniqueIndex"`
	User        User      `gorm:"foreignKey:UserID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
	HashedToken string    `gorm:"type:varchar(255);not null"`
	ExpiresAt   time.Time `gorm:"not null"`
	CreatedAt   time.Time
	// Email is the address the token was mailed to; NULL on rows from before the column existed.
	Email *string `gorm:"-:migration;type:varchar(100)"`
}

func (p *PasswordResetToken) IsExpired() bool {
	return time.Now().After(p.ExpiresAt)
}

func (u *User) GetAge() *int {
	if u.DateOfBirth == nil {
		return nil
	}

	now := time.Now()
	age := now.Year() - u.DateOfBirth.Year()
	if now.YearDay() < u.DateOfBirth.YearDay() {
		age--
	}

	return &age
}

func isValidGender(g GenderType) bool {
	switch g {
	case GenderMale, GenderFemale, GenderOther:
		return true
	default:
		return false
	}
}
