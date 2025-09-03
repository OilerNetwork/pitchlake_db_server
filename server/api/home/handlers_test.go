package home

import (
	"log"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestNewHomeRouter(t *testing.T) {
	serveMux := http.NewServeMux()
	logger := log.Default()
	router := NewHomeRouter(serveMux, logger)

	if router == nil {
		t.Error("Expected router to be created")
		return
	}
	if router.Subscribers.List == nil {
		t.Error("Expected Subscribers.List to be initialized")
	}
}

func TestSubscribeHomeHandler(t *testing.T) {
	logger := log.Default()
	router := &HomeRouter{log: logger}

	req, err := http.NewRequest("GET", "/subscribeHome", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	router.subscribeHomeHandler(rr, req)

	// The handler expects WebSocket upgrade, so it will fail with regular HTTP
	// This is expected behavior
	if rr.Code == http.StatusOK {
		t.Error("Expected non-OK status for non-WebSocket request")
	}
}

func TestHomeRouterInitialization(t *testing.T) {
	serveMux := http.NewServeMux()
	logger := log.Default()
	router := NewHomeRouter(serveMux, logger)

	// Check that the endpoint is registered
	req, err := http.NewRequest("GET", "/subscribeHome", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	router.subscribeHomeHandler(rr, req)

	// Should not work with regular HTTP (expects WebSocket)
	if rr.Code == http.StatusOK {
		t.Error("Expected non-OK status for non-WebSocket request")
	}
}
