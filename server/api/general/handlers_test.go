package general

import (
	"log"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestNewGeneralRouter(t *testing.T) {
	serveMux := http.NewServeMux()
	logger := log.Default()
	router := NewGeneralRouter(serveMux, logger)

	if router == nil {
		t.Error("Expected router to be created")
		return
	}
	if router.Subscribers.List == nil {
		t.Error("Expected Subscribers.List to be initialized")
	}
}

func TestHealthCheckHandler(t *testing.T) {
	logger := log.Default()
	router := &GeneralRouter{log: logger}

	req, err := http.NewRequest("GET", "/health", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	router.healthCheckHandler(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", rr.Code)
	}
	if rr.Body.String() != "OK" {
		t.Errorf("Expected body 'OK', got '%s'", rr.Body.String())
	}
}

func TestSubscribeGasDataHandler(t *testing.T) {
	logger := log.Default()
	router := &GeneralRouter{log: logger}

	req, err := http.NewRequest("GET", "/subscribeGas", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	router.subscribeGasDataHandler(rr, req)

	// The handler expects WebSocket upgrade, so it will fail with regular HTTP
	// This is expected behavior
	if rr.Code == http.StatusOK {
		t.Error("Expected non-OK status for non-WebSocket request")
	}
}
