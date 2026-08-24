package repository

import (
	"testing"

	"upcycle-hub/internal/domain"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func newTestNotificationDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	if err := db.AutoMigrate(&domain.Notification{}); err != nil {
		t.Fatalf("automigrate: %v", err)
	}
	return db
}

// seedUnreadNotifications inserts a mix of read/unread notifications across two
// users so user isolation can be observed.
func seedUnreadNotifications(t *testing.T, db *gorm.DB) {
	t.Helper()
	notes := []*domain.Notification{
		{UserID: 1, Type: domain.NotifSystem, Title: "u1 a", Read: false},
		{UserID: 1, Type: domain.NotifSystem, Title: "u1 b", Read: false},
		{UserID: 1, Type: domain.NotifSystem, Title: "u1 c (already read)", Read: true},
		{UserID: 2, Type: domain.NotifSystem, Title: "u2 a", Read: false},
		{UserID: 2, Type: domain.NotifSystem, Title: "u2 b", Read: false},
	}
	if err := db.Create(&notes).Error; err != nil {
		t.Fatalf("seed: %v", err)
	}
}

func TestNotificationRepo_MarkAllReadIsolatesByUser(t *testing.T) {
	db := newTestNotificationDB(t)
	seedUnreadNotifications(t, db)
	repo := NewNotificationRepo(db)

	n, err := repo.MarkAllRead(1)
	if err != nil {
		t.Fatalf("MarkAllRead: %v", err)
	}
	if n != 2 {
		t.Fatalf("expected 2 affected (only user 1's unread), got %d", n)
	}

	var stillUnread []domain.Notification
	if err := db.Where("user_id = ? AND `read` = ?", 2, false).Find(&stillUnread).Error; err != nil {
		t.Fatalf("query user2 unread: %v", err)
	}
	if len(stillUnread) != 2 {
		t.Fatalf("user 2 unread must be untouched (expected 2), got %d", len(stillUnread))
	}

	var u1Unread int64
	if err := db.Model(&domain.Notification{}).Where("user_id = ? AND `read` = ?", 1, false).Count(&u1Unread).Error; err != nil {
		t.Fatalf("count user1 unread: %v", err)
	}
	if u1Unread != 0 {
		t.Fatalf("user 1 should have 0 unread, got %d", u1Unread)
	}
}

func TestNotificationRepo_MarkAllReadReturnsAccurateCount(t *testing.T) {
	db := newTestNotificationDB(t)
	seedUnreadNotifications(t, db)
	repo := NewNotificationRepo(db)

	n, err := repo.MarkAllRead(2)
	if err != nil {
		t.Fatalf("MarkAllRead: %v", err)
	}
	if n != 2 {
		t.Fatalf("expected 2 affected for user 2, got %d", n)
	}

	// Second invocation on the now-empty set must report 0, not a stale count.
	n2, err := repo.MarkAllRead(2)
	if err != nil {
		t.Fatalf("MarkAllRead (idempotent): %v", err)
	}
	if n2 != 0 {
		t.Fatalf("expected 0 on idempotent re-run, got %d", n2)
	}
}
