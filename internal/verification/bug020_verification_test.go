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

func bug20DB(t *testing.T) *gorm.DB {
	dsn := fmt.Sprintf("file:bug20_%d?mode=memory&cache=shared", time.Now().UnixNano())
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
func TestBug020VerificationTagCountSync(t *testing.T) {
	db := bug20DB(t)
	db.Create(&domain.Tag{ID: 1, Name: "wood", Slug: "wood", TutorialCount: 5})
	db.Create(&domain.User{ID: 1, Username: "u", Email: "u@e.com", PasswordHash: "x", Status: 1})
	db.Create(&domain.Category{ID: 1, Code: "c", Name: "C", Status: 1})
	db.Create(&domain.Tutorial{ID: 1, UserID: 1, CategoryID: 1, Title: "t", Status: domain.TutorialStatusPublished})
	db.Create(&domain.TutorialTag{TutorialID: 1, TagID: 1})
	w := worker.NewStatsUpdater(repository.NewUserRepo(db), repository.NewTutorialRepo(db), repository.NewCommentRepo(db), repository.NewCategoryRepo(db), repository.NewTagRepo(db))
	w.RunOnce()
	var tag domain.Tag
	db.First(&tag, 1)
	if tag.TutorialCount != 1 {
		t.Fatalf("tag count=%d", tag.TutorialCount)
	}
}
func TestBug020VerificationTagCountSyncRegression(t *testing.T) {
	db := bug20DB(t)
	if err := repository.NewTagRepo(db).IncCount(1, 0); err != nil {
		t.Fatal(err)
	}
}
