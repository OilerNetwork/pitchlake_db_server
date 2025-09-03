package server

import (
	"context"
	"log"
	"net/http"
	"pitchlake-backend/db"
)

// dbServer enables broadcasting to a set of subscribers.

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
		subscribersGas:          make(map[*subscriberGas]struct{}),
		db:                      db,
		ctx:                     ctx,
		cancel:                  cancel,
	}
	InitializeRoutes(dbs)
	go dbs.listener()
	return dbs
}

func (dbs *dbServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	dbs.serveMux.ServeHTTP(w, r)
}

func InitializeRoutes(dbs *dbServer) {

	//Add new routes here
	dbs.serveMux.HandleFunc("/subscribeHome", dbs.subscribeHomeHandler)
	dbs.serveMux.HandleFunc("/subscribeVault", dbs.subscribeVaultHandler)
	dbs.serveMux.HandleFunc("/health", dbs.healthCheckHandler)
	dbs.serveMux.HandleFunc("/subscribeGas", dbs.subscribeGasDataHandler)
}
