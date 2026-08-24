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

func bug23DB(t *testing.T) *gorm.DB {
	dsn := fmt.Sprintf("file:bug23_%d?mode=memory&cache=shared", time.Now().UnixNano())
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
func TestBug023VerificationMarkAllReadUserScope(t *testing.T) {
	db := bug23DB(t)
	r := repository.NewNotificationRepo(db)
	db.Create(&domain.Notification{ID: 1, UserID: 1, Type: domain.NotifSystem, Title: "a"})
	db.Create(&domain.Notification{ID: 2, UserID: 2, Type: domain.NotifSystem, Title: "b"})
	n, err := service.NewNotificationService(r).MarkAllRead(1)
	if err != nil || n != 1 {
		t.Fatalf("marked=%d err=%v", n, err)
	}
	var other domain.Notification
	db.First(&other, 2)
	if other.Read {
		t.Fatal("other user's notification was marked read")
	}
}
func TestBug023VerificationMarkAllReadRegression(t *testing.T) {
	db := bug23DB(t)
	if _, err := repository.NewNotificationRepo(db).CountUnread(7); err != nil {
		t.Fatal(err)
	}
}
