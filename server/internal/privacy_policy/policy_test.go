package privacy_policy

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"errors"
	"io"
	"sync/atomic"
	"testing"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type stubQueryResult struct {
	cols []string
	rows [][]driver.Value
	err  error
}

type stubDriver struct {
	handler func(query string, args []driver.NamedValue) stubQueryResult
}

func (d *stubDriver) Open(string) (driver.Conn, error) {
	return &stubConn{handler: d.handler}, nil
}

type stubConn struct {
	handler func(query string, args []driver.NamedValue) stubQueryResult
}

func (c *stubConn) Prepare(string) (driver.Stmt, error) { return nil, errors.New("not implemented") }
func (c *stubConn) Close() error                        { return nil }
func (c *stubConn) Begin() (driver.Tx, error)           { return nil, errors.New("not implemented") }
func (c *stubConn) QueryContext(_ context.Context, query string, args []driver.NamedValue) (driver.Rows, error) {
	res := c.handler(query, args)
	if res.err != nil {
		return nil, res.err
	}
	return &stubRows{cols: res.cols, rows: res.rows}, nil
}

type stubRows struct {
	cols []string
	rows [][]driver.Value
	idx  int
}

func (r *stubRows) Columns() []string { return r.cols }
func (r *stubRows) Close() error      { return nil }
func (r *stubRows) Next(dest []driver.Value) error {
	if r.idx >= len(r.rows) {
		return io.EOF
	}
	copy(dest, r.rows[r.idx])
	r.idx++
	return nil
}

var driverSeq atomic.Int64

func newStubPolicy(t *testing.T, handler func(query string, args []driver.NamedValue) stubQueryResult) UserAccessPolicy {
	t.Helper()
	name := "privacy_policy_stub_" + string(rune('a'+driverSeq.Add(1)))
	sql.Register(name, &stubDriver{handler: handler})
	sqlDB, err := sql.Open(name, "")
	if err != nil {
		t.Fatalf("sql.Open: %v", err)
	}
	t.Cleanup(func() { _ = sqlDB.Close() })

	db, err := gorm.Open(postgres.New(postgres.Config{
		Conn:                 sqlDB,
		PreferSimpleProtocol: true,
	}), &gorm.Config{})
	if err != nil {
		t.Fatalf("gorm.Open: %v", err)
	}
	return NewUserAccessPolicy(db)
}

func TestCanViewProfileSectionViewerZero(t *testing.T) {
	ctx := context.Background()

	t.Run("private profile denies viewer 0 without querying follows", func(t *testing.T) {
		policy := newStubPolicy(t, func(query string, _ []driver.NamedValue) stubQueryResult {
			return stubQueryResult{
				cols: []string{"is_private"},
				rows: [][]driver.Value{{true}},
			}
		})
		for _, sec := range []ProfileSection{SectionProfile, SectionReviews, SectionLibrary, SectionAchievements} {
			ok, err := policy.CanViewProfileSection(ctx, 0, 10, sec)
			if err != nil || ok {
				t.Fatalf("section %s: got (%v, %v), want (false, nil)", sec, ok, err)
			}
		}
	})

	t.Run("public profile allows viewer 0", func(t *testing.T) {
		policy := newStubPolicy(t, func(query string, _ []driver.NamedValue) stubQueryResult {
			return stubQueryResult{
				cols: []string{"is_private"},
				rows: [][]driver.Value{{false}},
			}
		})
		for _, sec := range []ProfileSection{SectionProfile, SectionReviews, SectionLibrary, SectionAchievements} {
			ok, err := policy.CanViewProfileSection(ctx, 0, 10, sec)
			if err != nil || !ok {
				t.Fatalf("section %s: got (%v, %v), want (true, nil)", sec, ok, err)
			}
		}
	})

	t.Run("missing or unavailable user returns ErrNotFound for viewer 0", func(t *testing.T) {
		policy := newStubPolicy(t, func(query string, _ []driver.NamedValue) stubQueryResult {
			return stubQueryResult{cols: []string{"is_private"}, rows: nil}
		})
		ok, err := policy.CanViewProfileSection(ctx, 0, 999, SectionProfile)
		if ok || !errors.Is(err, ErrNotFound) {
			t.Fatalf("got (%v, %v), want (false, ErrNotFound)", ok, err)
		}
	})

	t.Run("ownerID 0 returns ErrNotFound even when viewerID is 0", func(t *testing.T) {
		policy := newStubPolicy(t, func(query string, _ []driver.NamedValue) stubQueryResult {
			t.Fatal("unexpected DB query for ownerID 0")
			return stubQueryResult{}
		})
		ok, err := policy.CanViewProfileSection(ctx, 0, 0, SectionProfile)
		if ok || !errors.Is(err, ErrNotFound) {
			t.Fatalf("got (%v, %v), want (false, ErrNotFound)", ok, err)
		}
	})
}

func TestCanViewActivityAndListViewerZero(t *testing.T) {
	ctx := context.Background()
	policy := newStubPolicy(t, func(query string, _ []driver.NamedValue) stubQueryResult {
		return stubQueryResult{
			cols: []string{"is_private"},
			rows: [][]driver.Value{{true}},
		}
	})

	if ok, err := policy.CanViewActivity(ctx, 0, 10, ActivityReviewCreated, VisibilityFollowers); err != nil || ok {
		t.Fatalf("CanViewActivity(followers): got (%v, %v), want (false, nil)", ok, err)
	}
	if ok, err := policy.CanViewActivity(ctx, 0, 10, ActivityReviewCreated, VisibilityPrivate); err != nil || ok {
		t.Fatalf("CanViewActivity(private): got (%v, %v), want (false, nil)", ok, err)
	}
	if ok, err := policy.CanViewActivity(ctx, 0, 10, ActivityReviewCreated, VisibilityDefault); err != nil || ok {
		t.Fatalf("CanViewActivity(default on private): got (%v, %v), want (false, nil)", ok, err)
	}
	if ok, err := policy.CanViewList(ctx, 0, 10, ListFollowers); err != nil || ok {
		t.Fatalf("CanViewList(followers): got (%v, %v), want (false, nil)", ok, err)
	}
	if ok, err := policy.CanViewList(ctx, 0, 10, ListPrivate); err != nil || ok {
		t.Fatalf("CanViewList(private): got (%v, %v), want (false, nil)", ok, err)
	}
	if ok, err := policy.CanViewList(ctx, 0, 0, ListPublic); !errors.Is(err, ErrNotFound) || ok {
		t.Fatalf("CanViewList(owner=0): got (%v, %v), want (false, ErrNotFound)", ok, err)
	}
}
