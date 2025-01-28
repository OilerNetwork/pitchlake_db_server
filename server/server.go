package server

import (
	"context"
	"log"
	"net/http"
	"pitchlake-backend/db"
)

// FossilStatus represents the status of a fossil

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
		subscribersFossil:       make(map[string]map[*subscriberFossil]struct{}),
		vaults:                  make(map[string][]*vaultStats),
		db:                      db,
		ctx:                     ctx,
		cancel:                  cancel,
	}
	dbs.serveMux.Handle("/", http.FileServer(http.Dir(".")))
	dbs.serveMux.HandleFunc("/subscribeHome", dbs.subscribeHomeHandler)
	dbs.serveMux.HandleFunc("/subscribeVault", dbs.subscribeVaultHandler)
	dbs.serveMux.HandleFunc("/health", dbs.healthCheckHandler)
	dbs.serveMux.HandleFunc("/subscribeFossil", dbs.subscribeFossilHandler)
	go dbs.dbListener()
	go dbs.fossilListener()
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

func (dbs *dbServer) addSubscriberFossil(s *subscriberFossil) {
	dbs.subscribersFossilMu.Lock()
	defer dbs.subscribersFossilMu.Unlock()

	// Initialize the slice if it doesn't exist
	if _, exists := dbs.subscribersFossil[s.vaultAddress]; !exists {
		dbs.subscribersFossil[s.vaultAddress] = make(map[*subscriberFossil]struct{})
	}

	dbs.subscribersFossil[s.vaultAddress][s] = struct{}{}

}

func (dbs *dbServer) deleteSubscriberFossil(s *subscriberFossil) {

	dbs.subscribersFossilMu.Lock()
	delete(dbs.subscribersFossil[s.vaultAddress], s)
	dbs.subscribersFossilMu.Unlock()
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
