package verification

import (
	"fmt"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"testing"
	"time"
	"upcycle-hub/config"
	"upcycle-hub/internal/domain"
	"upcycle-hub/internal/repository"
	"upcycle-hub/internal/service"
)

func bug1DB(t *testing.T) *gorm.DB {
	dsn := fmt.Sprintf("file:bug1_%d?mode=memory&cache=shared", time.Now().UnixNano())
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	if err := db.AutoMigrate(&domain.User{}); err != nil {
		t.Fatal(err)
	}
	sqlDB, err := db.DB()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = sqlDB.Close() })
	return db
}

func seedBugUser(t *testing.T, db *gorm.DB, id uint64, name string, hash string) *domain.User {
	u := &domain.User{ID: id, Username: name, Email: name + "@example.com", PasswordHash: hash, Nickname: name, Status: 1}
	if err := db.Create(u).Error; err != nil {
		t.Fatal(err)
	}
	return u
}
func TestBug001VerificationProfileUpdatePersistsAllFields(t *testing.T) {
	db := bug1DB(t)
	repo := repository.NewUserRepo(db)
	u := seedBugUser(t, db, 1, "alice", "hash")
	svc := service.NewAuthService(repo, &config.JWTConfig{Secret: "x", ExpireHour: 1})
	req := &domain.User{ID: u.ID, Nickname: "Alice New", Specialty: "woodwork", Bio: "long bio", Avatar: "avatar.png"}
	if err := svc.UpdateProfile(req); err != nil {
		t.Fatal(err)
	}
	got, err := svc.GetUserByID(u.ID)
	if err != nil {
		t.Fatal(err)
	}
	if got.Nickname != "Alice New" || got.Specialty != "woodwork" || got.Bio != "long bio" || got.Avatar != "avatar.png" {
		t.Fatalf("profile lost fields: %+v", got)
	}
}
func TestBug001VerificationProfileUpdateRegression(t *testing.T) {
	db := bug1DB(t)
	repo := repository.NewUserRepo(db)
	_ = seedBugUser(t, db, 1, "alice", "hash")
	if _, err := repo.GetByID(1); err != nil {
		t.Fatal(err)
	}
}
