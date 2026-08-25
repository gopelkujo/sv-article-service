package mysql

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/gopelkujo/sv-article-service/internal/domain"
)

// Ensure ArticleRepository satisfies domain.ArticleRepository at compile time.
var _ domain.ArticleRepository = (*ArticleRepository)(nil)

// ArticleRepository persists articles in the MySQL posts table.
type ArticleRepository struct {
	db *sql.DB
}

// NewArticleRepository constructs an ArticleRepository backed by db.
func NewArticleRepository(db *sql.DB) *ArticleRepository {
	return &ArticleRepository{db: db}
}

const articleColumns = `id, title, content, category, created_date, updated_date, status`

// Create inserts a new article and reloads generated fields into article.
func (r *ArticleRepository) Create(ctx context.Context, article *domain.Article) error {
	const query = `
		INSERT INTO posts (title, content, category, status)
		VALUES (?, ?, ?, ?)`

	result, err := r.db.ExecContext(ctx, query,
		article.Title,
		article.Content,
		article.Category,
		article.Status,
	)
	if err != nil {
		return fmt.Errorf("mysql create article: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("mysql create article last insert id: %w", err)
	}

	created, err := r.GetByID(ctx, int(id))
	if err != nil {
		return fmt.Errorf("mysql create article reload: %w", err)
	}

	*article = *created
	return nil
}

// GetByID returns a single article by id.
func (r *ArticleRepository) GetByID(ctx context.Context, id int) (*domain.Article, error) {
	query := `SELECT ` + articleColumns + ` FROM posts WHERE id = ? LIMIT 1`

	article, err := scanArticle(r.db.QueryRowContext(ctx, query, id))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("mysql get article by id: %w", err)
	}

	return article, nil
}

// List returns articles paginated by limit and offset, newest first.
func (r *ArticleRepository) List(ctx context.Context, limit, offset int) ([]domain.Article, error) {
	query := `SELECT ` + articleColumns + `
		FROM posts
		ORDER BY id DESC
		LIMIT ? OFFSET ?`

	rows, err := r.db.QueryContext(ctx, query, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("mysql list articles: %w", err)
	}
	defer func() { _ = rows.Close() }()

	articles := make([]domain.Article, 0)
	for rows.Next() {
		article, err := scanArticle(rows)
		if err != nil {
			return nil, fmt.Errorf("mysql list articles scan: %w", err)
		}
		articles = append(articles, *article)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("mysql list articles rows: %w", err)
	}

	return articles, nil
}

// Update replaces title, content, category, and status for an existing article.
func (r *ArticleRepository) Update(ctx context.Context, article *domain.Article) error {
	const query = `
		UPDATE posts
		SET title = ?, content = ?, category = ?, status = ?
		WHERE id = ?`

	result, err := r.db.ExecContext(ctx, query,
		article.Title,
		article.Content,
		article.Category,
		article.Status,
		article.ID,
	)
	if err != nil {
		return fmt.Errorf("mysql update article: %w", err)
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("mysql update article rows affected: %w", err)
	}
	if affected == 0 {
		return domain.ErrNotFound
	}

	updated, err := r.GetByID(ctx, article.ID)
	if err != nil {
		return fmt.Errorf("mysql update article reload: %w", err)
	}

	*article = *updated
	return nil
}

// Delete removes an article by id.
func (r *ArticleRepository) Delete(ctx context.Context, id int) error {
	const query = `DELETE FROM posts WHERE id = ?`

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("mysql delete article: %w", err)
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("mysql delete article rows affected: %w", err)
	}
	if affected == 0 {
		return domain.ErrNotFound
	}

	return nil
}

type rowScanner interface {
	Scan(dest ...any) error
}

func scanArticle(scanner rowScanner) (*domain.Article, error) {
	var (
		article     domain.Article
		updatedDate sql.NullTime
	)

	err := scanner.Scan(
		&article.ID,
		&article.Title,
		&article.Content,
		&article.Category,
		&article.CreatedDate,
		&updatedDate,
		&article.Status,
	)
	if err != nil {
		return nil, err
	}

	if updatedDate.Valid {
		t := updatedDate.Time
		article.UpdatedDate = &t
	}

	return &article, nil
}
