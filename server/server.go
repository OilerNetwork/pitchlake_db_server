package server

import (
	"context"
	"log"
	"net/http"
	"pitchlake-backend/db"
	"pitchlake-backend/models"
	"sync"
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
	subscribersVault   map[string][]*subscriberVault
	subscribersHomeMu  sync.Mutex
	subscribersHome    map[*subscriberHome]struct{}
	ctx                context.Context
	cancel             context.CancelFunc
}

// subscriber represents a subscriber.
// Messages are sent on the msgs channel and if the client
// cannot keep up with the messages, closeSlow is called.
type subscriberVault struct {
	msgs         chan []byte
	address      string
	userType     string
	vaultAddress string
	closeSlow    func()
}

type subscriberHome struct {
	msgs      chan []byte
	closeSlow func()
}

type subscriberMessage struct {
	Address      string `json:"address"`
	VaultAddress string `json:"vaultAddress"`
	UserType     string `json:"userType"`
	OptionRound  uint64 `json:"optionRound"`
}

type webSocketPayload struct {
	PayloadType            string                        `json:"payloadType"`
	LiquidityProviderState models.LiquidityProviderState `json:"liquidityProviderState"`
	OptionBuyerState       models.OptionBuyer            `json:"optionBuyerState"`
	VaultState             models.VaultState             `json:"vaultState"`
	OptionRoundStates      []*models.OptionRound         `json:"optionRoundStates"`
}

// newdbServer constructs a dbServer with the defaults.
// Create a custom context for the server here and pass it to the db package
func NewDBServer(ctx context.Context) *dbServer {

	ctx, cancel := context.WithCancel(ctx)
	db := &db.DB{}
	db.Init()
	dbs := &dbServer{
		subscriberMessageBuffer: 16,
		logf:                    log.Printf,
		subscribersVault:        make(map[string][]*subscriberVault),
		subscribersHome:         make(map[*subscriberHome]struct{}),
		db:                      db,
		ctx:                     ctx,
		cancel:                  cancel,
	}
	dbs.serveMux.Handle("/", http.FileServer(http.Dir(".")))
	dbs.serveMux.HandleFunc("/subscribeHome", dbs.subscribeHomeHandler)
	dbs.serveMux.HandleFunc("/subscribeVault", dbs.subscribeVaultHandler)
	dbs.serveMux.HandleFunc("/health", dbs.healthCheckHandler)
	go dbs.listener()
	return dbs
}

func (dbs *dbServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	dbs.serveMux.ServeHTTP(w, r)
}

// addSubscriber registers a subscriber.
func (dbs *dbServer) addSubscriberVault(s *subscriberVault) {

	dbs.subscribersVaultMu.Lock()
	defer dbs.subscribersVaultMu.Unlock()

	// Initialize the slice if it doesn't exist
	if _, exists := dbs.subscribersVault[s.vaultAddress]; !exists {
		dbs.subscribersVault[s.vaultAddress] = make([]*subscriberVault, 0)
	}

	dbs.subscribersVault[s.vaultAddress] = append(dbs.subscribersVault[s.vaultAddress], s)

}
func (dbs *dbServer) addSubscriberHome(s *subscriberHome) {

	dbs.subscribersHomeMu.Lock()
	dbs.subscribersHome[s] = struct{}{}
	dbs.subscribersHomeMu.Unlock()
}

func (dbs *dbServer) deleteSubscriberHome(s *subscriberHome) {

	dbs.subscribersHomeMu.Lock()
	delete(dbs.subscribersHome, s)
	dbs.subscribersHomeMu.Unlock()
}

// deleteSubscriber deletes the given subscriber.
func (dbs *dbServer) deleteSubscriberVault(s *subscriberVault) {

	dbs.subscribersVaultMu.Lock()
	defer dbs.subscribersVaultMu.Unlock()

	subscribers, exists := dbs.subscribersVault[s.vaultAddress]
	if !exists {
		return // Nothing to delete
	}

	for i, subscriber := range subscribers {
		if subscriber == s {
			// Replace the element to be deleted with the last element
			subscribers[i] = subscribers[len(subscribers)-1]
			// Truncate the slice
			dbs.subscribersVault[s.vaultAddress] = subscribers[:len(subscribers)-1]
			break
		}
	}

	// If the slice is empty after deletion, remove the key from the map
	if len(dbs.subscribersVault[s.vaultAddress]) == 0 {
		delete(dbs.subscribersVault, s.vaultAddress)
	}
}
