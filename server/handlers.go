package server

import (
	"context"
	"errors"
	"net/http"

	"github.com/coder/websocket"
)

func (dbs *dbServer) subscribeHomeHandler(w http.ResponseWriter, r *http.Request) {
	println("Subscribing to home")
	err := dbs.subscribeHome(r.Context(), w, r)
	if errors.Is(err, context.Canceled) {
		return
	}
	if websocket.CloseStatus(err) == websocket.StatusNormalClosure ||
		websocket.CloseStatus(err) == websocket.StatusGoingAway {
		return
	}
	if err != nil {
		dbs.logf("%v", err)
		return
	}
}

// subscribeHandler accepts the WebSocket connection and then subscribes
// it to all future messages.
func (dbs *dbServer) subscribeVaultHandler(w http.ResponseWriter, r *http.Request) {
	err := dbs.subscribeVault(r.Context(), w, r)
	if errors.Is(err, context.Canceled) {
		return
	}
	if websocket.CloseStatus(err) == websocket.StatusNormalClosure ||
		websocket.CloseStatus(err) == websocket.StatusGoingAway {
		return
	}
	if err != nil {
		dbs.logf("%v", err)
		return
	}
}

func (dbs *dbServer) healthCheckHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}

func (dbs *dbServer) subscribeFossilHandler(w http.ResponseWriter, r *http.Request) {
	err := dbs.subscribeFossil(r.Context(), w, r)
	if errors.Is(err, context.Canceled) {
		return
	}
	if websocket.CloseStatus(err) == websocket.StatusNormalClosure ||
		websocket.CloseStatus(err) == websocket.StatusGoingAway {
		return
	}
	if err != nil {
		dbs.logf("%v", err)
		return
	}
}
