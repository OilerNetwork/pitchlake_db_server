package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net"
	"net/http"
	"pitchlake-backend/db"
	"pitchlake-backend/models"
	"sync"
	"time"

	"github.com/coder/websocket"
)

// dbServer enables broadcasting to a set of subscribers.
type dbServer struct {
	// subscriberMessageBuffer controls the max number
	// of messages that can be queued for a subscriber
	// before it is kicked.
	//
	// Defaults to 16.
	subscriberMessageBuffer int
	db                      *db.DB

	// publishLimiter controls the rate limit applied to the publish endpoint.
	//
	// Defaults to one publish every 100ms with a burst of 8.

	// logf controls where logs are sent.
	// Defaults to log.Printf.
	logf func(f string, v ...interface{})

	// serveMux routes the various endpoints to the appropriate handler.
	serveMux http.ServeMux

	subscribersVaultMu sync.Mutex
	subscribersVault   map[string]map[*subscriber]struct{}
	subscribersHomeMu  sync.Mutex
	subscribersHome    map[*subscriber]struct{}
	ctx                context.Context
	cancel             context.CancelFunc
}

// subscriber represents a subscriber.
// Messages are sent on the msgs channel and if the client
// cannot keep up with the messages, closeSlow is called.
type subscriber struct {
	msgs        chan []byte
	address     string
	userType    string
	optionRound uint64
	closeSlow   func()
}

type subscriberMessage struct {
	Address     string `json:"address"`
	UserType    string `json:"userType"`
	OptionRound uint64 `json:"optionRound"`
}

// newdbServer constructs a dbServer with the defaults.
// Create a custom context for the server here and pass it to the db package
func newDBServer(ctx context.Context) *dbServer {

	ctx, cancel := context.WithCancel(ctx)
	db := &db.DB{}
	connString := ""
	db.Init(connString)
	dbs := &dbServer{
		subscriberMessageBuffer: 16,
		logf:                    log.Printf,
		subscribersVault:        make(map[string]map[*subscriber]struct{}),
		subscribersHome:         make(map[*subscriber]struct{}),
		db:                      db,
		ctx:                     ctx,
		cancel:                  cancel,
	}
	dbs.serveMux.Handle("/", http.FileServer(http.Dir(".")))
	dbs.serveMux.HandleFunc("/subscribeHome", dbs.subscribeHomeHandler)
	dbs.serveMux.HandleFunc("/subscribeVault", dbs.subscribeVaultHandler)
	go dbs.listener()
	return dbs
}

func (dbs *dbServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	dbs.serveMux.ServeHTTP(w, r)
}

func (dbs *dbServer) subscribeHomeHandler(w http.ResponseWriter, r *http.Request) {
	println("Subscribing to home")
	err := dbs.subscribeHome(r.Context(), w, r)
	if errors.Is(err, context.Canceled) {
		return
	}
	if websocket.CloseStatus(err) == websocket.StatusNormalClosure ||
		websocket.CloseStatus(err) == websocket.StatusGoingAway {
		return
	}
	if err != nil {
		dbs.logf("%v", err)
		return
	}
}

// subscribeHandler accepts the WebSocket connection and then subscribes
// it to all future messages.
func (dbs *dbServer) subscribeVaultHandler(w http.ResponseWriter, r *http.Request) {
	err := dbs.subscribeVault(r.Context(), w, r)
	if errors.Is(err, context.Canceled) {
		return
	}
	if websocket.CloseStatus(err) == websocket.StatusNormalClosure ||
		websocket.CloseStatus(err) == websocket.StatusGoingAway {
		return
	}
	if err != nil {
		dbs.logf("%v", err)
		return
	}
}

