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

	subscribersMu sync.Mutex
	subscribers   map[string]map[*subscriber]struct{}
}

// subscriber represents a subscriber.
// Messages are sent on the msgs channel and if the client
// cannot keep up with the messages, closeSlow is called.
type subscriber struct {
	msgs      chan []byte
	address   string
	userType  string
	closeSlow func()
}

type subscriberMessage struct {
	Address  string `json:"address"`
	UserType string `json:"user_type"`
	View     string `json:"view"`
}

// newdbServer constructs a dbServer with the defaults.
// Create a custom context for the server here and pass it to the db package
func newDBServer() *dbServer {

	db := &db.DB{}
	connString := ""
	db.Init(connString)
	dbs := &dbServer{
		subscriberMessageBuffer: 16,
		logf:                    log.Printf,
		subscribers:             make(map[string]map[*subscriber]struct{}),
		db:                      db,
	}
	dbs.serveMux.Handle("/", http.FileServer(http.Dir(".")))
	dbs.serveMux.HandleFunc("/subscribe", dbs.subscribeHandler)
	dbs.listener()
	return dbs
}

func (dbs *dbServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	dbs.serveMux.ServeHTTP(w, r)
}

// subscribeHandler accepts the WebSocket connection and then subscribes
// it to all future messages.
func (dbs *dbServer) subscribeHandler(w http.ResponseWriter, r *http.Request) {
	err := dbs.subscribe(r.Context(), w, r)
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

// publishHandler reads the request body with a limit of 8192 bytes and then publishes
// the received message.

// subscribe subscribes the given WebSocket to all broadcast messages.
// It creates a subscriber with a buffered msgs chan to give some room to slower
// connections and then registers the subscriber. It then listens for all messages
// and writes them to the WebSocket. If the context is cancelled or
// an error occurs, it returns and deletes the subscription.
//
// It uses CloseRead to keep reading from the connection to process control
// messages and cancel the context if the connection drops.
func (dbs *dbServer) subscribe(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
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
	dbs.addSubscriber(s)
	defer dbs.deleteSubscriber(s)

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
	writeTimeout(ctx, time.Second*5, c, []byte("subscribed"))
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

func (dbs *dbServer) listener() {
	_, err := dbs.db.Conn.Exec(context.Background(), "LISTEN lp_update")
	if err != nil {
		log.Fatal(err)
	}

	_, err = dbs.db.Conn.Exec(context.Background(), "LISTEN vault_update")
	if err != nil {
		log.Fatal(err)
	}

	_, err = dbs.db.Conn.Exec(context.Background(), "LISTEN state_transition")
	if err != nil {
		log.Fatal(err)
	}

	_, err = dbs.db.Conn.Exec(context.Background(), "LISTEN ob_update")
	if err != nil {
		log.Fatal(err)
	}

	_, err = dbs.db.Conn.Exec(context.Background(), "LISTEN or_update")
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Waiting for notifications...")

	for {
		// Wait for a notification
		notification, err := dbs.db.Conn.WaitForNotification(context.Background())
		if err != nil {
			log.Fatal(err)
		}
		//Process notification here
		switch notification.Channel {
		case "lp_update":
			fmt.Println("Received an update on lp_row_update")
		case "vault_update":
			fmt.Println("Received an update on vault_update")
		case "state_transition":
			//Push this to all channels (without address as well)
			fmt.Println("Received an update on state_transition")
		case "ob_update":
			fmt.Println("Received an update on ob_update")
		case "or_update":
			fmt.Println("Received an update on or_update")

		}
		dbs.publishAddress(notification.Channel, []byte(notification.Payload))
		dbs.publishAll([]byte(notification.Payload))
	}
}

// publishUser sends a message to all subscribers of a specific address.
func (dbs *dbServer) publishAddress(address string, msg []byte) {
	dbs.subscribersMu.Lock()
	defer dbs.subscribersMu.Unlock()

	for s := range dbs.subscribers[address] {
		select {
		case s.msgs <- msg:
		default:
			go s.closeSlow()
		}
	}
}

// publishAll sends a message to all subscribers of all addresses.
func (dbs *dbServer) publishAll(msg []byte) {
	dbs.subscribersMu.Lock()
	defer dbs.subscribersMu.Unlock()

	for address := range dbs.subscribers {
		for s := range dbs.subscribers[address] {
			select {
			case s.msgs <- msg:
			default:
				go s.closeSlow()
			}
		}
	}
}

// addSubscriber registers a subscriber.
func (dbs *dbServer) addSubscriber(s *subscriber) {
	dbs.subscribersMu.Lock()
	dbs.subscribers[s.address][s] = struct{}{}
	dbs.subscribersMu.Unlock()
}

// deleteSubscriber deletes the given subscriber.
func (dbs *dbServer) deleteSubscriber(s *subscriber) {
	dbs.subscribersMu.Lock()
	delete(dbs.subscribers, s.address)
	dbs.subscribersMu.Unlock()
}

func writeTimeout(ctx context.Context, timeout time.Duration, c *websocket.Conn, msg []byte) error {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	return c.Write(ctx, websocket.MessageText, msg)
}
