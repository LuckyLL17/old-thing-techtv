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

func bug7DB(t *testing.T) *gorm.DB {
	dsn := fmt.Sprintf("file:bug7_%d?mode=memory&cache=shared", time.Now().UnixNano())
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	if err := db.AutoMigrate(&domain.Follow{}); err != nil {
		t.Fatal(err)
	}
	sqlDB, err := db.DB()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = sqlDB.Close() })
	return db
}
func TestBug007VerificationFollowCountsDirection(t *testing.T) {
	db := bug7DB(t)
	r := repository.NewFollowRepo(db)
	db.Create(&domain.Follow{FollowerID: 1, FollowingID: 9})
	db.Create(&domain.Follow{FollowerID: 2, FollowingID: 9})
	db.Create(&domain.Follow{FollowerID: 9, FollowingID: 3})
	followers, following, err := service.NewInteractionService(nil, nil, nil, r, nil, nil, nil).FollowCounts(9)
	if err != nil {
		t.Fatal(err)
	}
	if followers != 2 || following != 1 {
		t.Fatalf("followers=%d following=%d", followers, following)
	}
}
func TestBug007VerificationFollowCountsRegression(t *testing.T) {
	db := bug7DB(t)
	if _, _, err := repository.NewFollowRepo(db).Counts(42); err != nil {
		t.Fatal(err)
	}
}
