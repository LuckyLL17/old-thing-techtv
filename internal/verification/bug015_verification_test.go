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

func bug15DB(t *testing.T) *gorm.DB {
	dsn := fmt.Sprintf("file:bug15_%d?mode=memory&cache=shared", time.Now().UnixNano())
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	if err := db.AutoMigrate(&domain.Attempt{}); err != nil {
		t.Fatal(err)
	}
	sqlDB, err := db.DB()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = sqlDB.Close() })
	return db
}
func TestBug015VerificationAttemptPagination(t *testing.T) {
	db := bug15DB(t)
	r := repository.NewAttemptRepo(db)
	for i := 0; i < 4; i++ {
		db.Create(&domain.Attempt{UserID: 1, TutorialID: uint64(i + 1)})
	}
	a, total, err := r.ListByUser(1, 1, 2)
	if err != nil {
		t.Fatal(err)
	}
	b, _, err := r.ListByUser(1, 2, 2)
	if err != nil {
		t.Fatal(err)
	}
	if total != 4 || len(a) != 2 || len(b) != 2 || a[1].ID == b[0].ID {
		t.Fatalf("pages overlap total=%d a=%v b=%v", total, a, b)
	}
}
func TestBug015VerificationAttemptPaginationRegression(t *testing.T) {
	db := bug15DB(t)
	if _, _, err := repository.NewAttemptRepo(db).CountByUser(1); err != nil {
		t.Fatal(err)
	}
}
