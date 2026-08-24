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

func bug22DB(t *testing.T) *gorm.DB {
	dsn := fmt.Sprintf("file:bug22_%d?mode=memory&cache=shared", time.Now().UnixNano())
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	if err := db.AutoMigrate(&domain.TutorialVersion{}); err != nil {
		t.Fatal(err)
	}
	sqlDB, err := db.DB()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = sqlDB.Close() })
	return db
}
func TestBug022VerificationLatestVersion(t *testing.T) {
	db := bug22DB(t)
	r := repository.NewTutorialVersionRepo(db)
	db.Create(&domain.TutorialVersion{TutorialID: 3, Version: 1, Title: "old"})
	db.Create(&domain.TutorialVersion{TutorialID: 3, Version: 2, Title: "new"})
	got, err := r.GetLatest(3)
	if err != nil {
		t.Fatal(err)
	}
	if got.Version != 2 || got.Title != "new" {
		t.Fatalf("latest=%+v", got)
	}
}
func TestBug022VerificationLatestVersionRegression(t *testing.T) {
	db := bug22DB(t)
	if _, _, err := repository.NewTutorialVersionRepo(db).ListByTutorial(3, 1, 10); err != nil {
		t.Fatal(err)
	}
}
