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

func bug13DB(t *testing.T) *gorm.DB {
	dsn := fmt.Sprintf("file:bug13_%d?mode=memory&cache=shared", time.Now().UnixNano())
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	if err := db.AutoMigrate(&domain.Comment{}); err != nil {
		t.Fatal(err)
	}
	sqlDB, err := db.DB()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = sqlDB.Close() })
	return db
}
func TestBug013VerificationCommentParentFilter(t *testing.T) {
	db := bug13DB(t)
	r := repository.NewCommentRepo(db)
	db.Create(&domain.Comment{ID: 1, UserID: 1, TargetType: domain.CommentTypeTutorial, TargetID: 8, ParentID: 0, Content: "root", Status: 1})
	db.Create(&domain.Comment{ID: 2, UserID: 2, TargetType: domain.CommentTypeTutorial, TargetID: 8, ParentID: 1, Content: "reply", Status: 1})
	list, total, err := r.List(domain.CommentTypeTutorial, 8, 1, 20)
	if err != nil {
		t.Fatal(err)
	}
	if total != 1 || len(list) != 1 || list[0].ID != 1 {
		t.Fatalf("top level comments: total=%d list=%+v", total, list)
	}
}
func TestBug013VerificationCommentParentFilterRegression(t *testing.T) {
	db := bug13DB(t)
	if _, err := repository.NewCommentRepo(db).Count(domain.CommentTypeTutorial, 8); err != nil {
		t.Fatal(err)
	}
}
