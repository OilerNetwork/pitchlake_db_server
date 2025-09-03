package server

import (
	"context"
	"log"
	"net/http"
	"pitchlake-backend/db"
	"pitchlake-backend/server/api/general"
	"pitchlake-backend/server/api/home"
	"pitchlake-backend/server/api/vault"
)

// dbServer enables broadcasting to a set of subscribers.

type dbServer struct {
	subscriberMessageBuffer int
	db                      *db.DB
	log                     log.Logger
	serveMux                http.ServeMux
	ctx                     context.Context
	cancel                  context.CancelFunc
}

// newdbServer constructs a dbServer with the defaults.
// Create a custom context for the server here and pass it to the db package
func NewDBServer(ctx context.Context) *dbServer {

	ctx, cancel := context.WithCancel(ctx)
	db, err := db.NewDB()
	if err != nil {
		log.Fatal("Failed to load db")
	}
	dbs := &dbServer{
		log:    *log.Default(),
		db:     db,
		ctx:    ctx,
		cancel: cancel,
	}
	homeRouter := home.NewHomeRouter(&dbs.serveMux, &dbs.log)
	vaultRouter := vault.NewVaultRouter(&dbs.serveMux, &dbs.log)
	generalRouter := general.NewGeneralRouter(&dbs.serveMux, &dbs.log)
	go dbs.listener(ctx, vaultRouter.Subscribers.List, homeRouter.Subscribers.List, generalRouter.Subscribers.List)
	return dbs
}

func (dbs *dbServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	dbs.serveMux.ServeHTTP(w, r)
}
