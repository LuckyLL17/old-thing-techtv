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

func bug6DB(t *testing.T) *gorm.DB {
	dsn := fmt.Sprintf("file:bug6_%d?mode=memory&cache=shared", time.Now().UnixNano())
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	if err := db.AutoMigrate(&domain.Favorite{}); err != nil {
		t.Fatal(err)
	}
	sqlDB, err := db.DB()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = sqlDB.Close() })
	return db
}
func TestBug006VerificationFavoriteTypeFilter(t *testing.T) {
	db := bug6DB(t)
	r := repository.NewFavoriteRepo(db)
	for _, f := range []*domain.Favorite{{UserID: 1, TargetType: domain.FavTypeTutorial, TargetID: 1}, {UserID: 1, TargetType: domain.FavTypeProject, TargetID: 2}} {
		if err := db.Create(f).Error; err != nil {
			t.Fatal(err)
		}
	}
	list, total, err := r.ListByUser(1, domain.FavTypeTutorial, 1, 20)
	if err != nil {
		t.Fatal(err)
	}
	if total != 1 || len(list) != 1 || list[0].TargetType != domain.FavTypeTutorial {
		t.Fatalf("target_type filter mismatch: expected tutorial favorites, total=%d list=%+v", total, list)
	}
}
func TestBug006VerificationFavoriteTypeFilterRegression(t *testing.T) {
	db := bug6DB(t)
	if err := repository.NewFavoriteRepo(db).Create(&domain.Favorite{UserID: 2, TargetType: domain.FavTypeProject, TargetID: 9}); err != nil {
		t.Fatal(err)
	}
}
