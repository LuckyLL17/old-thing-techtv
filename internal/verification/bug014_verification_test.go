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

func bug14DB(t *testing.T) *gorm.DB {
	dsn := fmt.Sprintf("file:bug14_%d?mode=memory&cache=shared", time.Now().UnixNano())
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	if err := db.AutoMigrate(&domain.Attempt{}); err != nil {
		t.Fatal(err)
	}
	sqlDB, err := db.DB()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = sqlDB.Close() })
	return db
}
func TestBug014VerificationAttemptUpdate(t *testing.T) {
	db := bug14DB(t)
	ar := repository.NewAttemptRepo(db)
	db.Create(&domain.Attempt{UserID: 1, TutorialID: 2, Completed: false, Note: "first"})
	svc := service.NewInteractionService(nil, nil, ar, nil, nil, nil, nil)
	if err := svc.MarkAttempt(1, 2, true, "finished"); err != nil {
		t.Fatal(err)
	}
	var got domain.Attempt
	db.Where("user_id=1 AND tutorial_id=2").First(&got)
	if !got.Completed || got.Note != "finished" {
		t.Fatalf("attempt=%+v", got)
	}
}
func TestBug014VerificationAttemptUpdateRegression(t *testing.T) {
	db := bug14DB(t)
	if err := repository.NewAttemptRepo(db).Create(&domain.Attempt{UserID: 9, TutorialID: 9}); err != nil {
		t.Fatal(err)
	}
}
