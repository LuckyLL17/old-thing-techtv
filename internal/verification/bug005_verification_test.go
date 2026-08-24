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

func bug5DB(t *testing.T) *gorm.DB {
	dsn := fmt.Sprintf("file:bug5_%d?mode=memory&cache=shared", time.Now().UnixNano())
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	if err := db.AutoMigrate(&domain.Favorite{}, &domain.Tutorial{}); err != nil {
		t.Fatal(err)
	}
	sqlDB, err := db.DB()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = sqlDB.Close() })
	return db
}
func TestBug005VerificationFavoriteCountConsistency(t *testing.T) {
	db := bug5DB(t)
	tr := repository.NewTutorialRepo(db)
	trow := &domain.Tutorial{ID: 7, UserID: 1, CategoryID: 1, Title: "t", Status: domain.TutorialStatusPublished}
	if err := db.Create(trow).Error; err != nil {
		t.Fatal(err)
	}
	svc := service.NewInteractionService(nil, repository.NewFavoriteRepo(db), nil, nil, nil, tr, nil)
	on, err := svc.ToggleFavorite(2, domain.FavTypeTutorial, 7)
	if err != nil || !on {
		t.Fatalf("on: %v %v", on, err)
	}
	on, err = svc.ToggleFavorite(2, domain.FavTypeTutorial, 7)
	if err != nil || on {
		t.Fatalf("off: %v %v", on, err)
	}
	var n int64
	db.Model(&domain.Favorite{}).Where("user_id=2").Count(&n)
	if n != 0 {
		t.Fatalf("favorite rows=%d", n)
	}
	var got domain.Tutorial
	db.First(&got, 7)
	if got.FavoriteCount != 0 {
		t.Fatalf("favorite count=%d", got.FavoriteCount)
	}
}
func TestBug005VerificationFavoriteCountRegression(t *testing.T) {
	db := bug5DB(t)
	if err := repository.NewFavoriteRepo(db).Create(&domain.Favorite{UserID: 1, TargetType: domain.FavTypeProject, TargetID: 3}); err != nil {
		t.Fatal(err)
	}
}
