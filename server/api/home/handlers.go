package home

import (
	"context"
	"errors"
	"log"
	"net/http"

	"pitchlake-backend/server/types"

	"github.com/coder/websocket"
)

func (router *HomeRouter) subscribeHomeHandler(w http.ResponseWriter, r *http.Request) {
	router.log.Println("New subscriber for home page data")
	err := router.SubscribeHome(r.Context(), w, r)
	if errors.Is(err, context.Canceled) {
		return
	}
	if websocket.CloseStatus(err) == websocket.StatusNormalClosure ||
		websocket.CloseStatus(err) == websocket.StatusGoingAway {
		return
	}
	if err != nil {
		router.log.Printf("%v", err)
		return
	}
}

func NewHomeRouter(serveMux *http.ServeMux, logger *log.Logger) *HomeRouter {
	router := &HomeRouter{
		Subscribers: SubscribersWithLock{
			List: make(map[*types.SubscriberHome]struct{}),
		},
		log: logger,
	}
	serveMux.HandleFunc("/subscribeHome", router.subscribeHomeHandler)
	return router
}
