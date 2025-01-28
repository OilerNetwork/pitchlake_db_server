package server

import (
	"context"
	"net/http"
	"pitchlake-backend/db"
	"pitchlake-backend/models"
	"sync"
)

type FossilStatus string

const (
	FossilStatusInitial   FossilStatus = "Initial"
	FossilStatusPending   FossilStatus = "Pending"
	FossilStatusCompleted FossilStatus = "Completed"
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

	subscribersVaultMu  sync.Mutex
	subscribersVault    map[string][]*subscriberVault
	subscribersHomeMu   sync.Mutex
	subscribersHome     map[*subscriberHome]struct{}
	subscribersFossilMu sync.Mutex
	subscribersFossil   map[string]map[uint64]map[*subscriberFossil]struct{}
	vaults              map[string]map[uint64]*FossilJob
	ctx                 context.Context
	cancel              context.CancelFunc
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
type subscriberFossil struct {
	vaultAddress string
	targetTime   uint64
	msgs         chan []byte
	closeSlow    func()
}
type FossilJob struct {
	Duration  uint64
	JobStatus FossilStatus
}

type FossilStatusPayload struct {
	FossilStatus FossilStatus `json:"fossilStatus"`
}

type subscriberFossilMessage struct {
	VaultAddress  string `json:"vaultAddress"`
	Duration      uint64 `json:"duration"`
	TargetTime    uint64 `json:"targetTime"`
	ClientAddress string `json:"clientAddress"`
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

type FossilPayload struct {
	Status FossilStatus `json:"status"`
}

type AllowedPayload interface {
	IsAllowedPayload() // Dummy method
}
type NotificationPayload[T AllowedPayload] struct {
	Operation string `json:"operation"`
	Type      string `json:"type"`
	Payload   T      `json:"payload"`
}
type InitialPayload struct {
	PayloadType            string                        `json:"payloadType"`
	LiquidityProviderState models.LiquidityProviderState `json:"liquidityProviderState"`
	OptionBuyerStates      []*models.OptionBuyer         `json:"optionBuyerStates"`
	VaultState             models.VaultState             `json:"vaultState"`
	OptionRoundStates      []*models.OptionRound         `json:"optionRoundStates"`
}
