package handler_test

import (
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/gopelkujo/sv-article-service/internal/handler"
)

func TestHealthz_Liveness(t *testing.T) {
	t.Parallel()

	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	health := handler.NewHealthHandler(nil)
	// Article handler is unused for this route but required by NewRouter.
	article := handler.NewArticleHandler(nil, logger)
	router := handler.NewRouter(article, health, logger, []string{"http://localhost:5173"})

	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)

	var body envelope
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))
	require.True(t, body.Success)

	var data map[string]string
	require.NoError(t, json.Unmarshal(body.Data, &data))
	require.Equal(t, "ok", data["status"])
}
