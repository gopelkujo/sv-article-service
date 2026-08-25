package handler

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"

	"github.com/gopelkujo/sv-article-service/internal/domain"
	"github.com/gopelkujo/sv-article-service/internal/middleware"
	"github.com/gopelkujo/sv-article-service/internal/validator"
)

// ArticleServicer defines the service methods required by ArticleHandler.
type ArticleServicer interface {
	Create(ctx context.Context, input validator.ArticleInput) (*domain.Article, error)
	GetByID(ctx context.Context, id int) (*domain.Article, error)
	List(ctx context.Context, limit, offset int) ([]domain.Article, error)
	Update(ctx context.Context, id int, input validator.ArticleInput) (*domain.Article, error)
	Delete(ctx context.Context, id int) error
}

// ArticleHandler handles HTTP requests for article endpoints.
type ArticleHandler struct {
	service ArticleServicer
	logger  *slog.Logger
}

// NewArticleHandler constructs an ArticleHandler.
func NewArticleHandler(service ArticleServicer, logger *slog.Logger) *ArticleHandler {
	return &ArticleHandler{service: service, logger: logger}
}

type articleRequest struct {
	Title    string `json:"title"`
	Content  string `json:"content"`
	Category string `json:"category"`
	Status   string `json:"status"`
}

type articleResponse struct {
	ID          int        `json:"id"`
	Title       string     `json:"title"`
	Content     string     `json:"content"`
	Category    string     `json:"category"`
	CreatedDate time.Time  `json:"created_date"`
	UpdatedDate *time.Time `json:"updated_date"`
	Status      string     `json:"status"`
}

// Create handles POST /article/.
func (h *ArticleHandler) Create(w http.ResponseWriter, r *http.Request) {
	req, ok := decodeArticleRequest(w, r)
	if !ok {
		return
	}

	article, err := h.service.Create(r.Context(), toInput(req))
	if err != nil {
		h.writeServiceError(w, r, err)
		return
	}

	writeSuccess(w, http.StatusCreated, toResponse(article))
}

// List handles GET /article/{limit}/{offset}.
func (h *ArticleHandler) List(w http.ResponseWriter, r *http.Request) {
	limit, ok := parsePositivePathInt(w, r, "limit")
	if !ok {
		return
	}
	offset, ok := parseNonNegativePathInt(w, r, "offset")
	if !ok {
		return
	}

	articles, err := h.service.List(r.Context(), limit, offset)
	if err != nil {
		h.writeServiceError(w, r, err)
		return
	}

	items := make([]articleResponse, 0, len(articles))
	for i := range articles {
		items = append(items, toResponse(&articles[i]))
	}

	writeSuccess(w, http.StatusOK, items)
}

// GetByID handles GET /article/{id}.
func (h *ArticleHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	id, ok := parsePositivePathInt(w, r, "id")
	if !ok {
		return
	}

	article, err := h.service.GetByID(r.Context(), id)
	if err != nil {
		h.writeServiceError(w, r, err)
		return
	}

	writeSuccess(w, http.StatusOK, toResponse(article))
}

// Update handles PUT/PATCH /article/{id}.
func (h *ArticleHandler) Update(w http.ResponseWriter, r *http.Request) {
	id, ok := parsePositivePathInt(w, r, "id")
	if !ok {
		return
	}

	req, ok := decodeArticleRequest(w, r)
	if !ok {
		return
	}

	article, err := h.service.Update(r.Context(), id, toInput(req))
	if err != nil {
		h.writeServiceError(w, r, err)
		return
	}

	writeSuccess(w, http.StatusOK, toResponse(article))
}

// Delete handles DELETE /article/{id}.
func (h *ArticleHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id, ok := parsePositivePathInt(w, r, "id")
	if !ok {
		return
	}

	if err := h.service.Delete(r.Context(), id); err != nil {
		h.writeServiceError(w, r, err)
		return
	}

	writeSuccess(w, http.StatusOK, map[string]string{
		"message": "article deleted successfully",
	})
}

func (h *ArticleHandler) writeServiceError(w http.ResponseWriter, r *http.Request, err error) {
	if ve, ok := validator.AsValidationError(err); ok {
		writeError(w, http.StatusBadRequest, ve.Message, ve.Details)
		return
	}

	if errors.Is(err, domain.ErrNotFound) {
		writeError(w, http.StatusNotFound, "article not found", nil)
		return
	}

	h.logger.Error("request failed",
		"request_id", middleware.GetRequestID(r.Context()),
		"error", err,
	)
	writeError(w, http.StatusInternalServerError, "internal server error", nil)
}

func decodeArticleRequest(w http.ResponseWriter, r *http.Request) (articleRequest, bool) {
	defer func() { _ = r.Body.Close() }()

	body, err := io.ReadAll(io.LimitReader(r.Body, 1<<20))
	if err != nil {
		writeError(w, http.StatusBadRequest, "unable to read request body", nil)
		return articleRequest{}, false
	}
	if len(body) == 0 {
		writeError(w, http.StatusBadRequest, "request body is required", nil)
		return articleRequest{}, false
	}

	var req articleRequest
	if err := json.Unmarshal(body, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON body", []validator.FieldError{
			{Field: "body", Message: err.Error()},
		})
		return articleRequest{}, false
	}

	return req, true
}

func parsePositivePathInt(w http.ResponseWriter, r *http.Request, name string) (int, bool) {
	raw := chi.URLParam(r, name)
	value, err := strconv.Atoi(raw)
	if err != nil || value <= 0 {
		writeError(w, http.StatusBadRequest, "validation failed", []validator.FieldError{
			{Field: name, Message: "must be a positive integer"},
		})
		return 0, false
	}
	return value, true
}

func parseNonNegativePathInt(w http.ResponseWriter, r *http.Request, name string) (int, bool) {
	raw := chi.URLParam(r, name)
	value, err := strconv.Atoi(raw)
	if err != nil || value < 0 {
		writeError(w, http.StatusBadRequest, "validation failed", []validator.FieldError{
			{Field: name, Message: "must be a non-negative integer"},
		})
		return 0, false
	}
	return value, true
}

func toInput(req articleRequest) validator.ArticleInput {
	return validator.ArticleInput{
		Title:    req.Title,
		Content:  req.Content,
		Category: req.Category,
		Status:   req.Status,
	}
}

func toResponse(article *domain.Article) articleResponse {
	return articleResponse{
		ID:          article.ID,
		Title:       article.Title,
		Content:     article.Content,
		Category:    article.Category,
		CreatedDate: article.CreatedDate,
		UpdatedDate: article.UpdatedDate,
		Status:      article.Status,
	}
}
