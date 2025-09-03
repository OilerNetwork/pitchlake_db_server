package general

import (
	"log"
	"sync"

	"pitchlake-backend/server/types"

	"github.com/jackc/pgx/v5/pgxpool"
)

type SubscribersWithLock struct {
	List map[*types.SubscriberGas]struct{}
	mux  sync.Mutex
}

type GeneralRouter struct {
	subscriberMessageBuffer int
	Subscribers             SubscribersWithLock
	log                     log.Logger
	pool                    pgxpool.Pool
}
type subscriberGas struct {
	StartTimestamp uint64
	EndTimestamp   uint64
	RoundDuration  uint64
	msgs           chan []byte
	closeSlow      func()
}

type subscriberGasMessage struct {
	StartTimestamp uint64 `json:"startTimestamp"`
	EndTimestamp   uint64 `json:"endTimestamp"`
	RoundDuration  uint64 `json:"roundDuration"`
}

type subscriberGasRequest struct {
	StartTimestamp uint64 `json:"startTimestamp"`
	EndTimestamp   uint64 `json:"endTimestamp"`
	RoundDuration  uint64 `json:"roundDuration"`
}
