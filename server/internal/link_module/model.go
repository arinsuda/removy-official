package link_module

import "time"

const (
	PlatformPlanzi  = "planzi"
	PlatformPopcorn = "popcorn"

	HandoffTTL    = 60 * time.Second
	PendingTTL    = 10 * time.Minute
	CodeTTL       = 10 * time.Minute
	MaxAttempts   = 5
	MethodPasswd  = "local_password"
	MethodEmail   = "email_code"
	ViaProvision  = "provisioned"
	ViaAuto       = "auto"
	ViaPassword   = "password"
	ViaEmailCode  = "email_code"
	CodeConfirmRC = "LINK_CONFIRM_REQUIRED"
)

type AccountLink struct {
	LocalUserID    uint      `gorm:"primaryKey;column:local_user_id" json:"-"`
	RemotePlatform string    `gorm:"primaryKey;column:remote_platform" json:"platform"`
	RemoteUserID   string    `gorm:"column:remote_user_id" json:"remote_user_id"`
	LinkedVia      string    `gorm:"column:linked_via" json:"-"`
	LinkedAt       time.Time `gorm:"column:linked_at" json:"linked_at"`
}

func (AccountLink) TableName() string { return "account_links" }

type HandoffToken struct {
	TokenHash      string `gorm:"primaryKey;column:token_hash"`
	UserID         uint
	TargetPlatform string
	ExpiresAt      time.Time
	UsedAt         *time.Time
	CreatedAt      time.Time
}

func (HandoffToken) TableName() string { return "link_handoff_tokens" }

type LinkPending struct {
	ID             string `gorm:"primaryKey;type:uuid"`
	LocalUserID    uint
	RemotePlatform string
	RemoteUserID   string
	RemoteEmail    string
	Attempts       int
	CodeHash       *string
	CodeExpiresAt  *time.Time
	CodeEmail      *string // address the code was mailed to; the code only counts while it is still the user's email
	ExpiresAt      time.Time
	CreatedAt      time.Time
}

func (LinkPending) TableName() string { return "link_pending" }

// RemoteUser is the shared user shape returned by internal endpoints.
type RemoteUser struct {
	Platform      string `json:"platform"`
	UserID        string `json:"user_id"`
	Email         string `json:"email"`
	DisplayName   string `json:"display_name"`
	EmailVerified bool   `json:"email_verified"`
	PasswordHash  string `json:"password_hash"`
}
