package home

import (
	"context"
	"encoding/json"
	"log"
	"net"
	"net/http"
	"sync"
	"time"

	"pitchlake-backend/db/repositories"
	"pitchlake-backend/server/api/utils"
	"pitchlake-backend/server/types"

	"github.com/coder/websocket"
)

func (router *HomeRouter) SubscribeHome(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	var mu sync.Mutex
	var c *websocket.Conn
	var closed bool

	//allowedOrigin := os.Getenv("APP_URL")
	// Accept the WebSocket connection
	c2, err := websocket.Accept(w, r, &websocket.AcceptOptions{
		InsecureSkipVerify: true,
	})
	if err != nil {
		return err
	}
	defer c2.Close(websocket.StatusInternalError, "Internal server error")

	// Read the first message to get the subscription data

	s := &types.SubscriberHome{
		Msgs: make(chan []byte, router.subscriberMessageBuffer),
		CloseSlow: func() {
			mu.Lock()
			defer mu.Unlock()
			closed = true
			if c != nil {
				c.Close(websocket.StatusPolicyViolation, "connection too slow to keep up with messages")
			}
		},
	}

	// Add the subscriber to the appropriate map based on the address
	router.AddSubscriberHome(s)

	defer router.DeleteSubscriberHome(s)

	log.Printf("Subscribed to home")
	mu.Lock()
	if closed {
		mu.Unlock()
		return net.ErrClosed
	}
	c = c2
	mu.Unlock()
	defer c.CloseNow()

	vaultRepo := repositories.NewVaultRepository(&router.pool)
	vaultAddresses, err := vaultRepo.GetVaultAddresses(ctx)
	if err != nil {
		return err
	}
	// Send initial payload here
	response := struct {
		VaultAddresses []string `json:"vaultAddresses"`
	}{
		VaultAddresses: vaultAddresses,
	}
	jsonPayload, err := json.Marshal(response)
	if err != nil {
		return err
	}

	utils.WriteTimeout(ctx, time.Second*5, c, jsonPayload)

	for {
		select {
		case msg := <-s.Msgs:
			//Loop to write update messages to client
			err := utils.WriteTimeout(ctx, time.Second*5, c, msg)
			if err != nil {
				return err
			}
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

func (router *HomeRouter) AddSubscriberHome(s *types.SubscriberHome) {

	router.Subscribers.mux.Lock()
	router.Subscribers.List[s] = struct{}{}
	router.Subscribers.mux.Unlock()
}

func (router *HomeRouter) DeleteSubscriberHome(s *types.SubscriberHome) {

	router.Subscribers.mux.Lock()
	delete(router.Subscribers.List, s)
	router.Subscribers.mux.Unlock()
}
