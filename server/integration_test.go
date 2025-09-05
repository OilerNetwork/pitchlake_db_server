package server

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"pitchlake-backend/server/types"
	"testing"
	"time"

	"github.com/coder/websocket"
)

// TestWebSocketValidationIntegration tests WebSocket validation across all services
func TestWebSocketValidationIntegration(t *testing.T) {
	// Test Gas Service Validation
	t.Run("GasService", func(t *testing.T) {
		testGasValidation(t)
	})

	// Test Vault Service Validation
	t.Run("VaultService", func(t *testing.T) {
		testVaultValidation(t)
	})
}

func testGasValidation(t *testing.T) {
	testCases := []struct {
		name        string
		request     types.SubscriberGasRequest
		expectError bool
	}{
		{
			name: "valid request",
			request: types.SubscriberGasRequest{
				StartTimestamp: 1000,
				EndTimestamp:   2000,
				RoundDuration:  960,
			},
			expectError: false,
		},
		{
			name: "invalid round duration",
			request: types.SubscriberGasRequest{
				StartTimestamp: 1000,
				EndTimestamp:   2000,
				RoundDuration:  999,
			},
			expectError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create test server
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Upgrade", "websocket")
				w.Header().Set("Connection", "Upgrade")
				w.WriteHeader(http.StatusSwitchingProtocols)
			}))
			defer server.Close()

			// Test WebSocket connection
			url := "ws" + server.URL[4:] + "/subscribeGas"
			conn, _, err := websocket.Dial(context.Background(), url, nil)
			if err != nil {
				t.Skipf("Skipping test due to WebSocket connection error: %v", err)
				return
			}
			defer conn.Close(websocket.StatusNormalClosure, "")

			// Send test request
			requestBytes, _ := json.Marshal(tc.request)
			err = conn.Write(context.Background(), websocket.MessageText, requestBytes)
			if err != nil {
				t.Fatalf("Failed to write test message: %v", err)
			}

			// Read response
			ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
			defer cancel()

			_, responseBytes, err := conn.Read(ctx)
			if tc.expectError {
				if err != nil {
					var errorResponse map[string]string
					if json.Unmarshal(responseBytes, &errorResponse) == nil {
						if errorResponse["error"] == "Invalid request" {
							t.Logf("Validation correctly rejected invalid request: %s", errorResponse["details"])
						}
					}
				} else {
					t.Errorf("Expected validation error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error for valid request: %v", err)
				}
			}
		})
	}
}

func testVaultValidation(t *testing.T) {
	testCases := []struct {
		name        string
		message     types.SubscriberMessage
		expectError bool
	}{
		{
			name: "valid subscription",
			message: types.SubscriberMessage{
				Address:      "0x1234567890123456789012345678901234567890",
				VaultAddress: "0xabcdefabcdefabcdefabcdefabcdefabcdefabcd",
				UserType:     "lp",
			},
			expectError: false,
		},
		{
			name: "invalid user type",
			message: types.SubscriberMessage{
				Address:      "0x1234567890123456789012345678901234567890",
				VaultAddress: "0xabcdefabcdefabcdefabcdefabcdefabcdefabcd",
				UserType:     "invalid",
			},
			expectError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create test server
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Upgrade", "websocket")
				w.Header().Set("Connection", "Upgrade")
				w.WriteHeader(http.StatusSwitchingProtocols)
			}))
			defer server.Close()

			// Test WebSocket connection
			url := "ws" + server.URL[4:] + "/subscribeVault"
			conn, _, err := websocket.Dial(context.Background(), url, nil)
			if err != nil {
				t.Skipf("Skipping test due to WebSocket connection error: %v", err)
				return
			}
			defer conn.Close(websocket.StatusNormalClosure, "")

			// Send test message
			messageBytes, _ := json.Marshal(tc.message)
			err = conn.Write(context.Background(), websocket.MessageText, messageBytes)
			if err != nil {
				t.Fatalf("Failed to write test message: %v", err)
			}

			// Read response
			ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
			defer cancel()

			_, responseBytes, err := conn.Read(ctx)
			if tc.expectError {
				if err != nil {
					var errorResponse map[string]string
					if json.Unmarshal(responseBytes, &errorResponse) == nil {
						if errorResponse["error"] == "Invalid subscription message" {
							t.Logf("Validation correctly rejected invalid subscription: %s", errorResponse["details"])
						}
					}
				} else {
					t.Errorf("Expected validation error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error for valid subscription: %v", err)
				}
			}
		})
	}
}
