package admin_module

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/arinsuda/movie-hub/config"
	"github.com/arinsuda/movie-hub/middleware"
	"github.com/gofiber/fiber/v3"
)

type mockRepository struct {
	overviewFunc     func(onlineUsers int) (*OverviewStats, error)
	growthFunc       func() ([]GrowthPoint, error)
	listUsersFunc    func(filter UserFilter) ([]AdminUserRow, int64, error)
	listReviewsFunc  func(filter ReviewFilter) ([]AdminReviewRow, int64, error)
	listAuditFunc    func(filter AuditLogFilter) ([]AdminAuditLogRow, int64, error)
	updateRoleFunc   func(adminID, targetUserID uint, newRole string, reason *string) error
	updateStatusFunc func(adminID, targetUserID uint, isActive bool, reason *string) error
	deleteUserFunc   func(adminID, targetUserID uint, reason *string) error
	deleteReviewFunc func(adminID, reviewID uint, reason *string) error
}

func (m *mockRepository) GetOverviewStats(onlineUsers int) (*OverviewStats, error) {
	if m.overviewFunc != nil {
		return m.overviewFunc(onlineUsers)
	}
	return &OverviewStats{}, nil
}

func (m *mockRepository) Get12MonthGrowthTrend() ([]GrowthPoint, error) {
	if m.growthFunc != nil {
		return m.growthFunc()
	}
	return []GrowthPoint{}, nil
}

func (m *mockRepository) ListUsers(filter UserFilter) ([]AdminUserRow, int64, error) {
	if m.listUsersFunc != nil {
		return m.listUsersFunc(filter)
	}
	return []AdminUserRow{}, 0, nil
}

func (m *mockRepository) ListReviews(filter ReviewFilter) ([]AdminReviewRow, int64, error) {
	if m.listReviewsFunc != nil {
		return m.listReviewsFunc(filter)
	}
	return []AdminReviewRow{}, 0, nil
}

func (m *mockRepository) ListAuditLogs(filter AuditLogFilter) ([]AdminAuditLogRow, int64, error) {
	if m.listAuditFunc != nil {
		return m.listAuditFunc(filter)
	}
	return []AdminAuditLogRow{}, 0, nil
}

func (m *mockRepository) UpdateUserRole(adminID, targetUserID uint, newRole string, reason *string) error {
	if m.updateRoleFunc != nil {
		return m.updateRoleFunc(adminID, targetUserID, newRole, reason)
	}
	return nil
}

func (m *mockRepository) UpdateUserStatus(adminID, targetUserID uint, isActive bool, reason *string) error {
	if m.updateStatusFunc != nil {
		return m.updateStatusFunc(adminID, targetUserID, isActive, reason)
	}
	return nil
}

func (m *mockRepository) DeleteUser(adminID, targetUserID uint, reason *string) error {
	if m.deleteUserFunc != nil {
		return m.deleteUserFunc(adminID, targetUserID, reason)
	}
	return nil
}

func (m *mockRepository) DeleteReview(adminID, reviewID uint, reason *string) error {
	if m.deleteReviewFunc != nil {
		return m.deleteReviewFunc(adminID, reviewID, reason)
	}
	return nil
}

func setupTestApp(repo Repository) *fiber.App {
	app := fiber.New()
	svc := NewService(repo, nil, nil)
	h := NewHandler(svc)

	mockAuth := func(userID uint, role string) fiber.Handler {
		return func(c fiber.Ctx) error {
			if userID > 0 {
				c.Locals("claims", &middleware.Claims{UserID: userID, Role: role})
			}
			return c.Next()
		}
	}

	app.Get("/api/admin/overview", mockAuth(1, "admin"), h.GetOverview)
	app.Get("/api/admin/growth", mockAuth(1, "admin"), h.GetGrowth)
	app.Get("/api/admin/users", mockAuth(1, "admin"), h.ListUsers)
	app.Patch("/api/admin/users/:userId/role", mockAuth(1, "admin"), h.UpdateUserRole)
	app.Patch("/api/admin/users/:userId/status", mockAuth(1, "admin"), h.UpdateUserStatus)
	app.Delete("/api/admin/users/:userId", mockAuth(1, "admin"), h.DeleteUser)
	app.Get("/api/admin/reviews", mockAuth(1, "admin"), h.ListReviews)
	app.Delete("/api/admin/reviews/:reviewId", mockAuth(1, "admin"), h.DeleteReview)
	app.Get("/api/admin/audit-logs", mockAuth(1, "admin"), h.ListAuditLogs)

	return app
}

// ── Unit Tests ─────────────────────────────────────────────────────────────

