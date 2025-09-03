package home

import (
	"context"
	"encoding/json"
	"log"
	"net"
	"net/http"
	"sync"
	"time"

	"pitchlake-backend/betterdb/repositories"
	"pitchlake-backend/server/ws/utils"

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

	s := &subscriberHome{
		msgs: make(chan []byte, router.subscriberMessageBuffer),
		closeSlow: func() {
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
		case msg := <-s.msgs:
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

func (router *HomeRouter) AddSubscriberHome(s *subscriberHome) {

	router.subscribers.mux.Lock()
	router.subscribers.list[s] = struct{}{}
	router.subscribers.mux.Unlock()
}

func (router *HomeRouter) DeleteSubscriberHome(s *subscriberHome) {

	router.subscribers.mux.Lock()
	delete(router.subscribers.list, s)
	router.subscribers.mux.Unlock()
}
