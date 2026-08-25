package service_test

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/gopelkujo/sv-article-service/internal/domain"
	"github.com/gopelkujo/sv-article-service/internal/service"
	"github.com/gopelkujo/sv-article-service/internal/validator"
)

// mockArticleRepo is an in-memory ArticleRepository for unit tests.
type mockArticleRepo struct {
	byID    map[int]*domain.Article
	nextID  int
	createE error
	getE    error
	listE   error
	updateE error
	deleteE error
}

func newMockRepo() *mockArticleRepo {
	return &mockArticleRepo{
		byID:   make(map[int]*domain.Article),
		nextID: 1,
	}
}

func (m *mockArticleRepo) Create(_ context.Context, article *domain.Article) error {
	if m.createE != nil {
		return m.createE
	}
	clone := *article
	clone.ID = m.nextID
	m.nextID++
	now := time.Now()
	clone.CreatedDate = now
	m.byID[clone.ID] = &clone
	*article = clone
	return nil
}

func (m *mockArticleRepo) GetByID(_ context.Context, id int) (*domain.Article, error) {
	if m.getE != nil {
		return nil, m.getE
	}
	article, ok := m.byID[id]
	if !ok {
		return nil, domain.ErrNotFound
	}
	clone := *article
	return &clone, nil
}

func (m *mockArticleRepo) List(_ context.Context, limit, offset int) ([]domain.Article, error) {
	if m.listE != nil {
		return nil, m.listE
	}
	all := make([]domain.Article, 0, len(m.byID))
	for _, article := range m.byID {
		all = append(all, *article)
	}
	if offset >= len(all) {
		return []domain.Article{}, nil
	}
	end := offset + limit
	if end > len(all) {
		end = len(all)
	}
	return all[offset:end], nil
}

func (m *mockArticleRepo) Update(_ context.Context, article *domain.Article) error {
	if m.updateE != nil {
		return m.updateE
	}
	existing, ok := m.byID[article.ID]
	if !ok {
		return domain.ErrNotFound
	}
	existing.Title = article.Title
	existing.Content = article.Content
	existing.Category = article.Category
	existing.Status = article.Status
	now := time.Now()
	existing.UpdatedDate = &now
	*article = *existing
	return nil
}

func (m *mockArticleRepo) Delete(_ context.Context, id int) error {
	if m.deleteE != nil {
		return m.deleteE
	}
	if _, ok := m.byID[id]; !ok {
		return domain.ErrNotFound
	}
	delete(m.byID, id)
	return nil
}

func validInput() validator.ArticleInput {
	return validator.ArticleInput{
		Title:    "A valid article title!",
		Content:  strings.Repeat("c", 200),
		Category: "tech",
		Status:   domain.StatusDraft,
	}
}

func TestArticleService_Create_Success(t *testing.T) {
	t.Parallel()

	repo := newMockRepo()
	svc := service.NewArticleService(repo)

	article, err := svc.Create(context.Background(), validInput())
	require.NoError(t, err)
	require.Equal(t, 1, article.ID)
	require.Equal(t, "A valid article title!", article.Title)
	require.Equal(t, domain.StatusDraft, article.Status)
}

func TestArticleService_Create_ValidationError(t *testing.T) {
	t.Parallel()

	svc := service.NewArticleService(newMockRepo())
	_, err := svc.Create(context.Background(), validator.ArticleInput{Title: "x"})
	_, ok := validator.AsValidationError(err)
	require.True(t, ok)
}

func TestArticleService_Create_RepoError(t *testing.T) {
	t.Parallel()

	repo := newMockRepo()
	repo.createE = errors.New("db down")
	svc := service.NewArticleService(repo)

	_, err := svc.Create(context.Background(), validInput())
	require.Error(t, err)
	require.ErrorContains(t, err, "create article")
	require.ErrorContains(t, err, "db down")
}

func TestArticleService_GetByID(t *testing.T) {
	t.Parallel()

	repo := newMockRepo()
	svc := service.NewArticleService(repo)
	created, err := svc.Create(context.Background(), validInput())
	require.NoError(t, err)

	got, err := svc.GetByID(context.Background(), created.ID)
	require.NoError(t, err)
	require.Equal(t, created.ID, got.ID)

	_, err = svc.GetByID(context.Background(), 0)
	_, ok := validator.AsValidationError(err)
	require.True(t, ok)

	_, err = svc.GetByID(context.Background(), 999)
	require.ErrorIs(t, err, domain.ErrNotFound)
}

func TestArticleService_List(t *testing.T) {
	t.Parallel()

	repo := newMockRepo()
	svc := service.NewArticleService(repo)
	_, err := svc.Create(context.Background(), validInput())
	require.NoError(t, err)

	items, err := svc.List(context.Background(), 10, 0)
	require.NoError(t, err)
	require.Len(t, items, 1)

	_, err = svc.List(context.Background(), 0, 0)
	_, ok := validator.AsValidationError(err)
	require.True(t, ok)
}

func TestArticleService_Update(t *testing.T) {
	t.Parallel()

	repo := newMockRepo()
	svc := service.NewArticleService(repo)
	created, err := svc.Create(context.Background(), validInput())
	require.NoError(t, err)

	input := validInput()
	input.Title = "Updated article title!!"
	input.Status = domain.StatusPublish

	updated, err := svc.Update(context.Background(), created.ID, input)
	require.NoError(t, err)
	require.Equal(t, "Updated article title!!", updated.Title)
	require.Equal(t, domain.StatusPublish, updated.Status)
	require.NotNil(t, updated.UpdatedDate)

	_, err = svc.Update(context.Background(), 999, input)
	require.ErrorIs(t, err, domain.ErrNotFound)
}

func TestArticleService_Delete(t *testing.T) {
	t.Parallel()

	repo := newMockRepo()
	svc := service.NewArticleService(repo)
	created, err := svc.Create(context.Background(), validInput())
	require.NoError(t, err)

	require.NoError(t, svc.Delete(context.Background(), created.ID))
	_, err = svc.GetByID(context.Background(), created.ID)
	require.ErrorIs(t, err, domain.ErrNotFound)

	err = svc.Delete(context.Background(), created.ID)
	require.ErrorIs(t, err, domain.ErrNotFound)
}
