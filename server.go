package main

import (
	"context"
	"database/sql"
	"errors"
	"io"
	"log"
	"net"
	"net/http"
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
	db                      *sql.DB

	// publishLimiter controls the rate limit applied to the publish endpoint.
	//
	// Defaults to one publish every 100ms with a burst of 8.

	// logf controls where logs are sent.
	// Defaults to log.Printf.
	logf func(f string, v ...interface{})

	// serveMux routes the various endpoints to the appropriate handler.
	serveMux http.ServeMux

	subscribersMu sync.Mutex
	subscribers   map[string]*subscriber
}

// newdbServer constructs a dbServer with the defaults.
func newDBServer(db *sql.DB) *dbServer {
	dbs := &dbServer{
		subscriberMessageBuffer: 16,
		logf:                    log.Printf,
		subscribers:             make(map[string]*subscriber),
		db:                      db,
	}
	dbs.serveMux.Handle("/", http.FileServer(http.Dir(".")))
	dbs.serveMux.HandleFunc("/subscribe", dbs.subscribeHandler)
	dbs.serveMux.HandleFunc("/publish", dbs.publishHandler)

	return dbs
}

// subscriber represents a subscriber.
// Messages are sent on the msgs channel and if the client
// cannot keep up with the messages, closeSlow is called.
type subscriber struct {
	msgs      chan []byte
	address   string
	closeSlow func()
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
func (dbs *dbServer) publishHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
		return
	}
	body := http.MaxBytesReader(w, r.Body, 8192)
	msg, err := io.ReadAll(body)
	if err != nil {
		http.Error(w, http.StatusText(http.StatusRequestEntityTooLarge), http.StatusRequestEntityTooLarge)
		return
	}

	dbs.publish(msg)

	w.WriteHeader(http.StatusAccepted)
}

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
	s := &subscriber{
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

// publish publishes the msg to all subscribers.
// It never blocks and so messages to slow subscribers
// are dropped.
func (dbs *dbServer) publish(msg []byte) {
	dbs.subscribersMu.Lock()
	defer dbs.subscribersMu.Unlock()

	for _, s := range dbs.subscribers {
		select {
		case s.msgs <- msg:
		default:
			go s.closeSlow()
		}
	}
}

// addSubscriber registers a subscriber.
func (dbs *dbServer) addSubscriber(s *subscriber) {
	dbs.subscribersMu.Lock()
	dbs.subscribers[s.address] = s
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
