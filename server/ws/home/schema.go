package home

import (
	"log"
	"sync"

	"github.com/jackc/pgx/v5/pgxpool"
)

type SubscribersWithLock struct {
	list map[*subscriberHome]struct{}
	mux  sync.Mutex
}
type HomeRouter struct {
	subscriberMessageBuffer int
	subscribers             SubscribersWithLock
	log                     log.Logger
	pool                    pgxpool.Pool
}

type subscriberHome struct {
	msgs      chan []byte
	closeSlow func()
}
