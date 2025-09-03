package home

import (
	"testing"

	"pitchlake-backend/server/types"
)

func TestAddSubscriberHome(t *testing.T) {
	router := &HomeRouter{
		Subscribers: SubscribersWithLock{
			List: make(map[*types.SubscriberHome]struct{}),
		},
	}

	subscriber := &types.SubscriberHome{
		Msgs: make(chan []byte, 10),
	}

	// Add subscriber
	router.Subscribers.mux.Lock()
	router.Subscribers.List[subscriber] = struct{}{}
	router.Subscribers.mux.Unlock()

	if len(router.Subscribers.List) != 1 {
		t.Errorf("Expected 1 subscriber, got %d", len(router.Subscribers.List))
	}
}

func TestRemoveSubscriberHome(t *testing.T) {
	router := &HomeRouter{
		Subscribers: SubscribersWithLock{
			List: make(map[*types.SubscriberHome]struct{}),
		},
	}

	subscriber := &types.SubscriberHome{
		Msgs: make(chan []byte, 10),
	}

	// Add subscriber first
	router.Subscribers.mux.Lock()
	router.Subscribers.List[subscriber] = struct{}{}
	router.Subscribers.mux.Unlock()

	if len(router.Subscribers.List) != 1 {
		t.Errorf("Expected 1 subscriber after add, got %d", len(router.Subscribers.List))
	}

	// Remove subscriber
	router.Subscribers.mux.Lock()
	delete(router.Subscribers.List, subscriber)
	router.Subscribers.mux.Unlock()

	if len(router.Subscribers.List) != 0 {
		t.Errorf("Expected 0 subscribers after delete, got %d", len(router.Subscribers.List))
	}
}

func TestHomeSubscriberConcurrency(t *testing.T) {
	router := &HomeRouter{
		Subscribers: SubscribersWithLock{
			List: make(map[*types.SubscriberHome]struct{}),
		},
	}

	// Add multiple subscribers
	for i := 0; i < 3; i++ {
		subscriber := &types.SubscriberHome{
			Msgs: make(chan []byte, 10),
		}
		router.Subscribers.mux.Lock()
		router.Subscribers.List[subscriber] = struct{}{}
		router.Subscribers.mux.Unlock()
	}

	if len(router.Subscribers.List) != 3 {
		t.Errorf("Expected 3 subscribers, got %d", len(router.Subscribers.List))
	}
}
