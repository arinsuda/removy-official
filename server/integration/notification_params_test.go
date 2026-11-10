//go:build integration

package integration

import (
	"context"
	"encoding/json"
	"net/http/httptest"
	"sort"
	"testing"

	"github.com/arinsuda/movie-hub/internal/notification_module"
	"github.com/arinsuda/movie-hub/internal/user_module"
	"github.com/arinsuda/movie-hub/middleware"
	"gorm.io/gorm"
)

type fakeUserProvider struct {
	actor     user_module.User
	followers []uint
}

func (f fakeUserProvider) FindByID(id uint) (*user_module.User, int, int, int, error) {
	u := f.actor
	u.ID = id
	return &u, 0, 0, 0, nil
}
func (f fakeUserProvider) FindFollowerIDs(uint) ([]uint, error) { return f.followers, nil }

// paramKeysByType is the allow-list of params keys each type may carry.
var mediaKeys = []string{"media_id", "media_title", "media_type"}
var paramKeysByType = map[notification_module.NotificationType][]string{
	notification_module.NotifFollowingReviewed:       mediaKeys,
	notification_module.NotifReviewLiked:             mediaKeys,
	notification_module.NotifFollowingLikedReview:    mediaKeys,
	notification_module.NotifReviewMarkedHelpful:     mediaKeys,
	notification_module.NotifFollowingMarkedHelpful:  mediaKeys,
	notification_module.NotifReviewCommented:         mediaKeys,
	notification_module.NotifFollowingCommented:      mediaKeys,
	notification_module.NotifFollowingAddedWatched:   mediaKeys,
	notification_module.NotifFollowingAddedWatchlist: mediaKeys,
	notification_module.NotifFollowingLikedMedia:     mediaKeys,
	notification_module.NotifMovieNowPlaying:         mediaKeys,
	notification_module.NotifAchievementUnlocked:     {"action_type", "exp", "name", "target_count"},
}

func sortedKeys(m map[string]any) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}

func newNotifService(t *testing.T, db *gorm.DB, fp fakeUserProvider, hub *notification_module.Hub) *notification_module.Service {
	t.Helper()
	if hub == nil {
		hub = notification_module.NewHub(nil, nil)
	}
	return notification_module.NewService(db, fp, hub)
}

func TestNotificationParams_PushStoresTypeAndParams(t *testing.T) {
	e := newEnv(t)
	owner := seedUser(t, e.db, "owner@notif.test", true, "pw-owner-1")
	actor := seedUser(t, e.db, "actor@notif.test", true, "pw-actor-1")
	follower := seedUser(t, e.db, "follower@notif.test", true, "pw-follow-1")
	// Followers only get activity the actor shares: enable every activity type.
	e.db.Exec(`INSERT INTO activity_privacy_settings (user_id, activity_type, enabled, created_at, updated_at)
		SELECT ?, t, true, now(), now() FROM unnest(ARRAY['review_created','review_commented','review_liked','media_liked','watchlist_added','watched_added']) AS t`, actor)

	svc := newNotifService(t, e.db, fakeUserProvider{actor: user_module.User{Username: "actor"}, followers: []uint{follower}}, nil)
	ctx := context.Background()
	media := notification_module.MediaRef{ID: 603, Type: "tv", Title: "The Matrix"}

	pushes := map[notification_module.NotificationType]func() error{
		notification_module.NotifWelcome:                 func() error { return svc.PushWelcome(ctx, owner) },
		notification_module.NotifFollowedYou:             func() error { return svc.PushFollowedYou(ctx, owner, actor) },
		notification_module.NotifFollowRequested:         func() error { return svc.PushFollowRequested(ctx, owner, actor) },
		notification_module.NotifFollowAccepted:          func() error { return svc.PushFollowAccepted(ctx, owner, actor) },
		notification_module.NotifReviewLiked:             func() error { return svc.PushReviewLiked(ctx, owner, actor, 7, media) },
		notification_module.NotifReviewMarkedHelpful:     func() error { return svc.PushReviewMarkedHelpful(ctx, owner, actor, 7, media) },
		notification_module.NotifReviewCommented:         func() error { return svc.PushReviewCommented(ctx, owner, actor, 7, media) },
		notification_module.NotifFollowingReviewed:       func() error { return svc.PushFollowingReviewed(ctx, actor, 7, media) },
		notification_module.NotifFollowingLikedReview:    func() error { return svc.PushFollowingLikedReview(ctx, actor, 7, media) },
		notification_module.NotifFollowingMarkedHelpful:  func() error { return svc.PushFollowingMarkedHelpful(ctx, actor, 7, media) },
		notification_module.NotifFollowingCommented:      func() error { return svc.PushFollowingCommented(ctx, actor, 7, media) },
		notification_module.NotifFollowingAddedWatched:   func() error { return svc.PushFollowingAddedWatched(ctx, actor, media) },
		notification_module.NotifFollowingAddedWatchlist: func() error { return svc.PushFollowingAddedWatchlist(ctx, actor, media) },
		notification_module.NotifFollowingLikedMedia:     func() error { return svc.PushFollowingLikedMedia(ctx, actor, media) },
		notification_module.NotifMovieNowPlaying:         func() error { return svc.PushMovieNowPlaying(ctx, []uint{owner}, media) },
		notification_module.NotifAchievementUnlocked: func() error {
			return svc.PushAchievementUnlocked(ctx, owner, notification_module.AchievementRef{ID: 5, ActionType: "watched_count", TargetCount: 10, Name: "Cinephile"}, 50)
		},
		notification_module.NotifEmailVerified:   func() error { return svc.PushEmailVerified(ctx, owner) },
		notification_module.NotifPasswordChanged: func() error { return svc.PushPasswordChanged(ctx, owner) },
	}

	for typ, push := range pushes {
		if err := push(); err != nil {
			t.Fatalf("%s: %v", typ, err)
		}
		var raw struct {
			Message string
			Params  *string
		}
		// Recipient is the follower for fan-out types, the owner otherwise.
		if err := e.db.Raw(`SELECT message, params::text AS params FROM notifications
			WHERE type = ? AND user_id IN (?, ?) ORDER BY id DESC LIMIT 1`, string(typ), owner, follower).Scan(&raw).Error; err != nil {
			t.Fatal(err)
		}
		if e.count(`SELECT count(*) FROM notifications WHERE type = ?`, string(typ)) != 1 {
			t.Fatalf("%s: want exactly one row", typ)
		}
		if raw.Message != "" {
			t.Errorf("%s: message must be empty, got %q", typ, raw.Message)
		}
		want := paramKeysByType[typ]
		if len(want) == 0 {
			if raw.Params != nil {
				t.Errorf("%s: params must be NULL, got %s", typ, *raw.Params)
			}
			continue
		}
		if raw.Params == nil {
			t.Fatalf("%s: params missing", typ)
		}
		var got map[string]any
		if err := json.Unmarshal([]byte(*raw.Params), &got); err != nil {
			t.Fatal(err)
		}
		gotKeys, wantKeys := sortedKeys(got), append([]string(nil), want...)
		sort.Strings(wantKeys)
		if len(gotKeys) != len(wantKeys) {
			t.Errorf("%s: params keys %v, want %v", typ, gotKeys, wantKeys)
			continue
		}
		for i := range gotKeys {
			if gotKeys[i] != wantKeys[i] {
				t.Errorf("%s: params keys %v, want %v", typ, gotKeys, wantKeys)
				break
			}
		}
	}

	// Library/like target the real media type; reviews target the review.
	var ref string
	e.db.Raw(`SELECT target_ref FROM notifications WHERE type = 'following_liked_media'`).Scan(&ref)
	if ref != "tv" {
		t.Errorf("following_liked_media target_ref = %q, want tv", ref)
	}
}

