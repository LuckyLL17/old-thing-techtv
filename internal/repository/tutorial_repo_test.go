package repository

import (
	"testing"

	"upcycle-hub/internal/domain"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// newTestDB spins up an in-memory SQLite database with the tutorial table and a
// couple of rows whose view/favorite counts make the popularity ordering
// unambiguous: a high-view/low-favorite row and a low-view/high-favorite row.
func newTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	if err := db.AutoMigrate(&domain.Tutorial{}); err != nil {
		t.Fatalf("automigrate: %v", err)
	}
	rows := []*domain.Tutorial{
		{Title: "high-view-low-fav", Status: domain.TutorialStatusPublished, ViewCount: 1000, FavoriteCount: 2},
		{Title: "low-view-high-fav", Status: domain.TutorialStatusPublished, ViewCount: 5, FavoriteCount: 500},
	}
	if err := db.Create(&rows).Error; err != nil {
		t.Fatalf("seed: %v", err)
	}
	return db
}

// TestListPopularOrdersByFavoriteFirst locks in the business-heat ranking for
// the "popular" sort: favorite_count is the primary key and view_count only
// breaks ties. Previously this case ordered by "view_count DESC, favorite_count
// DESC", which surfaced high-view / low-favorite tutorials above high-favorite
// ones — the exact inversion reported on the hot tutorials page.
func TestListPopularOrdersByFavoriteFirst(t *testing.T) {
	repo := NewTutorialRepo(newTestDB(t))

	list, _, err := repo.List(1, 10, 0, "", domain.TutorialStatusPublished, "popular", "", 0)
	if err != nil {
		t.Fatalf("List popular: %v", err)
	}
	if len(list) != 2 {
		t.Fatalf("expected 2 rows, got %d", len(list))
	}
	if got, want := list[0].Title, "low-view-high-fav"; got != want {
		t.Fatalf("popular sort put %q (favorite_count=%d) first; want the high-favorite tutorial %q",
			got, list[0].FavoriteCount, want)
	}
	if list[0].FavoriteCount < list[1].FavoriteCount {
		t.Fatalf("results not in favorite_count DESC order: %d before %d",
			list[0].FavoriteCount, list[1].FavoriteCount)
	}
}

// TestListViewsSortStillByPrimaryKey confirms the dedicated "views" sort keeps
// view_count primary — the fix must not have leaked the favorite-first ordering
// into the views-only ranking.
func TestListViewsSortStillByPrimaryKey(t *testing.T) {
	repo := NewTutorialRepo(newTestDB(t))

	list, _, err := repo.List(1, 10, 0, "", domain.TutorialStatusPublished, "views", "", 0)
	if err != nil {
		t.Fatalf("List views: %v", err)
	}
	if len(list) != 2 {
		t.Fatalf("expected 2 rows, got %d", len(list))
	}
	if got, want := list[0].Title, "high-view-low-fav"; got != want {
		t.Fatalf("views sort put %q (view_count=%d) first; want the high-view tutorial %q",
			got, list[0].ViewCount, want)
	}
}
