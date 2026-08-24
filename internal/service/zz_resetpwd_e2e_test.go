package service

import (
	"testing"
	"upcycle-hub/config"
	"upcycle-hub/internal/domain"
	"upcycle-hub/internal/repository"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// TestResetPasswordThenLogin reproduces the reported bug:
// after a successful reset, logging in with the NEW password must succeed
// and the OLD password must no longer work.
func TestResetPasswordThenLogin(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	if err := db.AutoMigrate(&domain.User{}); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	repo := repository.NewUserRepo(db)
	svc := NewAuthService(repo, &config.JWTConfig{Secret: "s", ExpireHour: 1})

	// register
	const oldPwd, newPwd = "OldPwd123", "NewPwd456"
	u, err := svc.Register("alice", "alice@example.com", oldPwd)
	if err != nil {
		t.Fatalf("register: %v", err)
	}

	// 1) reset password
	if err := svc.ResetPassword(u.ID, oldPwd, newPwd); err != nil {
		t.Fatalf("reset should succeed, got: %v", err)
	}

	// 2) login with NEW password — must succeed
	if _, _, err := svc.Login("alice", newPwd); err != nil {
		t.Fatalf("login with NEW password should succeed, got: %v", err)
	}

	// 3) login with OLD password — must fail
	if _, _, err := svc.Login("alice", oldPwd); err == nil {
		t.Fatalf("login with OLD password should fail after reset")
	}
}

// TestResetPasswordWrongOld verifies the old-password check still rejects.
func TestResetPasswordWrongOld(t *testing.T) {
	db, _ := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	_ = db.AutoMigrate(&domain.User{})
	repo := repository.NewUserRepo(db)
	svc := NewAuthService(repo, &config.JWTConfig{Secret: "s", ExpireHour: 1})
	u, _ := svc.Register("bob", "bob@example.com", "Secret123")
	if err := svc.ResetPassword(u.ID, "WrongOldPwd", "Another999"); err == nil {
		t.Fatalf("reset with wrong old password should fail")
	}
}
