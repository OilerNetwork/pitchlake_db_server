package vault

import (
	"log"
	"pitchlake-backend/models"
	"sync"

	"github.com/jackc/pgx/v5/pgxpool"
)

type SubscribersWithLock struct {
	list map[string][]*subscriberVault
	mux  sync.Mutex
}
type VaultRouter struct {
	subscriberMessageBuffer int
	subscribers             SubscribersWithLock
	log                     log.Logger
	pool                    pgxpool.Pool
}

type subscriberMessage struct {
	Address      string `json:"address"`
	VaultAddress string `json:"vaultAddress"`
	UserType     string `json:"userType"`
	OptionRound  uint64 `json:"optionRound"`
}

type subscriberVault struct {
	msgs         chan []byte
	address      string
	userType     string
	vaultAddress string
	closeSlow    func()
}

type InitialPayloadVault struct {
	PayloadType            string                        `json:"payloadType"`
	LiquidityProviderState models.LiquidityProviderState `json:"liquidityProviderState"`
	OptionBuyerStates      []*models.OptionBuyer         `json:"optionBuyerStates"`
	VaultState             models.VaultState             `json:"vaultState"`
	OptionRoundStates      []*models.OptionRound         `json:"optionRoundStates"`
}

type subscriberVaultRequest struct {
	UpdatedField string `json:"updatedField"`
	UpdatedValue string `json:"updatedValue"`
}
