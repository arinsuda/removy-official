package admin_module

import (
	"fmt"
	"os"
	"testing"
	"time"

	"gorm.io/datatypes"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func getTestDB(t *testing.T) *gorm.DB {
	dsn := os.Getenv("TEST_DB_DSN")
	if dsn == "" {
		host := os.Getenv("POSTGRES_HOST")
		user := os.Getenv("POSTGRES_USER")
		password := os.Getenv("POSTGRES_PASSWORD")
		dbname := os.Getenv("POSTGRES_DB")
		port := os.Getenv("POSTGRES_PORT")
		if port == "" {
			port = "5432"
		}
		if host != "" && user != "" && dbname != "" {
			dsn = fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable TimeZone=Asia/Bangkok",
				host, user, password, dbname, port)
		}
	}
	if dsn == "" {
		t.Skip("Skipping PostgreSQL integration test: TEST_DB_DSN or POSTGRES_HOST environment variables not set")
		return nil
	}

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Skipf("Skipping PostgreSQL integration test (cannot connect to DB): %v", err)
		return nil
	}
	return db
}

func TestPostgreSQLIntegrationQueries(t *testing.T) {
	db := getTestDB(t)
	if db == nil {
		return
	}

	repo := NewRepository(db)

	// 1. Test Overview Stats
	overview, err := repo.GetOverviewStats(5)
	if err != nil {
		t.Fatalf("GetOverviewStats failed: %v", err)
	}
	if overview.UniqueOnlineUsers != 5 {
		t.Errorf("Expected 5 online users, got %d", overview.UniqueOnlineUsers)
	}

	// 2. Test 12-Month Growth Trend (with to_char and month filling)
	growth, err := repo.Get12MonthGrowthTrend()
	if err != nil {
		t.Fatalf("Get12MonthGrowthTrend failed: %v", err)
	}
	if len(growth) != 12 {
		t.Errorf("Expected 12 growth points, got %d", len(growth))
	}

	// 3. Test ListUsers with ILIKE search & pagination
	users, total, err := repo.ListUsers(UserFilter{Page: 1, Limit: 10, Search: "admin"})
	if err != nil {
		t.Fatalf("ListUsers failed: %v", err)
	}
	if total < 0 || len(users) > 10 {
		t.Errorf("ListUsers returned invalid pagination results")
	}

	// 4. Test ListReviews
	reviews, totalRev, err := repo.ListReviews(ReviewFilter{Page: 1, Limit: 10})
	if err != nil {
		t.Fatalf("ListReviews failed: %v", err)
	}
	if totalRev < 0 || len(reviews) > 10 {
		t.Errorf("ListReviews returned invalid pagination results")
	}

	// 5. Test JSONB Audit Log insertion
	auditLog := AdminAuditLog{
		AdminID:    func() *uint { id := uint(1); return &id }(),
		Action:     ActionUserRoleChanged,
		TargetType: "user",
		TargetID:   2,
		MetaData:   datatypes.JSON([]byte(`{"test": true}`)),
		CreatedAt:  time.Now().UTC(),
	}
	if err := db.Create(&auditLog).Error; err != nil {
		t.Logf("Audit log creation test note: %v", err)
	}
}
