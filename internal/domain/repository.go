package domain

import "context"

// ArticleRepository defines persistence operations for articles.
// Implementations must honor context cancellation and return
// ErrNotFound when a targeted row does not exist.
type ArticleRepository interface {
	// Create inserts a new article and populates ID and timestamps on article.
	Create(ctx context.Context, article *Article) error

	// GetByID returns a single article by primary key.
	GetByID(ctx context.Context, id int) (*Article, error)

	// List returns a page of articles ordered by id descending.
	List(ctx context.Context, limit, offset int) ([]Article, error)

	// Update replaces mutable fields of an existing article and refreshes timestamps.
	Update(ctx context.Context, article *Article) error

	// Delete removes an article by primary key.
	Delete(ctx context.Context, id int) error
}
