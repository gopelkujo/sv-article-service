package handler_test

import (
	"bytes"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/gopelkujo/sv-article-service/internal/config"
	"github.com/gopelkujo/sv-article-service/internal/handler"
	"github.com/gopelkujo/sv-article-service/internal/repository/mysql"
	"github.com/gopelkujo/sv-article-service/internal/service"
)

type envelope struct {
	Success bool            `json:"success"`
	Data    json.RawMessage `json:"data"`
	Error   *struct {
		Message string `json:"message"`
		Details []any  `json:"details"`
	} `json:"error"`
}

type articleData struct {
	ID       int    `json:"id"`
	Title    string `json:"title"`
	Content  string `json:"content"`
	Category string `json:"category"`
	Status   string `json:"status"`
}

// TestArticleHTTP_CRUDHappyPath exercises create → get → update → delete
// through the HTTP layer against a real MySQL database.
// Skipped unless RUN_DB_TESTS=1.
func TestArticleHTTP_CRUDHappyPath(t *testing.T) {
	if os.Getenv("RUN_DB_TESTS") != "1" {
		t.Skip("set RUN_DB_TESTS=1 to run HTTP integration tests")
	}

	cfg, err := config.Load()
	require.NoError(t, err)

	ctx := t.Context()
	db, err := mysql.Open(ctx, cfg)
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })

	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	repo := mysql.NewArticleRepository(db)
	svc := service.NewArticleService(repo)
	h := handler.NewArticleHandler(svc, logger)
	health := handler.NewHealthHandler(db)
	router := handler.NewRouter(h, health, logger, []string{"http://localhost:5173"})

	content := strings.Repeat("integration content ", 20)
	require.GreaterOrEqual(t, len(content), 200)

	createBody := map[string]string{
		"title":    "Integration test article!",
		"content":  content,
		"category": "tech",
		"status":   "draft",
	}

	// Create
	createResp := doJSON(t, router, http.MethodPost, "/article/", createBody)
	require.Equal(t, http.StatusCreated, createResp.Code)

	var createdEnv envelope
	require.NoError(t, json.Unmarshal(createResp.Body.Bytes(), &createdEnv))
	require.True(t, createdEnv.Success)

	var created articleData
	require.NoError(t, json.Unmarshal(createdEnv.Data, &created))
	require.NotZero(t, created.ID)
	require.Equal(t, "draft", created.Status)

	id := strconv.Itoa(created.ID)

	// Get
	getResp := doJSON(t, router, http.MethodGet, "/article/"+id, nil)
	require.Equal(t, http.StatusOK, getResp.Code)

	var gotEnv envelope
	require.NoError(t, json.Unmarshal(getResp.Body.Bytes(), &gotEnv))
	require.True(t, gotEnv.Success)

	var got articleData
	require.NoError(t, json.Unmarshal(gotEnv.Data, &got))
	require.Equal(t, created.ID, got.ID)
	require.Equal(t, "Integration test article!", got.Title)

	// Update
	updateBody := map[string]string{
		"title":    "Integration updated title!",
		"content":  content,
		"category": "news",
		"status":   "publish",
	}
	updateResp := doJSON(t, router, http.MethodPut, "/article/"+id, updateBody)
	require.Equal(t, http.StatusOK, updateResp.Code)

	var updatedEnv envelope
	require.NoError(t, json.Unmarshal(updateResp.Body.Bytes(), &updatedEnv))
	require.True(t, updatedEnv.Success)

	var updated articleData
	require.NoError(t, json.Unmarshal(updatedEnv.Data, &updated))
	require.Equal(t, "Integration updated title!", updated.Title)
	require.Equal(t, "publish", updated.Status)
	require.Equal(t, "news", updated.Category)

	// Delete
	deleteResp := doJSON(t, router, http.MethodDelete, "/article/"+id, nil)
	require.Equal(t, http.StatusOK, deleteResp.Code)

	var deleteEnv envelope
	require.NoError(t, json.Unmarshal(deleteResp.Body.Bytes(), &deleteEnv))
	require.True(t, deleteEnv.Success)

	// Confirm gone
	missingResp := doJSON(t, router, http.MethodGet, "/article/"+id, nil)
	require.Equal(t, http.StatusNotFound, missingResp.Code)

	var missingEnv envelope
	require.NoError(t, json.Unmarshal(missingResp.Body.Bytes(), &missingEnv))
	require.False(t, missingEnv.Success)
	require.NotNil(t, missingEnv.Error)
	require.Equal(t, "article not found", missingEnv.Error.Message)
}

func doJSON(t *testing.T, router http.Handler, method, path string, body any) *httptest.ResponseRecorder {
	t.Helper()

	var reader io.Reader
	if body != nil {
		payload, err := json.Marshal(body)
		require.NoError(t, err)
		reader = bytes.NewReader(payload)
	}

	req := httptest.NewRequest(method, path, reader)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	req = req.WithContext(t.Context())

	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	return rec
}
