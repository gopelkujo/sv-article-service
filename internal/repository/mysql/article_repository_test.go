package mysql_test

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/gopelkujo/sv-article-service/internal/config"
	"github.com/gopelkujo/sv-article-service/internal/domain"
	"github.com/gopelkujo/sv-article-service/internal/repository/mysql"
)

// TestArticleRepositoryCRUD exercises the MySQL repository against a real database.
// Skipped unless RUN_DB_TESTS=1 (uses credentials from .env / environment).
func TestArticleRepositoryCRUD(t *testing.T) {
	if os.Getenv("RUN_DB_TESTS") != "1" {
		t.Skip("set RUN_DB_TESTS=1 to run MySQL repository tests")
	}

	cfg, err := config.Load()
	require.NoError(t, err)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	db, err := mysql.Open(ctx, cfg)
	require.NoError(t, err)
	defer func() { _ = db.Close() }()

	repo := mysql.NewArticleRepository(db)

	content := ""
	for len(content) < 200 {
		content += "repository integration content "
	}

	article := &domain.Article{
		Title:    "Repository smoke test title xx",
		Content:  content,
		Category: "tech",
		Status:   domain.StatusDraft,
	}

	require.NoError(t, repo.Create(ctx, article))
	require.NotZero(t, article.ID)
	require.False(t, article.CreatedDate.IsZero())

	got, err := repo.GetByID(ctx, article.ID)
	require.NoError(t, err)
	require.Equal(t, article.Title, got.Title)

	article.Title = "Repository smoke test title yy"
	article.Status = domain.StatusPublish
	require.NoError(t, repo.Update(ctx, article))
	require.Equal(t, domain.StatusPublish, article.Status)

	list, err := repo.List(ctx, 10, 0)
	require.NoError(t, err)
	require.NotEmpty(t, list)

	require.NoError(t, repo.Delete(ctx, article.ID))
	_, err = repo.GetByID(ctx, article.ID)
	require.ErrorIs(t, err, domain.ErrNotFound)
}
