package verification

import (
	"fmt"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"testing"
	"time"
	"upcycle-hub/internal/domain"
	"upcycle-hub/internal/repository"
	"upcycle-hub/internal/service"
)

func bug24DB(t *testing.T) *gorm.DB {
	dsn := fmt.Sprintf("file:bug24_%d?mode=memory&cache=shared", time.Now().UnixNano())
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	if err := db.AutoMigrate(&domain.Notification{}); err != nil {
		t.Fatal(err)
	}
	sqlDB, err := db.DB()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = sqlDB.Close() })
	return db
}
func TestBug024VerificationUnreadNotificationFilter(t *testing.T) {
	db := bug24DB(t)
	r := repository.NewNotificationRepo(db)
	db.Create(&domain.Notification{ID: 1, UserID: 1, Type: domain.NotifSystem, Title: "unread", Read: false})
	db.Create(&domain.Notification{ID: 2, UserID: 1, Type: domain.NotifSystem, Title: "read", Read: true})
	list, total, err := service.NewNotificationService(r).List(1, 1, 20, true)
	if err != nil {
		t.Fatal(err)
	}
	if total != 1 || len(list) != 1 || list[0].ID != 1 {
		t.Fatalf("unread list total=%d list=%+v", total, list)
	}
}
func TestBug024VerificationUnreadNotificationRegression(t *testing.T) {
	db := bug24DB(t)
	if _, err := repository.NewNotificationRepo(db).CountUnread(1); err != nil {
		t.Fatal(err)
	}
}
