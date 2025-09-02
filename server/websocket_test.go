package server

import (
	"context"
	"encoding/json"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/coder/websocket"
)

func TestWebSocketUpgrade(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	dbs := NewDBServer(ctx)
	defer dbs.cancel()

	// Create a test server
	server := httptest.NewServer(dbs)
	defer server.Close()

	// Test websocket upgrade for home subscription
	wsURL := "ws" + strings.TrimPrefix(server.URL, "http") + "/subscribeHome"
	ws, _, err := websocket.Dial(ctx, wsURL, nil)
	if err != nil {
		t.Fatalf("WebSocket connection failed: %v", err)
	}
	defer ws.Close(websocket.StatusNormalClosure, "")

	// Connection should be established
	if ws == nil {
		t.Fatal("WebSocket connection is nil")
	}
}

func TestWebSocketMessageValidation(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	dbs := NewDBServer(ctx)
	defer dbs.cancel()

	// Test message serialization/deserialization
	msg := subscriberMessage{
		Address:      "0x1234567890abcdef",
		VaultAddress: "0xfedcba0987654321",
		UserType:     "lp",
		OptionRound:  1,
	}

	// Test JSON marshaling
	jsonData, err := json.Marshal(msg)
	if err != nil {
		t.Fatalf("Failed to marshal message: %v", err)
	}
	if len(jsonData) == 0 {
		t.Fatal("Marshaled data is empty")
	}

	// Test JSON unmarshaling
	var unmarshaledMsg subscriberMessage
	err = json.Unmarshal(jsonData, &unmarshaledMsg)
	if err != nil {
		t.Fatalf("Failed to unmarshal message: %v", err)
	}
	if msg.Address != unmarshaledMsg.Address {
		t.Errorf("Address mismatch: expected %s, got %s", msg.Address, unmarshaledMsg.Address)
	}
	if msg.VaultAddress != unmarshaledMsg.VaultAddress {
		t.Errorf("VaultAddress mismatch: expected %s, got %s", msg.VaultAddress, unmarshaledMsg.VaultAddress)
	}
	if msg.UserType != unmarshaledMsg.UserType {
		t.Errorf("UserType mismatch: expected %s, got %s", msg.UserType, unmarshaledMsg.UserType)
	}
	if msg.OptionRound != unmarshaledMsg.OptionRound {
		t.Errorf("OptionRound mismatch: expected %d, got %d", msg.OptionRound, unmarshaledMsg.OptionRound)
	}
}

func TestInvalidWebSocketData(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	dbs := NewDBServer(ctx)
	defer dbs.cancel()

	// Test with invalid JSON data
	invalidJSON := []byte(`{"invalid": "json"`)

	// This should fail validation
	var msg subscriberMessage
	err := json.Unmarshal(invalidJSON, &msg)
	if err == nil {
		t.Error("Expected error when unmarshaling invalid JSON")
	}
}

func TestWebSocketSubscriptionFlow(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	dbs := NewDBServer(ctx)
	defer dbs.cancel()

	// Test that subscription adds subscriber correctly
	sub := &subscriberVault{
		address:      "0x123",
		vaultAddress: "0x456",
		userType:     "lp",
		msgs:         make(chan []byte, 16),
		closeSlow:    func() {},
	}

	// Add subscriber
	dbs.addSubscriberVault(sub)

	// Verify subscriber was added
	subscribers := dbs.subscribersVault["0x456"]
	if len(subscribers) != 1 {
		t.Errorf("Expected 1 subscriber, got %d", len(subscribers))
	}

	// Verify subscriber details
	if subscribers[0].address != "0x123" {
		t.Errorf("Expected address 0x123, got %s", subscribers[0].address)
	}
	if subscribers[0].userType != "lp" {
		t.Errorf("Expected userType lp, got %s", subscribers[0].userType)
	}

	// Clean up
	dbs.deleteSubscriberVault(sub)
}

func TestWebSocketMessageTypes(t *testing.T) {
	// Test different message types
	testCases := []struct {
		name    string
		msg     interface{}
		isValid bool
	}{
		{
			name: "Valid vault subscription",
			msg: subscriberMessage{
				Address:      "0x123",
				VaultAddress: "0x456",
				UserType:     "lp",
				OptionRound:  1,
			},
			isValid: true,
		},
		{
			name: "Valid gas request",
			msg: subscriberGasRequest{
				StartTimestamp: 1000,
				EndTimestamp:   2000,
				RoundDuration:  960,
			},
			isValid: true,
		},
		{
			name: "Invalid user type",
			msg: subscriberMessage{
				Address:      "0x123",
				VaultAddress: "0x456",
				UserType:     "invalid",
				OptionRound:  1,
			},
			isValid: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Test marshaling
			jsonData, err := json.Marshal(tc.msg)
			if err != nil {
				t.Fatalf("Failed to marshal message: %v", err)
			}

			// Test unmarshaling
			switch tc.msg.(type) {
			case subscriberMessage:
				var unmarshaledMsg subscriberMessage
				err = json.Unmarshal(jsonData, &unmarshaledMsg)
				if tc.isValid && err != nil {
					t.Errorf("Expected valid message but got error: %v", err)
				}
				if tc.isValid {
					// Validate user type
					if unmarshaledMsg.UserType != "lp" && unmarshaledMsg.UserType != "ob" {
						t.Errorf("Invalid user type: %s", unmarshaledMsg.UserType)
					}
				}
			case subscriberGasRequest:
				var unmarshaledMsg subscriberGasRequest
				err = json.Unmarshal(jsonData, &unmarshaledMsg)
				if tc.isValid && err != nil {
					t.Errorf("Expected valid message but got error: %v", err)
				}
			}
		})
	}
}
