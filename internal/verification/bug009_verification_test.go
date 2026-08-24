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

func bug9DB(t *testing.T) *gorm.DB {
	dsn := fmt.Sprintf("file:bug9_%d?mode=memory&cache=shared", time.Now().UnixNano())
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
func TestBug009VerificationMessageReadFlow(t *testing.T) {
	db := bug9DB(t)
	mr := repository.NewMessageRepo(db)
	if err := mr.Send(&domain.Message{SenderID: 1, ReceiverID: 2, Content: "hello"}); err != nil {
		t.Fatal(err)
	}
	svc := service.NewInteractionService(nil, nil, nil, nil, mr, nil, nil)
	if _, err := svc.ListMessages(2, 1, 1, 20); err != nil {
		t.Fatal(err)
	}
	n, err := svc.UnreadCount(2)
	if err != nil {
		t.Fatal(err)
	}
	if n != 0 {
		t.Fatalf("unread=%d", n)
	}
	var m domain.Message
	db.First(&m)
	if !m.IsRead {
		t.Fatal("message not marked read")
	}
}
func TestBug009VerificationMessageReadRegression(t *testing.T) {
	db := bug9DB(t)
	if _, err := repository.NewMessageRepo(db).UnreadCount(2); err != nil {
		t.Fatal(err)
	}
}
