// Package service implements article use cases and business rules.
package service

import (
	"context"
	"fmt"

	"github.com/gopelkujo/sv-article-service/internal/domain"
	"github.com/gopelkujo/sv-article-service/internal/validator"
)

// ArticleService orchestrates validation and persistence for articles.
type ArticleService struct {
	repo domain.ArticleRepository
}

// NewArticleService constructs an ArticleService backed by repo.
func NewArticleService(repo domain.ArticleRepository) *ArticleService {
	return &ArticleService{repo: repo}
}

// Create validates input and persists a new article.
func (s *ArticleService) Create(ctx context.Context, input validator.ArticleInput) (*domain.Article, error) {
	if err := validator.ValidateArticle(input); err != nil {
		return nil, err
	}

	article := &domain.Article{
		Title:    input.Title,
		Content:  input.Content,
		Category: input.Category,
		Status:   input.Status,
	}

	if err := s.repo.Create(ctx, article); err != nil {
		return nil, fmt.Errorf("create article: %w", err)
	}

	return article, nil
}

// GetByID returns a single article by id.
func (s *ArticleService) GetByID(ctx context.Context, id int) (*domain.Article, error) {
	if err := validator.ValidateID(id); err != nil {
		return nil, err
	}

	article, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("get article: %w", err)
	}

	return article, nil
}

// List returns a paginated list of articles.
func (s *ArticleService) List(ctx context.Context, limit, offset int) ([]domain.Article, error) {
	if err := validator.ValidatePagination(limit, offset); err != nil {
		return nil, err
	}

	articles, err := s.repo.List(ctx, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("list articles: %w", err)
	}

	return articles, nil
}

// Update validates input and replaces an existing article's mutable fields.
func (s *ArticleService) Update(ctx context.Context, id int, input validator.ArticleInput) (*domain.Article, error) {
	if err := validator.ValidateID(id); err != nil {
		return nil, err
	}
	if err := validator.ValidateArticle(input); err != nil {
		return nil, err
	}

	article := &domain.Article{
		ID:       id,
		Title:    input.Title,
		Content:  input.Content,
		Category: input.Category,
		Status:   input.Status,
	}

	if err := s.repo.Update(ctx, article); err != nil {
		return nil, fmt.Errorf("update article: %w", err)
	}

	return article, nil
}

// Delete removes an article by id.
func (s *ArticleService) Delete(ctx context.Context, id int) error {
	if err := validator.ValidateID(id); err != nil {
		return err
	}

	if err := s.repo.Delete(ctx, id); err != nil {
		return fmt.Errorf("delete article: %w", err)
	}

	return nil
}
