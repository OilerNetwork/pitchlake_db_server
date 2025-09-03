package general

import (
	"context"
	"errors"
	"net/http"

	"github.com/coder/websocket"
)

func (router *GeneralRouter) subscribeGasDataHandler(w http.ResponseWriter, r *http.Request) {
	err := router.subscribeGasData(r.Context(), w, r)
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

func (router *GeneralRouter) healthCheckHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}

func NewGeneralRouter(serveMux *http.ServeMux) *GeneralRouter {
	router := &GeneralRouter{}
	serveMux.HandleFunc("/health", router.healthCheckHandler)
	serveMux.HandleFunc("/subscribeGas", router.subscribeGasDataHandler)
	return router
}
