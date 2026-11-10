package notification_module

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
)

// NotificationParams is the structured, language-neutral payload of a notification.
// The client builds the text from type + params; the server stores no copy.
// It must never hold an email, a review body or a review owner.
type NotificationParams struct {
	MediaID     uint   `json:"media_id,omitempty"`
	MediaType   string `json:"media_type,omitempty"`
	MediaTitle  string `json:"media_title,omitempty"` // snapshot at event time
	ActionType  string `json:"action_type,omitempty"`
	TargetCount int    `json:"target_count,omitempty"`
	Exp         int    `json:"exp,omitempty"`
	Name        string `json:"name,omitempty"`
}

// Value stores nil as SQL NULL and anything else as jsonb.
func (p *NotificationParams) Value() (driver.Value, error) {
	if p == nil {
		return nil, nil
	}
	return json.Marshal(p)
}

func (p *NotificationParams) Scan(src any) error {
	var raw []byte
	switch v := src.(type) {
	case nil:
		return nil
	case []byte:
		raw = v
	case string:
		raw = []byte(v)
	default:
		return fmt.Errorf("notification params: unsupported type %T", src)
	}
	return json.Unmarshal(raw, p)
}

// MediaRef identifies a title in media notifications.
type MediaRef struct {
	ID    uint
	Type  string
	Title string
}

func (m MediaRef) params() *NotificationParams {
	return &NotificationParams{MediaID: m.ID, MediaType: m.Type, MediaTitle: m.Title}
}

// AchievementRef is the achievement data a notification needs (kept local to avoid an import).
type AchievementRef struct {
	ID          uint
	ActionType  string
	TargetCount int
	Name        string
}
