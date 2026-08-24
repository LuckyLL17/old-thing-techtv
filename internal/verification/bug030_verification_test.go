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

func bug30DB(t *testing.T) *gorm.DB {
	dsn := fmt.Sprintf("file:bug30_%d?mode=memory&cache=shared", time.Now().UnixNano())
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	if err := db.AutoMigrate(&domain.Tag{}, &domain.TutorialTag{}); err != nil {
		t.Fatal(err)
	}
	sqlDB, err := db.DB()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = sqlDB.Close() })
	return db
}
func TestBug030VerificationDuplicateTags(t *testing.T) {
	db := bug30DB(t)
	r := repository.NewTagRepo(db)
	tags, err := r.UpsertByName([]string{"wood", "wood"})
	if err != nil {
		t.Fatal(err)
	}
	if len(tags) != 1 {
		t.Fatalf("upsert returned %d tags", len(tags))
	}
	if err := r.LinkTutorial(7, tags); err != nil {
		t.Fatal(err)
	}
	if err := r.LinkTutorial(7, tags); err != nil {
		t.Fatal(err)
	}
	var n int64
	db.Model(&domain.TutorialTag{}).Where("tutorial_id=7").Count(&n)
	if n != 1 {
		t.Fatalf("links=%d", n)
	}
}
func TestBug030VerificationDuplicateTagsRegression(t *testing.T) {
	db := bug30DB(t)
	if _, err := repository.NewTagRepo(db).Search("wood", 10); err != nil {
		t.Fatal(err)
	}
}
