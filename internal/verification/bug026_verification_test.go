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

func bug26DB(t *testing.T) *gorm.DB {
	dsn := fmt.Sprintf("file:bug26_%d?mode=memory&cache=shared", time.Now().UnixNano())
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
func TestBug026VerificationProjectCreateCounters(t *testing.T) {
	db := bug26DB(t)
	db.Create(&domain.User{ID: 1, Username: "u", Email: "u@e.com", PasswordHash: "x", Status: 1})
	db.Create(&domain.Category{ID: 1, Code: "c", Name: "C", Status: 1})
	db.Create(&domain.Tutorial{ID: 2, UserID: 1, CategoryID: 1, Title: "t", Status: domain.TutorialStatusPublished})
	svc := service.NewProjectService(repository.NewProjectRepo(db), repository.NewTutorialRepo(db), repository.NewUserRepo(db))
	if _, err := svc.Create(&service.ProjectCreateReq{UserID: 1, TutorialID: 2, Title: "p", Rating: 3}); err != nil {
		t.Fatal(err)
	}
	var trow domain.Tutorial
	var u domain.User
	db.First(&trow, 2)
	db.First(&u, 1)
	if trow.ProjectCount != 1 || u.ProjectCount != 1 || u.Score != 50 {
		t.Fatalf("tutorial=%d user projects=%d score=%d", trow.ProjectCount, u.ProjectCount, u.Score)
	}
}
func TestBug026VerificationProjectCreateCountersRegression(t *testing.T) {
	db := bug26DB(t)
	if err := repository.NewProjectRepo(db).Create(&domain.Project{UserID: 1, TutorialID: 2, Title: "p", Status: 1}); err != nil {
		t.Fatal(err)
	}
}