func TestOverviewZeroBaselineHandling(t *testing.T) {
	mockRepo := &mockRepository{
		overviewFunc: func(onlineUsers int) (*OverviewStats, error) {
			absGrowth := int64(24)
			var growthPct *float64
			status := GrowthStatusNoPreviousBaseline

			return &OverviewStats{
				TotalRegisteredUsers:       24,
				ActiveUsersCount:           24,
				CurrentMonthRegistrations:  24,
				PreviousMonthRegistrations: 0,
				AbsoluteGrowth:             absGrowth,
				GrowthPercentage:           growthPct,
				GrowthStatus:               status,
			}, nil
		},
	}

	app := setupTestApp(mockRepo)
	req := httptest.NewRequest(http.MethodGet, "/api/admin/overview", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("Test request failed: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Expected 200 OK, got %d", resp.StatusCode)
	}

	var body map[string]OverviewStats
	json.NewDecoder(resp.Body).Decode(&body)

	overview := body["overview"]
	if overview.GrowthStatus != GrowthStatusNoPreviousBaseline {
		t.Errorf("Expected growth status %s, got %s", GrowthStatusNoPreviousBaseline, overview.GrowthStatus)
	}
	if overview.GrowthPercentage != nil {
		t.Errorf("Expected nil growth percentage for zero baseline, got %v", *overview.GrowthPercentage)
	}
}

func TestUpdateUserStatusSafetyRules(t *testing.T) {
	mockRepo := &mockRepository{
		updateStatusFunc: func(adminID, targetUserID uint, isActive bool, reason *string) error {
			if adminID == targetUserID && !isActive {
				return ErrSelfDeactivation
			}
			if !isActive && targetUserID == 1 {
				return ErrFinalAdminProtection
			}
			return nil
		},
	}

	app := setupTestApp(mockRepo)

	reqBody := `{"is_active": false}`
	req := httptest.NewRequest(http.MethodPatch, "/api/admin/users/1/status", strings.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("Test request failed: %v", err)
	}

	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("Expected 400 Bad Request for self-deactivation, got %d", resp.StatusCode)
	}
}

func TestDeleteUserSafetyRules(t *testing.T) {
	mockRepo := &mockRepository{
		deleteUserFunc: func(adminID, targetUserID uint, reason *string) error {
			if adminID == targetUserID {
				return ErrSelfDeletion
			}
			if targetUserID == 1 {
				return ErrFinalAdminProtection
			}
			return nil
		},
	}

	app := setupTestApp(mockRepo)

	// Self-deletion check -> 400 Bad Request
	reqSelf := httptest.NewRequest(http.MethodDelete, "/api/admin/users/1", nil)
	respSelf, err := app.Test(reqSelf)
	if err != nil {
		t.Fatalf("Test request failed: %v", err)
	}
	if respSelf.StatusCode != http.StatusBadRequest {
		t.Errorf("Expected 400 Bad Request for self-deletion, got %d", respSelf.StatusCode)
	}

	// Normal deletion of user 2 -> 240/204 No Content
	reqUser := httptest.NewRequest(http.MethodDelete, "/api/admin/users/2", nil)
	respUser, err := app.Test(reqUser)
	if err != nil {
		t.Fatalf("Test request failed: %v", err)
	}
	if respUser.StatusCode != http.StatusNoContent {
		t.Errorf("Expected 204 No Content for successful user deletion, got %d", respUser.StatusCode)
	}
}

func TestUpdateUserRoleInactiveUser(t *testing.T) {
	mockRepo := &mockRepository{
		updateRoleFunc: func(adminID, targetUserID uint, newRole string, reason *string) error {
			return ErrInactiveUserRoleChange
		},
	}

	app := setupTestApp(mockRepo)

	reqBody := `{"role": "admin"}`
	req := httptest.NewRequest(http.MethodPatch, "/api/admin/users/2/role", strings.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("Test request failed: %v", err)
	}

	if resp.StatusCode != http.StatusConflict {
		t.Errorf("Expected 409 Conflict for inactive user role change, got %d", resp.StatusCode)
	}
}

func TestListUsersPaginationEnvelope(t *testing.T) {
	mockRepo := &mockRepository{
		listUsersFunc: func(filter UserFilter) ([]AdminUserRow, int64, error) {
			items := []AdminUserRow{
				{ID: 1, Username: "admin", Email: "admin@example.com", Role: "admin", IsActive: true, CreatedAt: time.Now().Format(time.RFC3339)},
			}
			return items, 1, nil
		},
	}

	app := setupTestApp(mockRepo)

	req := httptest.NewRequest(http.MethodGet, "/api/admin/users?page=1&limit=20", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("Test request failed: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Expected 200 OK, got %d", resp.StatusCode)
	}

	var res PaginatedResponse[AdminUserRow]
	json.NewDecoder(resp.Body).Decode(&res)

	if res.Total != 1 {
		t.Errorf("Expected total 1, got %d", res.Total)
	}
	if len(res.Items) != 1 {
		t.Errorf("Expected 1 item, got %d", len(res.Items))
	}
	if res.TotalPages != 1 {
		t.Errorf("Expected 1 total_pages, got %d", res.TotalPages)
	}
}

// ── HTTP Authorization Tests ──────────────────────────────────────────────

