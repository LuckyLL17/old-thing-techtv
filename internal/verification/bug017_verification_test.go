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

func bug17DB(t *testing.T) *gorm.DB {
	dsn := fmt.Sprintf("file:bug17_%d?mode=memory&cache=shared", time.Now().UnixNano())
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
func TestBug017VerificationPopularTutorialOrder(t *testing.T) {
	db := bug17DB(t)
	db.Create(&domain.User{ID: 1, Username: "u", Email: "u@e.com", PasswordHash: "x", Status: 1})
	db.Create(&domain.Category{ID: 1, Code: "c", Name: "C", Status: 1})
	db.Create(&domain.Tutorial{ID: 1, UserID: 1, CategoryID: 1, Title: "low", Status: domain.TutorialStatusPublished, FavoriteCount: 1, ViewCount: 99})
	db.Create(&domain.Tutorial{ID: 2, UserID: 1, CategoryID: 1, Title: "high", Status: domain.TutorialStatusPublished, FavoriteCount: 10, ViewCount: 1})
	list, _, err := service.NewTutorialService(repository.NewTutorialRepo(db), nil, nil, nil, nil, nil).List(1, 20, 0, "", domain.TutorialStatusPublished, "popular", "", 0)
	if err != nil {
		t.Fatal(err)
	}
	if len(list) < 2 || list[0].ID != 2 {
		t.Fatalf("popular order=%v", list)
	}
}
func TestBug017VerificationPopularTutorialOrderRegression(t *testing.T) {
	db := bug17DB(t)
	if _, err := repository.NewTutorialRepo(db).Count(domain.TutorialStatusPublished); err != nil {
		t.Fatal(err)
	}
}
