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

func bug8DB(t *testing.T) *gorm.DB {
	dsn := fmt.Sprintf("file:bug8_%d?mode=memory&cache=shared", time.Now().UnixNano())
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
func TestBug008VerificationFollowIdempotency(t *testing.T) {
	db := bug8DB(t)
	r := repository.NewFollowRepo(db)
	svc := service.NewInteractionService(nil, nil, nil, r, nil, nil, nil)
	if err := svc.Follow(1, 2); err != nil {
		t.Fatal(err)
	}
	if err := svc.Follow(1, 2); err != nil {
		t.Fatal(err)
	}
	var n int64
	db.Model(&domain.Follow{}).Where("follower_id=1 AND following_id=2").Count(&n)
	if n != 1 {
		t.Fatalf("duplicate follow rows=%d", n)
	}
}
func TestBug008VerificationFollowIdempotencyRegression(t *testing.T) {
	db := bug8DB(t)
	if err := repository.NewFollowRepo(db).Unfollow(1, 2); err != nil {
		t.Fatal(err)
	}
}
