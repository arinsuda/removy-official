package event_module

import (
	"time"

	users "github.com/arinsuda/movie-hub/internal/user_module"
	"gorm.io/gorm"
)

type EventType string

const (
	TypeAnnouncement EventType = "announcement"
	TypeVoteMonthly  EventType = "vote_monthly"
	TypeVoteYearly   EventType = "vote_yearly"
)

type EventStatus string

const (
	StatusDraft     EventStatus = "draft"
	StatusPublished EventStatus = "published"
	StatusCancelled EventStatus = "cancelled"
)

func (s EventStatus) Valid() bool {
	return s == StatusDraft || s == StatusPublished || s == StatusCancelled
}

// State is derived from the time window; it is never stored.
type EventState string

const (
	StateUpcoming EventState = "upcoming"
	StateActive   EventState = "active"
	StatePast     EventState = "past"
)

func (s EventState) Valid() bool {
	return s == StateUpcoming || s == StateActive || s == StatePast
}

// Event is an admin-made announcement or (later phases) a vote round.
// Titles and bodies are admin content in both UI languages, so they live here, not in i18n files.
type Event struct {
	ID        uint           `gorm:"primarykey"`
	CreatedAt time.Time      `gorm:"not null"`
	UpdatedAt time.Time      `gorm:"not null"`
	DeletedAt gorm.DeletedAt `gorm:"index"`

	Type      EventType `gorm:"type:varchar(30);not null;index"`
	MediaType *string   `gorm:"type:varchar(10)"`
	Period    *string   `gorm:"type:varchar(7)"`

	TitleEN string `gorm:"type:varchar(120);not null"`
	TitleTH string `gorm:"type:varchar(120);not null"`
	BodyEN  string `gorm:"type:text;not null;default:''"`
	BodyTH  string `gorm:"type:text;not null;default:''"`

	// ImageKey is a MinIO object key under eventImageRoot; the API returns a presigned URL.
	ImageKey string `gorm:"type:varchar(255);not null;default:''"`
	LinkURL  string `gorm:"type:varchar(500);not null;default:''"`

	StartsAt time.Time `gorm:"not null;index"`
	EndsAt   time.Time `gorm:"not null;index"`

	ShowPopup     bool        `gorm:"not null;default:false"`
	PopupPriority int         `gorm:"not null;default:0"`
	Status        EventStatus `gorm:"type:varchar(20);not null;default:'draft';index"`

	// CreatedBy becomes NULL when the admin is hard-deleted; the event outlives them.
	CreatedBy *uint       `gorm:"index"`
	Creator   *users.User `gorm:"foreignKey:CreatedBy;constraint:OnUpdate:CASCADE,OnDelete:SET NULL;"`
}

func (Event) TableName() string { return "events" }

// StateAt reports where now falls in the event window.
func (e *Event) StateAt(now time.Time) EventState {
	switch {
	case now.Before(e.StartsAt):
		return StateUpcoming
	case now.Before(e.EndsAt):
		return StateActive
	default:
		return StatePast
	}
}
