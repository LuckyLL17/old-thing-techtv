package repository

import (
	"sync"
	"testing"
	"upcycle-hub/internal/domain"
	apperr "upcycle-hub/pkg/errors"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// newTestDB spins up an in-memory SQLite with the same TranslateError flag the
// server uses, and migrates the User model so the email UNIQUE index exists.
func newTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared&_pragma=busy_timeout=5000"), &gorm.Config{
		TranslateError: true,
	})
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	if err := db.AutoMigrate(&domain.User{}); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	t.Cleanup(func() { db.Exec("DROP TABLE users") })
	return db
}

// TestCreateRejectsDuplicateEmail verifies the database-level guard: a second
// INSERT with an already-used email must fail with a conflict AppError, never
// silently succeed. This is the guard that was missing (no unique index + no
// error translation) and let replayed registrations create duplicate accounts.
func TestCreateRejectsDuplicateEmail(t *testing.T) {
	r := NewUserRepo(newTestDB(t))

	first := &domain.User{Username: "alice", Email: "alice@example.com", PasswordHash: "x", Status: 1}
	if err := r.Create(first); err != nil {
		t.Fatalf("first create: %v", err)
	}

	// Same email, different username — the email unique index must fire.
	dup := &domain.User{Username: "alice2", Email: "alice@example.com", PasswordHash: "x", Status: 1}
	err := r.Create(dup)
	if err == nil {
		t.Fatal("duplicate email was accepted --- FAIL")
	}
	ae, ok := err.(*apperr.AppError)
	if !ok {
		t.Fatalf("expected AppError, got %T: %v", err, err)
	}
	if ae.Code != apperr.CodeConflict {
		t.Fatalf("expected conflict code %d, got %d (%s)", apperr.CodeConflict, ae.Code, ae.Message)
	}
}

// TestCreateConcurrentSameEmail simulates the original bug report: a network
// blip causes the client to fire N identical registrations at once. With the
// unique index in place, exactly one must win and the rest must get a conflict.
func TestCreateConcurrentSameEmail(t *testing.T) {
	r := NewUserRepo(newTestDB(t))
	const n = 20
	var wg sync.WaitGroup
	errs := make([]error, n)
	start := make(chan struct{})

	for i := 0; i < n; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			<-start // release all goroutines at once to maximize the race window
			errs[i] = r.Create(&domain.User{
				Username: "bob", Email: "bob@example.com", PasswordHash: "x", Status: 1,
			})
		}(i)
	}
	close(start)
	wg.Wait()

	ok, fail := 0, 0
	for _, e := range errs {
		if e == nil {
			ok++
		} else {
			fail++
		}
	}
	if ok != 1 {
		t.Fatalf("expected exactly 1 successful insert, got %d (fail=%d)", ok, fail)
	}
	if fail != n-1 {
		t.Fatalf("expected %d conflicts, got %d", n-1, fail)
	}
	// Sanity: exactly one row for that email in the table.
	var count int64
	r.db.Model(&domain.User{}).Where("email = ?", "bob@example.com").Count(&count)
	if count != 1 {
		t.Fatalf("expected 1 row for bob@example.com, got %d", count)
	}
}

// TestGetByEmailActuallyQueriesEmail is the regression guard for the bug that
// made the pre-check a no-op: GetByEmail previously matched against the
// username column, so it never found an existing email.
func TestGetByEmailActuallyQueriesEmail(t *testing.T) {
	db := newTestDB(t)
	r := NewUserRepo(db)
	if err := r.Create(&domain.User{Username: "carol", Email: "carol@example.com", PasswordHash: "x", Status: 1}); err != nil {
		t.Fatalf("create: %v", err)
	}

	got, err := r.GetByEmail("carol@example.com")
	if err != nil {
		t.Fatalf("get by email: %v", err)
	}
	if got.Email != "carol@example.com" {
		t.Fatalf("returned wrong row: email=%s", got.Email)
	}

	// Looking up the email must not match by username.
	if _, err := r.GetByEmail("carol"); err == nil {
		t.Fatal("GetByEmail matched the username column instead of email")
	}
}
