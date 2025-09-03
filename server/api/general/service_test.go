package general

import (
	"testing"

	"pitchlake-backend/server/types"
)

func TestAddSubscriberGas(t *testing.T) {
	router := &GeneralRouter{
		Subscribers: SubscribersWithLock{
			List: make(map[*types.SubscriberGas]struct{}),
		},
	}

	subscriber := &types.SubscriberGas{
		Msgs:           make(chan []byte, 10),
		StartTimestamp: 1000,
		EndTimestamp:   2000,
		RoundDuration:  960,
	}

	router.addSubscriberGas(subscriber)

	if len(router.Subscribers.List) != 1 {
		t.Errorf("Expected 1 subscriber, got %d", len(router.Subscribers.List))
	}
}

func TestDeleteSubscriberGas(t *testing.T) {
	router := &GeneralRouter{
		Subscribers: SubscribersWithLock{
			List: make(map[*types.SubscriberGas]struct{}),
		},
	}

	subscriber := &types.SubscriberGas{
		Msgs:           make(chan []byte, 10),
		StartTimestamp: 1000,
		EndTimestamp:   2000,
		RoundDuration:  960,
	}

	// Add subscriber first
	router.addSubscriberGas(subscriber)
	if len(router.Subscribers.List) != 1 {
		t.Errorf("Expected 1 subscriber after add, got %d", len(router.Subscribers.List))
	}

	// Remove subscriber
	router.deleteSubscriberGas(subscriber)
	if len(router.Subscribers.List) != 0 {
		t.Errorf("Expected 0 subscribers after delete, got %d", len(router.Subscribers.List))
	}
}

func TestSubscriberGasConcurrency(t *testing.T) {
	router := &GeneralRouter{
		Subscribers: SubscribersWithLock{
			List: make(map[*types.SubscriberGas]struct{}),
		},
	}

	// Add multiple subscribers
	for i := 0; i < 3; i++ {
		subscriber := &types.SubscriberGas{
			Msgs:           make(chan []byte, 10),
			StartTimestamp: uint64(1000 + i),
			EndTimestamp:   uint64(2000 + i),
			RoundDuration:  960,
		}
		router.addSubscriberGas(subscriber)
	}

	if len(router.Subscribers.List) != 3 {
		t.Errorf("Expected 3 subscribers, got %d", len(router.Subscribers.List))
	}
}
