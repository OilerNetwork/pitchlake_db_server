package server

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/coder/websocket"
)

func (dbs *dbServer) subscribeVault(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	var mu sync.Mutex
	var c *websocket.Conn
	var closed bool
	//Extract address from the request and add here

	allowedOrigin := os.Getenv("APP_URL")
	c2, err := websocket.Accept(w, r, &websocket.AcceptOptions{
		OriginPatterns: []string{allowedOrigin},
	})
	if err != nil {
		return err
	}
	defer c2.Close(websocket.StatusInternalError, "Internal server error")

	// Read the first message to get the subscription data
	_, msg, err := c2.Read(ctx)
	if err != nil {
		return err
	}

	var sm subscriberMessage
	err = json.Unmarshal(msg, &sm)
	if err != nil {
		return err
	}
	log.Printf("%v", sm)

	s := &subscriberVault{
		address:      sm.Address,
		vaultAddress: sm.VaultAddress,
		userType:     sm.UserType,
		msgs:         make(chan []byte, dbs.subscriberMessageBuffer),
		closeSlow: func() {
			mu.Lock()
			defer mu.Unlock()
			closed = true
			if c != nil {
				c.Close(websocket.StatusPolicyViolation, "connection too slow to keep up with messages")
			}
		},
	}
	dbs.addSubscriberVault(s)
	defer dbs.deleteSubscriberVault(s)

	mu.Lock()
	if closed {
		mu.Unlock()
		return net.ErrClosed
	}
	c = c2
	mu.Unlock()
	defer c.CloseNow()

	//Send initial payload here
	var payload webSocketPayload

	payload.PayloadType = "initial"
	vaultState, err := dbs.db.GetVaultStateByID(s.vaultAddress)
	if err != nil {
		return err
	}
	optionRounds, err := dbs.db.GetOptionRoundsByVaultAddress(s.vaultAddress)
	if err != nil {
		return err
	}
	payload.OptionRoundStates = optionRounds
	// if sm.OptionRound != 0 {
	// 	optionRoundState, err := dbs.db.GetOptionRoundByID(sm.OptionRound)
	// 	if err != nil {
	// 		return err
	// 	}
	// 	vaultSubscription.OptionRoundState = *optionRoundState
	// } else {
	// 	optionRoundState, err := dbs.db.GetOptionRoundByAddress(vaultState.CurrentRoundAddress)
	// 	if err != nil {
	// 		return err
	// 	}
	// 	vaultSubscription.OptionRoundState = *optionRoundState
	// }
	//@note replace this to fetch all option rounds for the vault

	payload.VaultState = *vaultState

	if sm.UserType == "lp" {

		lpState, err := dbs.db.GetLiquidityProviderStateByAddress(s.address)
		if err != nil {
			fmt.Printf("Error fetching lp state %v", err)
		} else {
			payload.LiquidityProviderState = *lpState
		}
	} else if sm.UserType == "ob" {

		obState, err := dbs.db.GetOptionBuyerByAddress(s.address)
		if err != nil {
			return err
		}
		payload.OptionBuyerState = *obState
	} else {
		return errors.New("invalid user type")
	}

	// Marshal the VaultState to a JSON byte array
	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	dbs.writeTimeout(ctx, time.Second*5, c, jsonPayload)
	go func() {
		for {
			_, msg, err := c.Read(ctx)
			if err != nil {
				log.Printf("Error reading message: %v", err)
				return
			}
			log.Printf("Received message from client: %s", msg)
			//Unmarshall the json here and send the updates respectively
			s.msgs <- []byte("RECEIVED")
			log.Printf("Client Info %v", s.address)
			// Handle the received message here
		}
	}()

	dbs.writeTimeout(ctx, time.Second*5, c, jsonPayload)
	for {
		select {
		case msg := <-s.msgs:
			//Push messages received on the subscriber channel to the client
			err := dbs.writeTimeout(ctx, time.Second*5, c, msg)
			if err != nil {
				return err
			}
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

func (dbs *dbServer) subscribeHome(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	var mu sync.Mutex
	var c *websocket.Conn
	var closed bool

	allowedOrigin := os.Getenv("APP_URL")
	// Accept the WebSocket connection
	c2, err := websocket.Accept(w, r, &websocket.AcceptOptions{
		OriginPatterns: []string{allowedOrigin},
	})
	if err != nil {
		return err
	}
	defer c2.Close(websocket.StatusInternalError, "Internal server error")

	// Read the first message to get the subscription data

	s := &subscriberHome{
		msgs: make(chan []byte, dbs.subscriberMessageBuffer),
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
	dbs.addSubscriberHome(s)

	defer dbs.deleteSubscriberHome(s)

	log.Printf("Subscribed to home")
	mu.Lock()
	if closed {
		mu.Unlock()
		return net.ErrClosed
	}
	c = c2
	mu.Unlock()
	defer c.CloseNow()

	//@dev Use if need to read data from the channel
	// go func() {
	// 	for {
	// 		_, msg, err := c.Read(ctx)
	// 		if err != nil {
	// 			log.Printf("Error reading message: %v", err)
	// 			return
	// 		}
	// 		log.Printf("Received message from client: %s", msg)
	// 		s.msgs <- []byte("RECEIVED")
	// 		// Handle the received message here
	// 	}
	// }()

	vaultAddresses, err := dbs.db.GetVaultAddresses()
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

	dbs.writeTimeout(ctx, time.Second*5, c, jsonPayload)

	for {
		select {
		case msg := <-s.msgs:
			//Loop to write update messages to client
			err := dbs.writeTimeout(ctx, time.Second*5, c, msg)
			if err != nil {
				return err
			}
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

func (dbs *dbServer) writeTimeout(ctx context.Context, timeout time.Duration, c *websocket.Conn, msg []byte) error {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	return c.Write(ctx, websocket.MessageText, msg)
}
