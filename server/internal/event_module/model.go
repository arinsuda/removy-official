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

func (t EventType) Valid() bool {
	return t == TypeAnnouncement || t == TypeVoteMonthly || t == TypeVoteYearly
}

func (t EventType) IsVoting() bool {
	return t == TypeVoteMonthly || t == TypeVoteYearly
}

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
	IsFinalized   bool        `gorm:"not null;default:false"`

	// CreatedBy becomes NULL when the admin is hard-deleted; the event outlives them.
	CreatedBy *uint       `gorm:"index"`
	Creator   *users.User `gorm:"foreignKey:CreatedBy;constraint:OnUpdate:CASCADE,OnDelete:SET NULL;"`

	Nominees []EventNominee `gorm:"foreignKey:EventID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
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

// EventNominee represents a candidate title in a voting round.
type EventNominee struct {
	ID           uint      `gorm:"primarykey"`
	CreatedAt    time.Time `gorm:"not null"`
	UpdatedAt    time.Time `gorm:"not null"`
	EventID      uint      `gorm:"not null;index"`
	Event        Event     `gorm:"foreignKey:EventID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
	MediaID      int       `gorm:"not null;index"`
	MediaType    string    `gorm:"type:varchar(10);not null"`
	Position     int       `gorm:"not null;default:0"`
	MediaTitle   string    `gorm:"type:varchar(255);not null"`
	PosterPath   string    `gorm:"type:varchar(255);not null;default:''"`
	BackdropPath string    `gorm:"type:varchar(255);not null;default:''"`
	Overview     string    `gorm:"type:text;not null;default:''"`
	// FinalVotes is the settled tally, written once when the round finalizes (nil while open).
	// Results are read from here, so later changes to voters cannot alter a closed round.
	FinalVotes *int `gorm:"column:final_votes"`
}

func (EventNominee) TableName() string { return "event_nominees" }

// EventVote represents a single user's vote in an event.
// Unique constraint on (event_id, user_id) guarantees 1 vote per user per event.
type EventVote struct {
	EventID   uint         `gorm:"primaryKey"`
	UserID    uint         `gorm:"primaryKey;not null"`
	User      users.User   `gorm:"foreignKey:UserID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
	NomineeID uint         `gorm:"not null;index"`
	Nominee   EventNominee `gorm:"foreignKey:NomineeID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
	CreatedAt time.Time    `gorm:"not null"`
	UpdatedAt time.Time    `gorm:"not null"`
}

func (EventVote) TableName() string { return "event_votes" }

// EventAward stores finalized winners of a voting round.
type EventAward struct {
	ID uint `gorm:"primarykey"`
	// The unique (event_id, nominee_id) index stops a racing finalize from writing a second award set.
	EventID   uint         `gorm:"not null;uniqueIndex:idx_event_awards_event_nominee"`
	Event     Event        `gorm:"foreignKey:EventID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
	NomineeID uint         `gorm:"not null;index;uniqueIndex:idx_event_awards_event_nominee"`
	Nominee   EventNominee `gorm:"foreignKey:NomineeID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
	MediaID   int          `gorm:"not null;index"`
	MediaType string       `gorm:"type:varchar(10);not null"`
	AwardKind AwardKind    `gorm:"type:varchar(20);not null"`
	Period    string       `gorm:"type:varchar(7);not null"` // "YYYY-MM" or "YYYY"
	Votes     int          `gorm:"not null;default:0"`
	CreatedAt time.Time    `gorm:"not null"`
}

func (EventAward) TableName() string { return "event_awards" }
