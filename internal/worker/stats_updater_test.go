package worker

import (
	"fmt"
	"testing"
	"upcycle-hub/internal/domain"
	"upcycle-hub/internal/repository"
	"upcycle-hub/internal/service"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func newWorkerTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	if err := db.AutoMigrate(
		&domain.User{}, &domain.Category{}, &domain.Tag{},
		&domain.Tutorial{}, &domain.TutorialTag{}, &domain.Step{},
		&domain.Material{}, &domain.Tool{}, &domain.Project{},
		&domain.Comment{}, &domain.Favorite{}, &domain.Attempt{},
		&domain.Follow{}, &domain.Message{}, &domain.Notification{},
		&domain.AuditLog{},
	); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	return db
}

// runLevelChain drives the full IncStats → cron → UserCenter path for a user
// whose score has reached the given value, and returns the level shown on the
// personal page.
func runLevelChain(t *testing.T, score int) string {
	t.Helper()
	db := newWorkerTestDB(t)
	ur := repository.NewUserRepo(db)
	tr := repository.NewTutorialRepo(db)
	pr := repository.NewProjectRepo(db)
	cr := repository.NewCategoryRepo(db)
	fr := repository.NewFavoriteRepo(db)
	ar := repository.NewAttemptRepo(db)
	tgr := repository.NewTagRepo(db)
	commr := repository.NewCommentRepo(db)

	u := &domain.User{
		Username: "alice", Email: "a@x.com", PasswordHash: "x",
		Level: domain.UserLevelNovice, Score: 0, Status: 1,
	}
	if err := ur.Create(u); err != nil {
		t.Fatalf("create: %v", err)
	}
	if err := ur.IncStats(u.ID, 0, 0, score); err != nil {
		t.Fatalf("incstats: %v", err)
	}

	w := NewStatsUpdater(ur, tr, commr, cr, tgr)
	w.refreshUserLevels()

	statsSvc := service.NewStatsService(tr, pr, ur, cr, fr, ar)
	st, err := statsSvc.UserCenter(u.ID)
	if err != nil {
		t.Fatalf("usercenter: %v", err)
	}
	if st.Score != score {
		t.Fatalf("score lost across refresh: got %d want %d", st.Score, score)
	}
	return st.Level
}

func TestLevelChain_ExactThresholds(t *testing.T) {
	cases := []struct {
		score int
		want  string
	}{
		{199, domain.UserLevelNovice},
		{200, domain.UserLevelApprentice}, // >= 200
		{1000, domain.UserLevelCraftsman}, // threshold reached → must upgrade
		{1001, domain.UserLevelCraftsman},
		{5000, domain.UserLevelMaster}, // >= 5000
		{4999, domain.UserLevelCraftsman},
	}
	for _, c := range cases {
		got := runLevelChain(t, c.score)
		if got != c.want {
			t.Errorf("score=%d: personal page level=%q want %q", c.score, got, c.want)
		}
	}
}

func TestLevelChain_PaginationCoversAll(t *testing.T) {
	db := newWorkerTestDB(t)
	ur := repository.NewUserRepo(db)
	tr := repository.NewTutorialRepo(db)
	cr := repository.NewCategoryRepo(db)
	tgr := repository.NewTagRepo(db)
	commr := repository.NewCommentRepo(db)

	const n = 250
	for i := 0; i < n; i++ {
		u := &domain.User{
			Username: fmt.Sprintf("u%03d", i), Email: fmt.Sprintf("e%03d@x.com", i),
			PasswordHash: "x", Level: domain.UserLevelNovice, Score: 1500, Status: 1,
		}
		if err := ur.Create(u); err != nil {
			t.Fatalf("create %d: %v", i, err)
		}
	}

	w := NewStatsUpdater(ur, tr, commr, cr, tgr)
	w.refreshUserLevels()

	stale := 0
	for i := 0; i < n; i++ {
		got, err := ur.GetByEmail(fmt.Sprintf("e%03d@x.com", i))
		if err != nil {
			t.Fatalf("get %d: %v", i, err)
		}
		if got.Level != domain.UserLevelCraftsman {
			stale++
		}
	}
	if stale != 0 {
		t.Fatalf("%d users still at stale level after refresh (pagination missed some?)", stale)
	}
}
