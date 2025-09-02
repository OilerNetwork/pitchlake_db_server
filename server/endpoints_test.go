package server

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"
)

func init() {
	// Set up test database connection
	// Using local Homebrew PostgreSQL with trust authentication
	if os.Getenv("DB_URL") == "" {
		os.Setenv("DB_URL", "postgresql://localhost:5432/pitchlake_test")
	}
}

func TestHealthEndpoint(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	dbs := NewDBServer(ctx)
	defer dbs.cancel()

	req := httptest.NewRequest("GET", "/health", nil)
	w := httptest.NewRecorder()

	dbs.healthCheckHandler(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status OK, got %d", w.Code)
	}
}

func TestSubscribeHomeEndpoint(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	dbs := NewDBServer(ctx)
	defer dbs.cancel()

	req := httptest.NewRequest("GET", "/subscribeHome", nil)
	w := httptest.NewRecorder()

	dbs.subscribeHomeHandler(w, req)

	// Should not panic and should handle the request
	if w.Code == http.StatusInternalServerError {
		t.Error("Handler should not return internal server error")
	}
}

func TestSubscribeVaultEndpoint(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	dbs := NewDBServer(ctx)
	defer dbs.cancel()

	req := httptest.NewRequest("GET", "/subscribeVault", nil)
	w := httptest.NewRecorder()

	dbs.subscribeVaultHandler(w, req)

	// Should not panic and should handle the request
	if w.Code == http.StatusInternalServerError {
		t.Error("Handler should not return internal server error")
	}
}

func TestSubscribeGasEndpoint(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	dbs := NewDBServer(ctx)
	defer dbs.cancel()

	req := httptest.NewRequest("GET", "/subscribeGas", nil)
	w := httptest.NewRecorder()

	dbs.subscribeGasDataHandler(w, req)

	// Should not panic and should handle the request
	if w.Code == http.StatusInternalServerError {
		t.Error("Handler should not return internal server error")
	}
}
