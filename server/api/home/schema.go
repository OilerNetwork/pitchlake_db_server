package home

import (
	"log"
	"pitchlake-backend/server/types"
	"sync"

	"github.com/jackc/pgx/v5/pgxpool"
)

type SubscribersWithLock struct {
	List map[*types.SubscriberHome]struct{}
	mux  sync.Mutex
}
type HomeRouter struct {
	subscriberMessageBuffer int
	Subscribers             SubscribersWithLock
	log                     *log.Logger
	pool                    pgxpool.Pool
}

type subscriberHome struct {
	msgs      chan []byte
	closeSlow func()
}
