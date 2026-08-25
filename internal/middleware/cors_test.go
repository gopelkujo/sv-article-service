package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/gopelkujo/sv-article-service/internal/middleware"
)

func TestCORS_AllowsConfiguredOrigin(t *testing.T) {
	t.Parallel()

	next := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	handler := middleware.CORS([]string{"https://sv-article-client-react.vercel.app"})(next)

	req := httptest.NewRequest(http.MethodGet, "/article/10/0", nil)
	req.Header.Set("Origin", "https://sv-article-client-react.vercel.app")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)
	require.Equal(t, "https://sv-article-client-react.vercel.app", rec.Header().Get("Access-Control-Allow-Origin"))
}

func TestCORS_Preflight(t *testing.T) {
	t.Parallel()

	called := false
	next := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		called = true
		w.WriteHeader(http.StatusOK)
	})
	handler := middleware.CORS([]string{"https://sv-article-client-react.vercel.app"})(next)

	req := httptest.NewRequest(http.MethodOptions, "/article/", nil)
	req.Header.Set("Origin", "https://sv-article-client-react.vercel.app")
	req.Header.Set("Access-Control-Request-Method", "POST")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	require.Equal(t, http.StatusNoContent, rec.Code)
	require.False(t, called)
	require.Equal(t, "https://sv-article-client-react.vercel.app", rec.Header().Get("Access-Control-Allow-Origin"))
	require.Contains(t, rec.Header().Get("Access-Control-Allow-Methods"), "POST")
}

func TestCORS_RejectsUnknownOrigin(t *testing.T) {
	t.Parallel()

	next := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	handler := middleware.CORS([]string{"https://sv-article-client-react.vercel.app"})(next)

	req := httptest.NewRequest(http.MethodGet, "/article/10/0", nil)
	req.Header.Set("Origin", "https://evil.example")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)
	require.Empty(t, rec.Header().Get("Access-Control-Allow-Origin"))
}
