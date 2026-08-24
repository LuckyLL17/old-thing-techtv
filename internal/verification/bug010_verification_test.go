package verification

import (
	"fmt"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"testing"
	"time"
	"upcycle-hub/internal/domain"
	"upcycle-hub/internal/repository"
)

func bug10DB(t *testing.T) *gorm.DB {
	dsn := fmt.Sprintf("file:bug10_%d?mode=memory&cache=shared", time.Now().UnixNano())
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
func TestBug010VerificationMessagePagination(t *testing.T) {
	db := bug10DB(t)
	r := repository.NewMessageRepo(db)
	for i := 0; i < 4; i++ {
		if err := db.Create(&domain.Message{SenderID: 1, ReceiverID: 2, Content: fmt.Sprintf("m%d", i)}).Error; err != nil {
			t.Fatal(err)
		}
	}
	a, err := r.List(1, 2, 1, 2)
	if err != nil {
		t.Fatal(err)
	}
	b, err := r.List(1, 2, 2, 2)
	if err != nil {
		t.Fatal(err)
	}
	if len(a) != 2 || len(b) != 2 || a[1].ID == b[0].ID {
		t.Fatalf("pages overlap: %v %v", a, b)
	}
}
func TestBug010VerificationMessagePaginationRegression(t *testing.T) {
	db := bug10DB(t)
	if err := repository.NewMessageRepo(db).Send(&domain.Message{SenderID: 1, ReceiverID: 2, Content: "ok"}); err != nil {
		t.Fatal(err)
	}
}
