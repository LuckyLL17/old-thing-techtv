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

func bug16DB(t *testing.T) *gorm.DB {
	dsn := fmt.Sprintf("file:bug16_%d?mode=memory&cache=shared", time.Now().UnixNano())
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	if err := db.AutoMigrate(&domain.User{}, &domain.Category{}, &domain.Tutorial{}); err != nil {
		t.Fatal(err)
	}
	sqlDB, err := db.DB()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = sqlDB.Close() })
	return db
}
func TestBug016VerificationTutorialStatusFilter(t *testing.T) {
	db := bug16DB(t)
	db.Create(&domain.User{ID: 1, Username: "u", Email: "u@e.com", PasswordHash: "x", Status: 1})
	db.Create(&domain.Category{ID: 1, Code: "c", Name: "C", Status: 1})
	db.Create(&domain.Tutorial{ID: 1, UserID: 1, CategoryID: 1, Title: "pub", Status: domain.TutorialStatusPublished})
	db.Create(&domain.Tutorial{ID: 2, UserID: 1, CategoryID: 1, Title: "draft", Status: domain.TutorialStatusDraft})
	list, total, err := repository.NewTutorialRepo(db).List(1, 20, 0, "", domain.TutorialStatusPublished, "new", "", 0)
	if err != nil {
		t.Fatal(err)
	}
	if total != 1 || len(list) != 1 || list[0].Status != domain.TutorialStatusPublished {
		t.Fatalf("status filter total=%d list=%+v", total, list)
	}
}
func TestBug016VerificationTutorialStatusFilterRegression(t *testing.T) {
	db := bug16DB(t)
	if _, err := repository.NewTutorialRepo(db).Count(""); err != nil {
		t.Fatal(err)
	}
}