func TestNotificationParams_LegacyNullRowStillLists(t *testing.T) {
	e := newEnv(t)
	uid := seedUser(t, e.db, "legacy@notif.test", true, "pw-legacy-1")
	if err := e.db.Exec(`INSERT INTO notifications (created_at, updated_at, user_id, type, category, message, is_read)
		VALUES (now(), now(), ?, 'welcome', 'system', 'legacy text', false)`, uid).Error; err != nil {
		t.Fatal(err)
	}
	svc := newNotifService(t, e.db, fakeUserProvider{}, nil)
	res, err := svc.ListNotifications(context.Background(), uid, notification_module.ListNotificationsQuery{})
	if err != nil {
		t.Fatal(err)
	}
	if len(res.Notifications) != 1 || res.Notifications[0].Message != "legacy text" || res.Notifications[0].Params != nil {
		t.Fatalf("legacy row not listed as-is: %+v", res.Notifications)
	}
}

// The socket payload and the REST list item are the same NotificationResponse: same keys, params included.
func TestNotificationParams_SocketPayloadMatchesListItem(t *testing.T) {
	e := newEnv(t)
	uid := seedUser(t, e.db, "sock@notif.test", true, "pw-sock-1")
	hub := notification_module.NewHub(middleware.NewTokenVerifier(e.cfg.JWT.AccessSecret, e.db), []string{"http://app.test"})
	srv := httptest.NewServer(hub.Handler())
	t.Cleanup(srv.Close)
	sock := e.openSocket(t, srv.URL, e.accessToken(uid, 0))
	if !waitCount(t, hub, 1) {
		t.Fatal("socket did not register")
	}

	svc := newNotifService(t, e.db, fakeUserProvider{}, hub)
	if err := svc.PushMovieNowPlaying(context.Background(), []uint{uid}, notification_module.MediaRef{ID: 9, Type: "movie", Title: "Dune"}); err != nil {
		t.Fatal(err)
	}

	payload := sock.waitEvent("notification:new")
	var pushed map[string]any
	if err := json.Unmarshal(payload, &pushed); err != nil {
		t.Fatal(err)
	}
	res, err := svc.ListNotifications(context.Background(), uid, notification_module.ListNotificationsQuery{})
	if err != nil || len(res.Notifications) != 1 {
		t.Fatalf("list: %v %+v", err, res)
	}
	rawItem, _ := json.Marshal(res.Notifications[0])
	var item map[string]any
	_ = json.Unmarshal(rawItem, &item)

	if _, ok := pushed["params"]; !ok {
		t.Fatalf("socket payload lacks params: %s", payload)
	}
	a, b := sortedKeys(pushed), sortedKeys(item)
	if len(a) != len(b) {
		t.Fatalf("socket keys %v != list keys %v", a, b)
	}
	for i := range a {
		if a[i] != b[i] {
			t.Fatalf("socket keys %v != list keys %v", a, b)
		}
	}
}
