package vault

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"
	"pitchlake-backend/db/repositories"
	"pitchlake-backend/server/api/utils"
	"pitchlake-backend/server/types"
	"pitchlake-backend/server/validations"
	"sync"
	"time"

	"github.com/coder/websocket"
)

func (router *VaultRouter) subscribeVault(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	var mu sync.Mutex
	var c *websocket.Conn
	var closed bool
	//Extract address from the request and add here

	//allowedOrigin := os.Getenv("APP_URL")
	c2, err := websocket.Accept(w, r, &websocket.AcceptOptions{
		InsecureSkipVerify: true,
	})
	if err != nil {
		return err
	}
	defer c2.Close(websocket.StatusInternalError, "Internal server error")

	// Read the first message to get the subscription data
	_, msg, err := c2.Read(ctx)
	if err != nil {
		return err
	}

	var sm types.SubscriberMessage
	err = json.Unmarshal(msg, &sm)
	if err != nil {
		return err
	}

	// Validate subscription message
	if err := validations.ValidateSubscriptionMessage(sm); err != nil {
		log.Printf("Invalid subscription message: %v", err)
		// Send error response to client
		errorResponse := map[string]string{
			"error": "Invalid subscription message",
			"details": err.Error(),
		}
		errorJson, _ := json.Marshal(errorResponse)
		c2.Write(ctx, websocket.MessageText, errorJson)
		return err
	}

	log.Printf("%v", sm)

	s := &types.SubscriberVault{
		Address:      sm.Address,
		VaultAddress: sm.VaultAddress,
		UserType:     sm.UserType,
		Msgs:         make(chan []byte, router.subscriberMessageBuffer),
		CloseSlow: func() {
			mu.Lock()
			defer mu.Unlock()
			closed = true
			if c != nil {
				c.Close(websocket.StatusPolicyViolation, "connection too slow to keep up with messages")
			}
		},
	}
	router.addSubscriberVault(s)
	defer router.deleteSubscriberVault(s)

	mu.Lock()
	if closed {
		mu.Unlock()
		return net.ErrClosed
	}
	c = c2
	mu.Unlock()
	defer c.CloseNow()

	//Send initial payload here
	var payload InitialPayloadVault

	payload.PayloadType = "initial"

	//Create repositories
	vaultRepo := repositories.NewVaultRepository(&router.pool)
	optionRoundRepo := repositories.NewOptionRepository(&router.pool)
	optionBuyerRepo := repositories.NewOptionBuyerRepository(&router.pool)
	lpRepo := repositories.NewLiquidityRepository(&router.pool)

	vaultState, err := vaultRepo.GetVaultStateByID(ctx, s.VaultAddress)

	if err != nil {
		return err
	}
	optionRounds, err := optionRoundRepo.GetOptionRoundsByVaultAddress(ctx, s.VaultAddress)
	if err != nil {
		return err
	}
	payload.OptionRoundStates = optionRounds
	payload.VaultState = *vaultState
	lpState, err := lpRepo.GetLiquidityProviderStateByAddress(ctx, s.Address, s.VaultAddress)
	if err != nil {
		fmt.Printf("Error fetching lp state %v", err)
	} else {
		payload.LiquidityProviderState = *lpState
	}

	obStates, err := optionBuyerRepo.GetOptionBuyerByAddress(ctx, s.Address)
	if err != nil {
		fmt.Printf("Error fetching ob state %v", err)
	}
	payload.OptionBuyerStates = obStates

	// if sm.UserType == "lp" {

	// } else if sm.UserType == "ob" {

	// } else {
	// 	return errors.New("invalid user type")
	// }

	// Marshal the VaultState to a JSON byte array
	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	utils.WriteTimeout(ctx, time.Second*5, c, jsonPayload)
	go func() {
		for {
			var request types.SubscriberVaultRequest
			_, msg, err := c.Read(ctx)
			if err != nil {
				log.Printf("Error reading message: %v", err)
				break
			}
			log.Printf("Received message from client: %s", msg)
			err = json.Unmarshal(msg, &request)
			if err != nil {
				log.Printf("Incorrect message format: %v", err)
				break
			}

			// Validate vault request
			if err := validations.ValidateVaultRequest(request); err != nil {
				log.Printf("Invalid vault request: %v", err)
				// Send error response to client
				errorResponse := map[string]string{
					"error": "Invalid vault request",
					"details": err.Error(),
				}
				errorJson, _ := json.Marshal(errorResponse)
				s.Msgs <- errorJson
				break
			}

			var payload InitialPayloadVault
			if request.UpdatedField == "address" {
				s.Address = request.UpdatedValue

				payload.PayloadType = "account_update"
				lpState, err := lpRepo.GetLiquidityProviderStateByAddress(ctx, s.Address, s.VaultAddress)
				if err != nil {
					fmt.Printf("Error fetching lp state %v", err)
				} else {
					payload.LiquidityProviderState = *lpState
				}

				obStates, err := optionBuyerRepo.GetOptionBuyerByAddress(ctx, s.Address)
				if err != nil {
					fmt.Printf("Error fetching ob state %v", err)
				}
				payload.OptionBuyerStates = obStates
			}
			jsonPayload, err := json.Marshal(payload)
			if err != nil {
				log.Printf("Incorrect response generated: %v", err)
			}
			s.Msgs <- []byte(jsonPayload)
			log.Printf("Client Info %v", s)
			// Handle the received message here
		}
	}()
	for {
		select {
		case msg := <-s.Msgs:
			//Push messages received on the subscriber channel to the client
			err := utils.WriteTimeout(ctx, time.Second*5, c, msg)
			if err != nil {
				return err
			}
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

func (router *VaultRouter) addSubscriberVault(s *types.SubscriberVault) {

	router.Subscribers.mux.Lock()
	defer router.Subscribers.mux.Unlock()

	// Initialize the slice if it doesn't exist
	if _, exists := router.Subscribers.List[s.VaultAddress]; !exists {
		router.Subscribers.List[s.VaultAddress] = make([]*types.SubscriberVault, 0)
	}

	router.Subscribers.List[s.VaultAddress] = append(router.Subscribers.List[s.VaultAddress], s)

}

// deleteSubscriber deletes the given subscriber.
func (router *VaultRouter) deleteSubscriberVault(s *types.SubscriberVault) {

	router.Subscribers.mux.Lock()
	defer router.Subscribers.mux.Unlock()

	Subscribers, exists := router.Subscribers.List[s.VaultAddress]
	if !exists {
		return // Nothing to delete
	}

	for i, subscriber := range Subscribers {
		if subscriber == s {
			// Replace the element to be deleted with the last element
			Subscribers[i] = Subscribers[len(Subscribers)-1]
			// Truncate the slice
			router.Subscribers.List[s.VaultAddress] = Subscribers[:len(Subscribers)-1]
			break
		}
	}

	// If the slice is empty after deletion, remove the key from the map
	if len(router.Subscribers.List[s.VaultAddress]) == 0 {
		delete(router.Subscribers.List, s.VaultAddress)
	}
}
