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

func bug11DB(t *testing.T) *gorm.DB {
	dsn := fmt.Sprintf("file:bug11_%d?mode=memory&cache=shared", time.Now().UnixNano())
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	if err := db.AutoMigrate(&domain.Message{}); err != nil {
		t.Fatal(err)
	}
	sqlDB, err := db.DB()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = sqlDB.Close() })
	return db
}
func TestBug011VerificationUnreadCountAfterSend(t *testing.T) {
	db := bug11DB(t)
	mr := repository.NewMessageRepo(db)
	svc := service.NewInteractionService(nil, nil, nil, nil, mr, nil, nil)
	if err := svc.SendMessage(2, 1, "hello"); err != nil {
		t.Fatal(err)
	}
	n, err := svc.UnreadCount(1)
	if err != nil {
		t.Fatal(err)
	}
	if n != 1 {
		t.Fatalf("receiver unread=%d", n)
	}
}
func TestBug011VerificationUnreadCountRegression(t *testing.T) {
	db := bug11DB(t)
	if _, err := repository.NewMessageRepo(db).UnreadCount(99); err != nil {
		t.Fatal(err)
	}
}
