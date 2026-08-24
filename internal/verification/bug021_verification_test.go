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

func bug21DB(t *testing.T) *gorm.DB {
	dsn := fmt.Sprintf("file:bug21_%d?mode=memory&cache=shared", time.Now().UnixNano())
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
func TestBug021VerificationUserLevelRefresh(t *testing.T) {
	db := bug21DB(t)
	db.Create(&domain.User{ID: 1, Username: "u", Email: "u@e.com", PasswordHash: "x", Status: 1, Score: 1000, Level: domain.UserLevelNovice})
	w := worker.NewStatsUpdater(repository.NewUserRepo(db), repository.NewTutorialRepo(db), repository.NewCommentRepo(db), repository.NewCategoryRepo(db), repository.NewTagRepo(db))
	w.RunOnce()
	var u domain.User
	db.First(&u, 1)
	if u.Level != domain.UserLevelCraftsman {
		t.Fatalf("level=%s", u.Level)
	}
}
func TestBug021VerificationUserLevelRefreshRegression(t *testing.T) {
	db := bug21DB(t)
	var u domain.User
	if err := db.Create(&u).Error; err != nil {
		t.Fatal(err)
	}
}
