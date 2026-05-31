package app

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/donamo/todo-backend/internal/config"
)

func TestHealth(t *testing.T) {
	router := New(config.Config{
		FrontendURL:           "http://localhost:5173",
		HTTPShutdownTimeout:   10 * time.Second,
		HTTPReadHeaderTimeout: 5 * time.Second,
	}, nil)

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	if rec.Body.String() != "{\"status\":\"ok\"}\n" {
		t.Fatalf("body = %q", rec.Body.String())
	}
}
