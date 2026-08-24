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

func bug28DB(t *testing.T) *gorm.DB {
	dsn := fmt.Sprintf("file:bug28_%d?mode=memory&cache=shared", time.Now().UnixNano())
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	if err := db.AutoMigrate(&domain.Step{}); err != nil {
		t.Fatal(err)
	}
	sqlDB, err := db.DB()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = sqlDB.Close() })
	return db
}
func TestBug028VerificationStepReorder(t *testing.T) {
	db := bug28DB(t)
	r := repository.NewStepRepo(db)
	for i := uint64(1); i <= 3; i++ {
		db.Create(&domain.Step{ID: i, TutorialID: 7, StepOrder: int(i), Title: fmt.Sprintf("s%d", i), Content: "x"})
	}
	if err := r.UpdateOrder(7, []uint64{3, 1, 2}); err != nil {
		t.Fatal(err)
	}
	list, err := r.ListByTutorial(7)
	if err != nil {
		t.Fatal(err)
	}
	if len(list) != 3 || list[0].ID != 3 || list[0].StepOrder != 1 || list[2].StepOrder != 3 {
		t.Fatalf("steps=%+v", list)
	}
}
func TestBug028VerificationStepReorderRegression(t *testing.T) {
	db := bug28DB(t)
	if _, err := repository.NewStepRepo(db).ListByTutorial(7); err != nil {
		t.Fatal(err)
	}
}
