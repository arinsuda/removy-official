package event_module

import "time"

type LocalizedText struct {
	EN string `json:"en"`
	TH string `json:"th"`
}

// EventResponse is the public shape. Admin-only fields are in AdminEventResponse.
type EventResponse struct {
	ID            uint          `json:"id"`
	Type          EventType     `json:"type"`
	MediaType     *string       `json:"media_type"`
	Period        *string       `json:"period"`
	Title         LocalizedText `json:"title"`
	Body          LocalizedText `json:"body"`
	ImageURL      *string       `json:"image_url"`
	LinkURL       *string       `json:"link_url"`
	StartsAt      time.Time     `json:"starts_at"`
	EndsAt        time.Time     `json:"ends_at"`
	ShowPopup     bool          `json:"show_popup"`
	PopupPriority int           `json:"popup_priority"`
	State         EventState    `json:"state"`
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

// EventInput is the admin create/update body. PR1 creates announcements only.
type EventInput struct {
	TitleEN       string      `json:"title_en"`
	TitleTH       string      `json:"title_th"`
	BodyEN        string      `json:"body_en"`
	BodyTH        string      `json:"body_th"`
	LinkURL       string      `json:"link_url"`
	StartsAt      time.Time   `json:"starts_at"`
	EndsAt        time.Time   `json:"ends_at"`
	ShowPopup     bool        `json:"show_popup"`
	PopupPriority int         `json:"popup_priority"`
	Status        EventStatus `json:"status"`
}
