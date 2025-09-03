package types

import (
	"pitchlake-backend/models"
)

// subscriber represents a subscriber.
// Messages are sent on the msgs channel and if the client
// cannot keep up with the messages, closeSlow is called.
type SubscriberVault struct {
	Msgs         chan []byte
	Address      string
	UserType     string
	VaultAddress string
	CloseSlow    func()
}
type BlockResponse struct {
	BlockNumber uint64 `json:"blockNumber"`
	Timestamp   uint64 `json:"timestamp"`
	BaseFee     string `json:"baseFee"`
	IsConfirmed bool   `json:"isConfirmed"`
	Twap        string `json:"twap"`
}

type SubscriberHome struct {
	Msgs      chan []byte
	CloseSlow func()
}
type SubscriberGas struct {
	StartTimestamp uint64
	EndTimestamp   uint64
	RoundDuration  uint64
	Msgs           chan []byte
	CloseSlow      func()
}

type SubscriberMessage struct {
	Address      string `json:"address"`
	VaultAddress string `json:"vaultAddress"`
	UserType     string `json:"userType"`
	OptionRound  uint64 `json:"optionRound"`
}

type SubscriberVaultRequest struct {
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

type SubscriberGasMessage struct {
	StartTimestamp uint64 `json:"startTimestamp"`
	EndTimestamp   uint64 `json:"endTimestamp"`
	RoundDuration  uint64 `json:"roundDuration"`
}

type SubscriberGasRequest struct {
	StartTimestamp uint64 `json:"startTimestamp"`
	EndTimestamp   uint64 `json:"endTimestamp"`
	RoundDuration  uint64 `json:"roundDuration"`
}
