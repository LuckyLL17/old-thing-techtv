package verification

import (
	"fmt"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"testing"
	"time"
	"upcycle-hub/config"
	"upcycle-hub/internal/domain"
	"upcycle-hub/internal/repository"
	"upcycle-hub/internal/service"
)

func bug4DB(t *testing.T) *gorm.DB {
	dsn := fmt.Sprintf("file:bug4_%d?mode=memory&cache=shared", time.Now().UnixNano())
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
func TestBug004VerificationDisabledAccountCannotLogin(t *testing.T) {
	db := bug4DB(t)
	h, _ := bcrypt.GenerateFromPassword([]byte("Valid1"), bcrypt.MinCost)
	u := seedBugUser(t, db, 1, "alice", string(h))
	u.Status = 0
	if err := db.Save(u).Error; err != nil {
		t.Fatal(err)
	}
	svc := service.NewAuthService(repository.NewUserRepo(db), &config.JWTConfig{Secret: "x", ExpireHour: 1})
	if _, _, err := svc.Login("alice", "Valid1"); err == nil {
		t.Fatal("disabled account received a token")
	}
}
func TestBug004VerificationDisabledAccountRegression(t *testing.T) {
	db := bug4DB(t)
	h, _ := bcrypt.GenerateFromPassword([]byte("Valid1"), bcrypt.MinCost)
	seedBugUser(t, db, 1, "alice", string(h))
	if _, _, err := service.NewAuthService(repository.NewUserRepo(db), &config.JWTConfig{Secret: "x", ExpireHour: 1}).Login("alice", "Valid1"); err != nil {
		t.Fatal(err)
	}
}
