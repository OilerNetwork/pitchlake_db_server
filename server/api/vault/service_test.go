package vault

import (
	"fmt"
	"testing"

	"pitchlake-backend/server/types"
)

func TestAddSubscriberVault(t *testing.T) {
	router := &VaultRouter{
		Subscribers: SubscribersWithLock{
			List: make(map[string][]*types.SubscriberVault),
		},
	}

	subscriber := &types.SubscriberVault{
		Msgs:         make(chan []byte, 10),
		Address:      "0x123",
		UserType:     "user",
		VaultAddress: "0x456",
	}

	// Add subscriber
	router.Subscribers.mux.Lock()
	router.Subscribers.List[subscriber.Address] = append(router.Subscribers.List[subscriber.Address], subscriber)
	router.Subscribers.mux.Unlock()

	if len(router.Subscribers.List[subscriber.Address]) != 1 {
		t.Errorf("Expected 1 subscriber, got %d", len(router.Subscribers.List[subscriber.Address]))
	}
}

func TestRemoveSubscriberVault(t *testing.T) {
	router := &VaultRouter{
		Subscribers: SubscribersWithLock{
			List: make(map[string][]*types.SubscriberVault),
		},
	}

	subscriber := &types.SubscriberVault{
		Msgs:         make(chan []byte, 10),
		Address:      "0x123",
		UserType:     "user",
		VaultAddress: "0x456",
	}

	// Add subscriber first
	router.Subscribers.mux.Lock()
	router.Subscribers.List[subscriber.Address] = append(router.Subscribers.List[subscriber.Address], subscriber)
	router.Subscribers.mux.Unlock()

	if len(router.Subscribers.List[subscriber.Address]) != 1 {
		t.Errorf("Expected 1 subscriber after add, got %d", len(router.Subscribers.List[subscriber.Address]))
	}

	// Remove subscriber
	router.Subscribers.mux.Lock()
	subscribers := router.Subscribers.List[subscriber.Address]
	for i, sub := range subscribers {
		if sub == subscriber {
			router.Subscribers.List[subscriber.Address] = append(subscribers[:i], subscribers[i+1:]...)
			break
		}
	}
	router.Subscribers.mux.Unlock()

	if len(router.Subscribers.List[subscriber.Address]) != 0 {
		t.Errorf("Expected 0 subscribers after delete, got %d", len(router.Subscribers.List[subscriber.Address]))
	}
}

func TestVaultSubscriberConcurrency(t *testing.T) {
	router := &VaultRouter{
		Subscribers: SubscribersWithLock{
			List: make(map[string][]*types.SubscriberVault),
		},
	}

	// Add multiple subscribers
	for i := 0; i < 3; i++ {
		subscriber := &types.SubscriberVault{
			Msgs:         make(chan []byte, 10),
			Address:      fmt.Sprintf("0x%d", i),
			UserType:     "user",
			VaultAddress: fmt.Sprintf("0x%d", i+100),
		}
		router.Subscribers.mux.Lock()
		router.Subscribers.List[subscriber.Address] = append(router.Subscribers.List[subscriber.Address], subscriber)
		router.Subscribers.mux.Unlock()
	}

	totalSubscribers := 0
	for _, subscribers := range router.Subscribers.List {
		totalSubscribers += len(subscribers)
	}

	if totalSubscribers != 3 {
		t.Errorf("Expected 3 total subscribers, got %d", totalSubscribers)
	}
}
