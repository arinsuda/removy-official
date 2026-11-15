package notification_module

import (
	"encoding/json"
	"reflect"
	"sort"
	"testing"
)

func TestEventResultNotificationsCarryOnlyRoundParams(t *testing.T) {
	ns := EventResultNotifications([]uint{3, 7}, 42, NotificationParams{EventType: "vote_monthly", MediaType: "movie", Period: "2026-05"})
	if len(ns) != 2 {
		t.Fatalf("want 2 notifications, got %d", len(ns))
	}
	for _, n := range ns {
		if n.Type != NotifEventResults || n.Category != CategorySystem || n.Message != "" || n.ActorID != nil {
			t.Errorf("unexpected notification: %+v", n)
		}
		if n.TargetID == nil || *n.TargetID != 42 || n.TargetRef == nil || *n.TargetRef != "event" {
			t.Errorf("target must be the event: %+v", n)
		}
		raw, err := json.Marshal(n.Params)
		if err != nil {
			t.Fatal(err)
		}
		var keys map[string]any
		_ = json.Unmarshal(raw, &keys)
		got := make([]string, 0, len(keys))
		for k := range keys {
			got = append(got, k)
		}
		sort.Strings(got)
		if want := []string{"event_type", "media_type", "period"}; !reflect.DeepEqual(got, want) {
			t.Errorf("params keys = %v, want %v", got, want)
		}
	}
	if ns[0].Params == ns[1].Params {
		t.Error("each notification must own its params")
	}
}
