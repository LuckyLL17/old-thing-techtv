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

func bug25DB(t *testing.T) *gorm.DB {
	dsn := fmt.Sprintf("file:bug25_%d?mode=memory&cache=shared", time.Now().UnixNano())
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	if err := db.AutoMigrate(&domain.AuditLog{}); err != nil {
		t.Fatal(err)
	}
	sqlDB, err := db.DB()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = sqlDB.Close() })
	return db
}
func TestBug025VerificationAuditStatsWindow(t *testing.T) {
	db := bug25DB(t)
	now := time.Now()
	db.Create(&domain.AuditLog{UserID: 1, Action: "login", CreatedAt: now.Add(-24 * time.Hour)})
	db.Create(&domain.AuditLog{UserID: 1, Action: "login", CreatedAt: now.Add(-30 * 24 * time.Hour)})
	got, err := repository.NewAuditLogRepo(db).StatsByAction(7)
	if err != nil {
		t.Fatal(err)
	}
	if got["login"] != 1 {
		t.Fatalf("StatsByAction days window mismatch: created_at filter returned %v", got)
	}
}
func TestBug025VerificationAuditStatsRegression(t *testing.T) {
	db := bug25DB(t)
	if _, err := service.NewAuditService(repository.NewAuditLogRepo(db)).Stats(0); err != nil {
		t.Fatal(err)
	}
}
