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

func bug3DB(t *testing.T) *gorm.DB {
	dsn := fmt.Sprintf("file:bug3_%d?mode=memory&cache=shared", time.Now().UnixNano())
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
func TestBug003VerificationDuplicateRegistration(t *testing.T) {
	db := bug3DB(t)
	svc := service.NewAuthService(repository.NewUserRepo(db), &config.JWTConfig{Secret: "x", ExpireHour: 1})
	if _, err := svc.Register("alice", "same@example.com", "Valid1"); err != nil {
		t.Fatal(err)
	}
	if _, err := svc.Register("alice2", "same@example.com", "Valid1"); err == nil {
		t.Fatal("duplicate email accepted")
	}
	var n int64
	db.Model(&domain.User{}).Where("email = ?", "same@example.com").Count(&n)
	if n != 1 {
		t.Fatalf("duplicate rows=%d", n)
	}
}
func TestBug003VerificationDuplicateRegistrationRegression(t *testing.T) {
	db := bug3DB(t)
	svc := service.NewAuthService(repository.NewUserRepo(db), &config.JWTConfig{Secret: "x", ExpireHour: 1})
	if _, err := svc.Register("alice", "a@example.com", "Valid1"); err != nil {
		t.Fatal(err)
	}
}
