package shared

import (
	"context"

	"github.com/arinsuda/movie-hub/internal/notification_module"
)

// NotificationTitleCheck is the age gate for notifications: it says, per title, whether a notification about
// it may reach the user (userID 0: safe for everyone). The router wires it into the notification service.
// It lives here because shared imports notification_module, not the other way round.
func NotificationTitleCheck(ctx context.Context, userID uint, titles []notification_module.TitleKey) (map[notification_module.TitleKey]bool, error) {
	keys := make([]MediaKey, len(titles))
	for i, t := range titles {
		keys[i] = MediaKey{ID: t.ID, Type: MediaType(t.Type)}
	}
	ok, err := ReceivableMedia(ctx, userID, keys)
	if err != nil {
		return nil, err
	}
	out := make(map[notification_module.TitleKey]bool, len(titles))
	for i, t := range titles {
		out[t] = ok[keys[i]]
	}
	return out, nil
}

// NotificationMangaRestricted says whether adult and unknown manga must be left out of userID's notification
// counts and pages. Notifications have no owner exemption, so only an eligible viewer sees adult manga.
// The router wires it into the notification service next to NotificationTitleCheck.
func NotificationMangaRestricted(ctx context.Context, userID uint) (bool, error) {
	vis, err := NewMediaVisibility(ctx, userID)
	if err != nil {
		return false, err
	}
	return vis.Restricted(0), nil
}
