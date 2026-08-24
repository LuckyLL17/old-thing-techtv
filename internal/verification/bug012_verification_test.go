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

func bug12DB(t *testing.T) *gorm.DB {
	dsn := fmt.Sprintf("file:bug12_%d?mode=memory&cache=shared", time.Now().UnixNano())
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	if err := db.AutoMigrate(&domain.Comment{}, &domain.Project{}, &domain.Tutorial{}); err != nil {
		t.Fatal(err)
	}
	sqlDB, err := db.DB()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = sqlDB.Close() })
	return db
}
func TestBug012VerificationProjectCommentCounter(t *testing.T) {
	db := bug12DB(t)
	pr := repository.NewProjectRepo(db)
	p := &domain.Project{ID: 5, UserID: 1, TutorialID: 2, Title: "p", Status: 1}
	if err := db.Create(p).Error; err != nil {
		t.Fatal(err)
	}
	svc := service.NewInteractionService(repository.NewCommentRepo(db), nil, nil, nil, nil, repository.NewTutorialRepo(db), pr)
	c, err := svc.Comment(1, domain.CommentTypeProject, 5, "great", 0)
	if err != nil {
		t.Fatal(err)
	}
	var got domain.Project
	db.First(&got, 5)
	if got.CommentCount != 1 {
		t.Fatalf("after create=%d", got.CommentCount)
	}
	if err := svc.DeleteComment(c.ID, 1); err != nil {
		t.Fatal(err)
	}
	db.First(&got, 5)
	if got.CommentCount != 0 {
		t.Fatalf("after delete=%d", got.CommentCount)
	}
}
func TestBug012VerificationProjectCommentCounterRegression(t *testing.T) {
	db := bug12DB(t)
	if err := repository.NewCommentRepo(db).Create(&domain.Comment{UserID: 1, TargetType: domain.CommentTypeProject, TargetID: 5, Content: "ok", Status: 1}); err != nil {
		t.Fatal(err)
	}
}
