package verification

import (
	"fmt"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"testing"
	"time"
	"upcycle-hub/internal/domain"
	"upcycle-hub/internal/repository"
	"upcycle-hub/internal/worker"
)

func bug19DB(t *testing.T) *gorm.DB {
	dsn := fmt.Sprintf("file:bug19_%d?mode=memory&cache=shared", time.Now().UnixNano())
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	if err := db.AutoMigrate(&domain.User{}, &domain.Category{}, &domain.Tag{}, &domain.Tutorial{}, &domain.TutorialTag{}); err != nil {
		t.Fatal(err)
	}
	sqlDB, err := db.DB()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = sqlDB.Close() })
	return db
}
func TestBug019VerificationCategoryCountSync(t *testing.T) {
	db := bug19DB(t)
	db.Create(&domain.Category{ID: 1, Code: "c", Name: "C", Status: 1, TutorialCount: 0})
	db.Create(&domain.User{ID: 1, Username: "u", Email: "u@e.com", PasswordHash: "x", Status: 1})
	db.Create(&domain.Tutorial{ID: 1, UserID: 1, CategoryID: 1, Title: "t", Status: domain.TutorialStatusPublished})
	w := worker.NewStatsUpdater(repository.NewUserRepo(db), repository.NewTutorialRepo(db), repository.NewCommentRepo(db), repository.NewCategoryRepo(db), repository.NewTagRepo(db))
	w.RunOnce()
	var c domain.Category
	db.First(&c, 1)
	if c.TutorialCount != 1 {
		t.Fatalf("category count=%d", c.TutorialCount)
	}
}
func TestBug019VerificationCategoryCountSyncRegression(t *testing.T) {
	db := bug19DB(t)
	w := worker.NewStatsUpdater(repository.NewUserRepo(db), repository.NewTutorialRepo(db), repository.NewCommentRepo(db), repository.NewCategoryRepo(db), repository.NewTagRepo(db))
	w.RunOnce()
}