// subscribe subscribes the given WebSocket to all broadcast messages.
// It creates a subscriber with a buffered msgs chan to give some room to slower
// connections and then registers the subscriber. It then listens for all messages
// and writes them to the WebSocket. If the context is cancelled or
// an error occurs, it returns and deletes the subscription.
//
// It uses CloseRead to keep reading from the connection to process control
// messages and cancel the context if the connection drops.
func (dbs *dbServer) subscribeVault(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	var mu sync.Mutex
	var c *websocket.Conn
	var closed bool
	//Extract address from the request and add here
	decoder := json.NewDecoder(r.Body)
	var sm subscriberMessage
	decoder.Decode(&sm)
	s := &subscriber{
		address:  sm.Address,
		userType: sm.UserType,
		msgs:     make(chan []byte, dbs.subscriberMessageBuffer),
		closeSlow: func() {
			mu.Lock()
			defer mu.Unlock()
			closed = true
			if c != nil {
				c.Close(websocket.StatusPolicyViolation, "connection too slow to keep up with messages")
			}
		},
	}

	dbs.addSubscriber(s, "Vault")
	defer dbs.deleteSubscriber(s, "Vault")

	c2, err := websocket.Accept(w, r, nil)
	if err != nil {
		return err
	}
	mu.Lock()
	if closed {
		mu.Unlock()
		return net.ErrClosed
	}
	c = c2
	mu.Unlock()
	defer c.CloseNow()

	ctx = c.CloseRead(ctx)
	//Send initial payload here
	var vaultSubscription models.VaultSubscription
	vaultState, err := dbs.db.GetVaultStateByID(s.address)
	if sm.OptionRound != 0 {
		optionRoundState, err := dbs.db.GetOptionRoundByID(sm.OptionRound)
		if err != nil {
			return err
		}
		vaultSubscription.OptionRoundState = *optionRoundState
	} else {
		optionRoundState, err := dbs.db.GetOptionRoundByAddress(vaultState.CurrentRoundAddress)
		if err != nil {
			return err
		}
		vaultSubscription.OptionRoundState = *optionRoundState
	}

	vaultSubscription.VaultState = *vaultState
	if err != nil {
		return err
	}

	if sm.UserType == "lp" {

		lpState, err := dbs.db.GetLiquidityProviderStateByAddress(s.address)
		if err != nil {
			return err
		}
		vaultSubscription.LiquidityProviderState = *lpState
	} else if sm.UserType == "ob" {

		obState, err := dbs.db.GetOptionBuyerByAddress(s.address)
		if err != nil {
			return err
		}
		vaultSubscription.OptionBuyerState = *obState
	} else {
		return errors.New("invalid user type")
	}

	// Marshal the VaultState to a JSON byte array
	jsonPayload, err := json.Marshal(vaultSubscription)
	if err != nil {
		return err
	}

	writeTimeout(ctx, time.Second*5, c, jsonPayload)
	for {
		select {
		case msg := <-s.msgs:
			err := writeTimeout(ctx, time.Second*5, c, msg)
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

	// Accept the WebSocket connection
	c2, err := websocket.Accept(w, r, &websocket.AcceptOptions{
		OriginPatterns: []string{"localhost:3000"},
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

	s := &subscriber{
		address:     sm.Address,
		userType:    sm.UserType,
		optionRound: sm.OptionRound,
		msgs:        make(chan []byte, dbs.subscriberMessageBuffer),
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
	dbs.addSubscriber(s, sm.Address)
	defer dbs.deleteSubscriber(s, sm.Address)

	log.Printf("Subscribed to home")
	mu.Lock()
	if closed {
		mu.Unlock()
		return net.ErrClosed
	}
	c = c2
	mu.Unlock()
	defer c.CloseNow()
	// Send initial payload here

	go func() {
		for {
			_, msg, err := c.Read(ctx)
			if err != nil {
				log.Printf("Error reading message: %v", err)
				return
			}
			log.Printf("Received message from client: %s", msg)
			s.msgs <- []byte("RECEIVED")
			log.Printf("Client Info %v", s.address)
			// Handle the received message here
		}
	}()

	// Send initial payload here
	writeTimeout(ctx, time.Second*5, c, []byte("subscribed"))

	for {
		select {
		case msg := <-s.msgs:
			log.Printf("HELLO")
			err := writeTimeout(ctx, time.Second*5, c, msg)
			if err != nil {
				return err
			}
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

func (dbs *dbServer) listener() {
	for {
		select {
		case <-dbs.ctx.Done():
			dbs.logf("Listener shutting down")
			return
		default:
			// Wait for a notification
			notification, err := dbs.db.Conn.WaitForNotification(dbs.ctx)
			if err != nil {
				if dbs.ctx.Err() != nil {
					// Context was canceled, exit the loop
					return
				}
				dbs.logf("Error waiting for notification: %v", err)
				continue
			}
			// Process notification here
			switch notification.Channel {
			case "lp_update":
				fmt.Println("Received an update on lp_row_update")
			case "vault_update":
				fmt.Println("Received an update on vault_update")
			case "state_transition":
				// Push this to all channels (without address as well)
				fmt.Println("Received an update on state_transition")
			case "ob_update":
				fmt.Println("Received an update on ob_update")
			case "or_update":
				fmt.Println("Received an update on or_update")
			default:
				fmt.Println("Received an update on unknown channel", notification.Channel)
			}
			dbs.publishAddress(notification.Channel, []byte(notification.Payload))
			dbs.publishAll([]byte(notification.Payload))
		}
	}
}

// publishUser sends a message to all subscribers of a specific address.
func (dbs *dbServer) publishAddress(address string, msg []byte) {
	dbs.subscribersVaultMu.Lock()
	defer dbs.subscribersVaultMu.Unlock()

	for s := range dbs.subscribersVault[address] {
		select {
		case s.msgs <- msg:
		default:
			go s.closeSlow()
		}
	}
}

// publishAll sends a message to all subscribers of all addresses.
func (dbs *dbServer) publishAll(msg []byte) {
	dbs.subscribersVaultMu.Lock()
	defer dbs.subscribersVaultMu.Unlock()

	for address := range dbs.subscribersVault {
		for s := range dbs.subscribersVault[address] {
			select {
			case s.msgs <- msg:
			default:
				go s.closeSlow()
			}
		}
	}
}

// addSubscriber registers a subscriber.
func (dbs *dbServer) addSubscriber(s *subscriber, subscriptionType string) {
	println("CP3")
	if subscriptionType == "Vault" {
		dbs.subscribersVaultMu.Lock()
		dbs.subscribersVault[s.address][s] = struct{}{}
		dbs.subscribersVaultMu.Unlock()
	} else if subscriptionType == "Home" {
		dbs.subscribersHomeMu.Lock()
		dbs.subscribersHome[s] = struct{}{}
		dbs.subscribersHomeMu.Unlock()
	}
}

// deleteSubscriber deletes the given subscriber.
func (dbs *dbServer) deleteSubscriber(s *subscriber, subscriptionType string) {
	if subscriptionType == "Vault" {
		dbs.subscribersVaultMu.Lock()
		delete(dbs.subscribersVault[s.address], s)
		dbs.subscribersVaultMu.Unlock()
	} else if subscriptionType == "Home" {
		dbs.subscribersHomeMu.Lock()
		delete(dbs.subscribersHome, s)
		dbs.subscribersHomeMu.Unlock()
	}

}

func writeTimeout(ctx context.Context, timeout time.Duration, c *websocket.Conn, msg []byte) error {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	return c.Write(ctx, websocket.MessageText, msg)
}
