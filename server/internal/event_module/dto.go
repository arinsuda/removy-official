package event_module

import "time"

type LocalizedText struct {
	EN string `json:"en"`
	TH string `json:"th"`
}

// NomineeResponse is a candidate title in a voting round.
// VoteCount and Percentage are hidden while voting is active (nil/omitted).
type NomineeResponse struct {
	ID           uint     `json:"id"`
	EventID      uint     `json:"event_id"`
	MediaID      int      `json:"media_id"`
	MediaType    string   `json:"media_type"`
	Position     int      `json:"position"`
	MediaTitle   string   `json:"media_title"`
	PosterPath   *string  `json:"poster_path"`
	BackdropPath *string  `json:"backdrop_path"`
	Overview     *string  `json:"overview"`
	VoteCount    *int     `json:"vote_count,omitempty"`
	Percentage   *float64 `json:"percentage,omitempty"`
	IsWinner     bool     `json:"is_winner"`
}

// EventResponse is the public shape. Admin-only fields are in AdminEventResponse.
type EventResponse struct {
	ID              uint              `json:"id"`
	Type            EventType         `json:"type"`
	MediaType       *string           `json:"media_type"`
	Period          *string           `json:"period"`
	Title           LocalizedText     `json:"title"`
	Body            LocalizedText     `json:"body"`
	ImageURL        *string           `json:"image_url"`
	LinkURL         *string           `json:"link_url"`
	StartsAt        time.Time         `json:"starts_at"`
	EndsAt          time.Time         `json:"ends_at"`
	ShowPopup       bool              `json:"show_popup"`
	PopupPriority   int               `json:"popup_priority"`
	State           EventState        `json:"state"`
	Nominees        []NomineeResponse `json:"nominees,omitempty"`
	MyVoteNomineeID *uint             `json:"my_vote_nominee_id,omitempty"`
	TotalVotes      *int              `json:"total_votes,omitempty"`
	IsFinalized     bool              `json:"is_finalized"`
}

type AdminEventResponse struct {
	EventResponse
	Status    EventStatus `json:"status"`
	CreatedAt time.Time   `json:"created_at"`
	UpdatedAt time.Time   `json:"updated_at"`
}

type Pagination struct {
	Page       int   `json:"page"`
	Limit      int   `json:"limit"`
	Total      int64 `json:"total"`
	TotalPages int   `json:"total_pages"`
}

func newPagination(page, limit int, total int64) Pagination {
	pages := int((total + int64(limit) - 1) / int64(limit))
	return Pagination{Page: page, Limit: limit, Total: total, TotalPages: pages}
}

// NomineeInput is admin input for one candidate in a voting round.
type NomineeInput struct {
	ID           uint   `json:"id"`
	MediaID      int    `json:"media_id"`
	MediaType    string `json:"media_type"`
	Position     int    `json:"position"`
	MediaTitle   string `json:"media_title"`
	PosterPath   string `json:"poster_path"`
	BackdropPath string `json:"backdrop_path"`
	Overview     string `json:"overview"`
}

// EventInput is the admin create/update body.
type EventInput struct {
	Type          EventType      `json:"type"`
	MediaType     *string        `json:"media_type"`
	Period        *string        `json:"period"`
	TitleEN       string         `json:"title_en"`
	TitleTH       string         `json:"title_th"`
	BodyEN        string         `json:"body_en"`
	BodyTH        string         `json:"body_th"`
	LinkURL       string         `json:"link_url"`
	StartsAt      time.Time      `json:"starts_at"`
	EndsAt        time.Time      `json:"ends_at"`
	ShowPopup     bool           `json:"show_popup"`
	PopupPriority int            `json:"popup_priority"`
	Status        EventStatus    `json:"status"`
	Nominees      []NomineeInput `json:"nominees"`
}

// VoteInput is the body for PUT /events/:id/vote.
type VoteInput struct {
	NomineeID uint `json:"nominee_id"`
}

// SuggestedNominee is returned by GET /admin/events/suggest-nominees.
type SuggestedNominee struct {
	MediaID      int           `json:"media_id"`
	MediaType    string        `json:"media_type"`
	MediaTitle   string        `json:"media_title"`
	PosterPath   string        `json:"poster_path"`
	BackdropPath string        `json:"backdrop_path"`
	Overview     string        `json:"overview"`
	Score        float64       `json:"score"`
	Source       NomineeSource `json:"source"`
}

// GenerateMonthlyInput is the query/body for POST /admin/events/generate-monthly.
type GenerateMonthlyInput struct {
	MediaType string `json:"media_type"` // "movie" | "tv"
	Period    string `json:"period"`     // "YYYY-MM"
}
