package vault

import (
	"context"
	"errors"
	"log"
	"net/http"

	"pitchlake-backend/server/types"

	"github.com/coder/websocket"
)

func (router *VaultRouter) subscribeVaultHandler(w http.ResponseWriter, r *http.Request) {
	err := router.subscribeVault(r.Context(), w, r)
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

func NewVaultRouter(serveMux *http.ServeMux, logger *log.Logger) *VaultRouter {
	router := &VaultRouter{
		Subscribers: SubscribersWithLock{
			List: make(map[string][]*types.SubscriberVault),
		},
		log: logger,
	}
	serveMux.HandleFunc("/subscribeVault", router.subscribeVaultHandler)
	return router
}
