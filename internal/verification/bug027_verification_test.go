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

func bug27DB(t *testing.T) *gorm.DB {
	dsn := fmt.Sprintf("file:bug27_%d?mode=memory&cache=shared", time.Now().UnixNano())
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	if err := db.AutoMigrate(&domain.User{}, &domain.Category{}, &domain.Tutorial{}, &domain.Project{}); err != nil {
		t.Fatal(err)
	}
	sqlDB, err := db.DB()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = sqlDB.Close() })
	return db
}
func TestBug027VerificationProjectDeleteCounters(t *testing.T) {
	db := bug27DB(t)
	db.Create(&domain.User{ID: 1, Username: "u", Email: "u@e.com", PasswordHash: "x", Status: 1, ProjectCount: 1, Score: 50})
	db.Create(&domain.Category{ID: 1, Code: "c", Name: "C", Status: 1})
	db.Create(&domain.Tutorial{ID: 2, UserID: 1, CategoryID: 1, Title: "t", Status: domain.TutorialStatusPublished, ProjectCount: 1})
	db.Create(&domain.Project{ID: 3, UserID: 1, TutorialID: 2, Title: "p", Rating: 3, Status: 1})
	svc := service.NewProjectService(repository.NewProjectRepo(db), repository.NewTutorialRepo(db), repository.NewUserRepo(db))
	if err := svc.Delete(3, 1); err != nil {
		t.Fatal(err)
	}
	var trow domain.Tutorial
	var u domain.User
	db.First(&trow, 2)
	db.First(&u, 1)
	if trow.ProjectCount != 0 || u.ProjectCount != 0 || u.Score != 0 {
		t.Fatalf("tutorial=%d projects=%d score=%d", trow.ProjectCount, u.ProjectCount, u.Score)
	}
}
func TestBug027VerificationProjectDeleteCountersRegression(t *testing.T) {
	db := bug27DB(t)
	if err := repository.NewProjectRepo(db).Delete(99); err != nil {
		t.Fatal(err)
	}
}
