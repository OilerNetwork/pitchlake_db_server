package server

import (
	"context"
	"net/http"
	"pitchlake-backend/db"
	"pitchlake-backend/models"
	"sync"
)

type dbServer struct {
	subscriberMessageBuffer int
	db                      *db.DB
	logf                    func(f string, v ...interface{})

	serveMux http.ServeMux

	subscribersVaultMu sync.Mutex
	subscribersVault   map[string][]*subscriberVault
	subscribersHomeMu  sync.Mutex
	subscribersHome    map[*subscriberHome]struct{}
	subscribersGasMu   sync.Mutex
	subscribersGas     map[*subscriberGas]struct{}
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
type subscriberGas struct {
	msgs      chan []byte
	closeSlow func()
}

type subscriberMessage struct {
	Address      string `json:"address"`
	VaultAddress string `json:"vaultAddress"`
	UserType     string `json:"userType"`
	OptionRound  uint64 `json:"optionRound"`
}

type subscriberVaultRequest struct {
	UpdatedField string `json:"updatedField"`
	UpdatedValue string `json:"updatedValue"`
}

type BidData struct {
	Operation string     `json:"operation"`
	Bid       models.Bid `json:"bid"`
}

type AllowedPayload interface {
	IsAllowedPayload() // Dummy method
}

type subscriberGasMessage struct {
	StartBlock uint64 `json:"startBlock"`
	EndBlock   uint64 `json:"endBlock"`
}

type subscriberGasRequest struct {
	StartBlock uint64 `json:"startBlock"`
	EndBlock   uint64 `json:"endBlock"`
}
