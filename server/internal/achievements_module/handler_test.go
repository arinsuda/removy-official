package achievementsmodule

import (
	"context"
	"errors"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/arinsuda/movie-hub/internal/privacy_policy"
	"github.com/gofiber/fiber/v3"
)

type fakeService struct {
	Service
	err error
}

func (f fakeService) ListUserAchievements(context.Context, uint, uint, UserAchievementFilter) (ListUserAchievementsResponse, error) {
	return ListUserAchievementsResponse{}, f.err
}

func TestListUserAchievementsMapsPrivacyErrors(t *testing.T) {
	cases := []struct {
		name string
		err  error
		code int
		body string
	}{
		{"forbidden", privacy_policy.ErrForbidden, 403, `{"error":"forbidden"}`},
		{"wrapped forbidden", errors.Join(errors.New("ctx"), privacy_policy.ErrForbidden), 403, `{"error":"forbidden"}`},
		{"not found", privacy_policy.ErrNotFound, 404, `{"error":"user not found"}`},
		{"other", errors.New("pq: secret detail"), 500, `{"error":"internal server error"}`},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			app := fiber.New()
			app.Get("/users/:userID/achievements", newHandler(fakeService{err: tc.err}).listUserAchievements)
			resp, err := app.Test(httptest.NewRequest("GET", "/users/7/achievements", nil))
			if err != nil {
				t.Fatal(err)
			}
			defer resp.Body.Close()
			var sb strings.Builder
			buf := make([]byte, 256)
			n, _ := resp.Body.Read(buf)
			sb.Write(buf[:n])
			if resp.StatusCode != tc.code || sb.String() != tc.body {
				t.Fatalf("got %d %s, want %d %s", resp.StatusCode, sb.String(), tc.code, tc.body)
			}
		})
	}
}
