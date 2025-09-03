package vault

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"
	"pitchlake-backend/betterdb/repositories"
	"pitchlake-backend/server/ws/utils"
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

	var sm subscriberMessage
	err = json.Unmarshal(msg, &sm)
	if err != nil {
		return err
	}

	// Validate subscription message
	// if err := validateSubscriptionMessage(sm); err != nil {
	// 	log.Printf("Invalid subscription message: %v", err)
	// 	return err
	// }

	log.Printf("%v", sm)

	s := &subscriberVault{
		address:      sm.Address,
		vaultAddress: sm.VaultAddress,
		userType:     sm.UserType,
		msgs:         make(chan []byte, router.subscriberMessageBuffer),
		closeSlow: func() {
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

	vaultState, err := vaultRepo.GetVaultStateByID(ctx, s.vaultAddress)

	if err != nil {
		return err
	}
	optionRounds, err := optionRoundRepo.GetOptionRoundsByVaultAddress(ctx, s.vaultAddress)
	if err != nil {
		return err
	}
	payload.OptionRoundStates = optionRounds
	payload.VaultState = *vaultState
	lpState, err := lpRepo.GetLiquidityProviderStateByAddress(ctx, s.address, s.vaultAddress)
	if err != nil {
		fmt.Printf("Error fetching lp state %v", err)
	} else {
		payload.LiquidityProviderState = *lpState
	}

	obStates, err := optionBuyerRepo.GetOptionBuyerByAddress(ctx, s.address)
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
			var request subscriberVaultRequest
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
			// if err := validateVaultRequest(request); err != nil {
			// 	log.Printf("Invalid vault request: %v", err)
			// 	break
			// }

			var payload InitialPayloadVault
			if request.UpdatedField == "address" {
				s.address = request.UpdatedValue

				payload.PayloadType = "account_update"
				lpState, err := lpRepo.GetLiquidityProviderStateByAddress(ctx, s.address, s.vaultAddress)
				if err != nil {
					fmt.Printf("Error fetching lp state %v", err)
				} else {
					payload.LiquidityProviderState = *lpState
				}

				obStates, err := optionBuyerRepo.GetOptionBuyerByAddress(ctx, s.address)
				if err != nil {
					fmt.Printf("Error fetching ob state %v", err)
				}
				payload.OptionBuyerStates = obStates
			}
			jsonPayload, err := json.Marshal(payload)
			if err != nil {
				log.Printf("Incorrect response generated: %v", err)
			}
			s.msgs <- []byte(jsonPayload)
			log.Printf("Client Info %v", s)
			// Handle the received message here
		}
	}()
	for {
		select {
		case msg := <-s.msgs:
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

func (router *VaultRouter) addSubscriberVault(s *subscriberVault) {

	router.subscribers.mux.Lock()
	defer router.subscribers.mux.Unlock()

	// Initialize the slice if it doesn't exist
	if _, exists := router.subscribers.list[s.vaultAddress]; !exists {
		router.subscribers.list[s.vaultAddress] = make([]*subscriberVault, 0)
	}

	router.subscribers.list[s.vaultAddress] = append(router.subscribers.list[s.vaultAddress], s)

}

// deleteSubscriber deletes the given subscriber.
func (router *VaultRouter) deleteSubscriberVault(s *subscriberVault) {

	router.subscribers.mux.Lock()
	defer router.subscribers.mux.Unlock()

	subscribers, exists := router.subscribers.list[s.vaultAddress]
	if !exists {
		return // Nothing to delete
	}

	for i, subscriber := range subscribers {
		if subscriber == s {
			// Replace the element to be deleted with the last element
			subscribers[i] = subscribers[len(subscribers)-1]
			// Truncate the slice
			router.subscribers.list[s.vaultAddress] = subscribers[:len(subscribers)-1]
			break
		}
	}

	// If the slice is empty after deletion, remove the key from the map
	if len(router.subscribers.list[s.vaultAddress]) == 0 {
		delete(router.subscribers.list, s.vaultAddress)
	}
}
