package vault

import (
	"log"
	"pitchlake-backend/models"
	"pitchlake-backend/server/types"
	"sync"

	"github.com/jackc/pgx/v5/pgxpool"
)

type SubscribersWithLock struct {
	List map[string][]*types.SubscriberVault
	mux  sync.Mutex
}
type VaultRouter struct {
	subscriberMessageBuffer int
	Subscribers             SubscribersWithLock
	log                     *log.Logger
	pool                    pgxpool.Pool
}

type InitialPayloadVault struct {
	PayloadType            string                        `json:"payloadType"`
	LiquidityProviderState models.LiquidityProviderState `json:"liquidityProviderState"`
	OptionBuyerStates      []*models.OptionBuyer         `json:"optionBuyerStates"`
	VaultState             models.VaultState             `json:"vaultState"`
	OptionRoundStates      []*models.OptionRound         `json:"optionRoundStates"`
}
