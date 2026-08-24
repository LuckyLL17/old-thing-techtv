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

func bug18DB(t *testing.T) *gorm.DB {
	dsn := fmt.Sprintf("file:bug18_%d?mode=memory&cache=shared", time.Now().UnixNano())
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
func TestBug018VerificationTutorialViewIncrement(t *testing.T) {
	db := bug18DB(t)
	db.Create(&domain.User{ID: 1, Username: "u", Email: "u@e.com", PasswordHash: "x", Status: 1})
	db.Create(&domain.Category{ID: 1, Code: "c", Name: "C", Status: 1})
	db.Create(&domain.Tutorial{ID: 1, UserID: 1, CategoryID: 1, Title: "t", Status: domain.TutorialStatusPublished})
	svc := service.NewTutorialService(repository.NewTutorialRepo(db), nil, nil, nil, nil, nil)
	if _, err := svc.Get(1, true); err != nil {
		t.Fatal(err)
	}
	var got domain.Tutorial
	db.First(&got, 1)
	if got.ViewCount != 1 {
		t.Fatalf("views=%d", got.ViewCount)
	}
}
func TestBug018VerificationTutorialViewIncrementRegression(t *testing.T) {
	db := bug18DB(t)
	if err := repository.NewTutorialRepo(db).IncView(1); err != nil {
		t.Fatal(err)
	}
}