func TestHTTPAuthorizationMiddleware(t *testing.T) {
	app := fiber.New()
	mwAuth := middleware.NewAuthMiddleware(&config.Config{}, nil)

	mockClaimsSetter := func(token string) fiber.Handler {
		return func(c fiber.Ctx) error {
			if token == "unauthenticated" {
				return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "unauthorized"})
			}
			if token == "normal_user" {
				c.Locals("claims", &middleware.Claims{UserID: 2, Role: "user"})
				return mwAuth.RequireRole("admin")(c)
			}
			if token == "active_admin" {
				c.Locals("claims", &middleware.Claims{UserID: 1, Role: "admin"})
				return c.Next()
			}
			return c.Next()
		}
	}

	app.Get("/api/admin/test-401", mockClaimsSetter("unauthenticated"), func(c fiber.Ctx) error {
		return c.SendString("ok")
	})
	app.Get("/api/admin/test-403", mockClaimsSetter("normal_user"), func(c fiber.Ctx) error {
		return c.SendString("ok")
	})
	app.Get("/api/admin/test-200", mockClaimsSetter("active_admin"), func(c fiber.Ctx) error {
		return c.SendString("ok")
	})

	// 1. Unauthenticated -> 401
	req1 := httptest.NewRequest(http.MethodGet, "/api/admin/test-401", nil)
	resp1, _ := app.Test(req1)
	if resp1.StatusCode != http.StatusUnauthorized {
		t.Errorf("Expected 401 Unauthorized, got %d", resp1.StatusCode)
	}

	// 2. Normal user -> 403
	req2 := httptest.NewRequest(http.MethodGet, "/api/admin/test-403", nil)
	resp2, _ := app.Test(req2)
	if resp2.StatusCode != http.StatusForbidden {
		t.Errorf("Expected 403 Forbidden for normal user, got %d", resp2.StatusCode)
	}

	// 3. Active admin -> 200
	req3 := httptest.NewRequest(http.MethodGet, "/api/admin/test-200", nil)
	resp3, _ := app.Test(req3)
	if resp3.StatusCode != http.StatusOK {
		t.Errorf("Expected 200 OK for active admin, got %d", resp3.StatusCode)
	}
}

// ── Concurrency & Realtime Connection Tracking Tests ───────────────────────

type threadSafeSocketTracker struct {
	mu          sync.RWMutex
	userSockets map[uint]map[string]bool
}

func newThreadSafeTracker() *threadSafeSocketTracker {
	return &threadSafeSocketTracker{
		userSockets: make(map[uint]map[string]bool),
	}
}

func (t *threadSafeSocketTracker) Add(userID uint, socketID string) {
	t.mu.Lock()
	defer t.mu.Unlock()
	if _, ok := t.userSockets[userID]; !ok {
		t.userSockets[userID] = make(map[string]bool)
	}
	t.userSockets[userID][socketID] = true
}

func (t *threadSafeSocketTracker) Remove(userID uint, socketID string) {
	t.mu.Lock()
	defer t.mu.Unlock()
	if sockets, ok := t.userSockets[userID]; ok {
		delete(sockets, socketID)
		if len(sockets) == 0 {
			delete(t.userSockets, userID)
		}
	}
}

func (t *threadSafeSocketTracker) UniqueCount() int {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return len(t.userSockets)
}

func TestUniqueOnlineUserTracking(t *testing.T) {
	tracker := newThreadSafeTracker()

	// User 1 opens Tab A and Tab B
	tracker.Add(1, "sock_1_a")
	tracker.Add(1, "sock_1_b")

	// User 2 opens Tab A
	tracker.Add(2, "sock_2_a")

	// Unique online count should be 2
	if count := tracker.UniqueCount(); count != 2 {
		t.Fatalf("Expected 2 unique online users, got %d", count)
	}

	// User 1 closes Tab A (Tab B still active)
	tracker.Remove(1, "sock_1_a")
	if count := tracker.UniqueCount(); count != 2 {
		t.Fatalf("Expected User 1 to remain online while Tab B is active, got %d", count)
	}

	// User 1 closes Tab B (no active tabs left)
	tracker.Remove(1, "sock_1_b")
	if count := tracker.UniqueCount(); count != 1 {
		t.Fatalf("Expected User 1 to be removed from online users, got %d", count)
	}
}

func TestConcurrentOnlineUserRaceCondition(t *testing.T) {
	tracker := newThreadSafeTracker()
	var wg sync.WaitGroup

	// Concurrently connect and disconnect 100 users with multiple tabs
	for i := 1; i <= 100; i++ {
		wg.Add(2)
		userID := uint(i)

		go func(u uint) {
			defer wg.Done()
			tracker.Add(u, "sock_1")
			tracker.Add(u, "sock_2")
		}(userID)

		go func(u uint) {
			defer wg.Done()
			time.Sleep(1 * time.Millisecond)
			tracker.Remove(u, "sock_1")
			tracker.Remove(u, "sock_2")
		}(userID)
	}

	wg.Wait()
}
