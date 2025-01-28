package server

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"
	"sync"
	"time"

	"pitchlake-backend/fossil"

	"github.com/coder/websocket"
)

func (dbs *dbServer) subscribeFossil(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	var mu sync.Mutex
	var c *websocket.Conn
	var closed bool

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

	var sm subscriberFossilMessage
	err = json.Unmarshal(msg, &sm)
	if err != nil {
		return err
	}
	s := &subscriberFossil{
		vaultAddress: sm.VaultAddress,
		targetTime:   sm.TargetTime,
		msgs:         make(chan []byte, dbs.subscriberMessageBuffer),
		closeSlow: func() {
			// mu.Lock()
			// defer mu.Unlock()
			// closed = true
		},
	}
	dbs.addSubscriberFossil(s)
	defer dbs.deleteSubscriberFossil(s)

	mu.Lock()
	if closed {
		mu.Unlock()
		return net.ErrClosed
	}
	c = c2
	mu.Unlock()
	defer c.CloseNow()

	payload := FossilPayload{}
	//Check if the vault has a fossil status, if not, send a fossil request
	stats, exists := dbs.vaults[sm.VaultAddress][sm.TargetTime]
	if !exists {

		fossilStatus, err := fossil.GetFossilStatus(sm.TargetTime, sm.Duration)
		//Send Fossil Request
		if err != nil {
			//Check if the error is request not found, create a new fossil job
			if fossilStatus.Status == "request not found" {
				//Make a new fossil request, need to make the request so the timestamp is verified before adding to the map
				fossilData, err := fossil.MakeFossilRequest(sm.TargetTime, sm.Duration, sm.VaultAddress, sm.VaultAddress)
				if err != nil {
					return err
				}
				status := FossilStatus(fossilData.Status)
				dbs.vaults[sm.VaultAddress][sm.TargetTime] = &FossilJob{
					Duration:  sm.Duration,
					JobStatus: status,
				}
				//Start a goroutine to monitor the fossil job status
				go dbs.monitorFossilJobStatus(sm.VaultAddress, sm.TargetTime, sm.Duration)
				payload.Status = status
			}
			return err
		}
		dbs.vaults[sm.VaultAddress][sm.TargetTime] = &FossilJob{
			Duration:  sm.Duration,
			JobStatus: FossilStatus(fossilStatus.Status),
		}
		//Start a goroutine to monitor the fossil job status if the status is not completed
		if FossilStatus(fossilStatus.Status) != FossilStatusCompleted {
			go dbs.monitorFossilJobStatus(sm.VaultAddress, sm.TargetTime, sm.Duration)
		}
		payload.Status = FossilStatus(fossilStatus.Status)

	} else {
		if stats.Duration != sm.Duration {
			return fmt.Errorf("round duration mismatch")
		}
		payload.Status = stats.JobStatus

	}

	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	dbs.writeTimeout(ctx, time.Second*5, c, jsonPayload)
	for {
		select {
		case msg := <-s.msgs:
			//Loop to listen for messages from the client
			err := dbs.writeTimeout(ctx, time.Second*5, c, msg)
			if err != nil {
				return err
			}
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

func (dbs *dbServer) subscribeVault(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
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
	log.Printf("%v", sm)

	s := &subscriberVault{
		address:      sm.Address,
		vaultAddress: sm.VaultAddress,
		userType:     sm.UserType,
		msgs:         make(chan []byte, dbs.subscriberMessageBuffer),
		closeSlow: func() {
			// mu.Lock()
			// defer mu.Unlock()
			// closed = true
			// if c != nil {
			// 	c.Close(websocket.StatusPolicyViolation, "connection too slow to keep up with messages")
			// }
		},
	}
	dbs.addSubscriberVault(s)
	defer dbs.deleteSubscriberVault(s)

	mu.Lock()
	if closed {
		mu.Unlock()
		return net.ErrClosed
	}
	c = c2
	mu.Unlock()
	defer c.CloseNow()

	//Send initial payload here
	var payload InitialPayload

	payload.PayloadType = "initial"
	vaultState, err := dbs.db.GetVaultStateByID(s.vaultAddress)
	if err != nil {
		return err
	}
	optionRounds, err := dbs.db.GetOptionRoundsByVaultAddress(s.vaultAddress)
	if err != nil {
		return err
	}
	payload.OptionRoundStates = optionRounds
	payload.VaultState = *vaultState
	lpState, err := dbs.db.GetLiquidityProviderStateByAddress(s.address, s.vaultAddress)
	if err != nil {
		fmt.Printf("Error fetching lp state %v", err)
	} else {
		payload.LiquidityProviderState = *lpState
	}

	obStates, err := dbs.db.GetOptionBuyerByAddress(s.address)
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
	dbs.writeTimeout(ctx, time.Second*5, c, jsonPayload)
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
			var payload InitialPayload
			if request.UpdatedField == "address" {
				s.address = request.UpdatedValue

				payload.PayloadType = "account_update"
				lpState, err := dbs.db.GetLiquidityProviderStateByAddress(s.address, s.vaultAddress)
				if err != nil {
					fmt.Printf("Error fetching lp state %v", err)
				} else {
					payload.LiquidityProviderState = *lpState
				}

				obStates, err := dbs.db.GetOptionBuyerByAddress(s.address)
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
			err := dbs.writeTimeout(ctx, time.Second*5, c, msg)
			if err != nil {
				return err
			}
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

func (dbs *dbServer) subscribeHome(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	var mu sync.Mutex
	var c *websocket.Conn
	var closed bool

	//allowedOrigin := os.Getenv("APP_URL")
	// Accept the WebSocket connection
	c2, err := websocket.Accept(w, r, &websocket.AcceptOptions{
		InsecureSkipVerify: true,
	})
	if err != nil {
		return err
	}
	defer c2.Close(websocket.StatusInternalError, "Internal server error")

	// Read the first message to get the subscription data

	s := &subscriberHome{
		msgs: make(chan []byte, dbs.subscriberMessageBuffer),
		closeSlow: func() {
			// mu.Lock()
			// defer mu.Unlock()
			// closed = true
			// if c != nil {
			// 	c.Close(websocket.StatusPolicyViolation, "connection too slow to keep up with messages")
			// }
		},
	}

	// Add the subscriber to the appropriate map based on the address
	dbs.addSubscriberHome(s)

	defer dbs.deleteSubscriberHome(s)

	log.Printf("Subscribed to home")
	mu.Lock()
	if closed {
		mu.Unlock()
		return net.ErrClosed
	}
	c = c2
	mu.Unlock()
	defer c.CloseNow()

	vaultAddresses, err := dbs.db.GetVaultAddresses()
	if err != nil {
		return err
	}
	log.Printf("vaultAddresses %v", vaultAddresses)
	// Send initial payload here
	response := struct {
		VaultAddresses []string `json:"vaultAddresses"`
	}{
		VaultAddresses: vaultAddresses,
	}
	jsonPayload, err := json.Marshal(response)
	if err != nil {
		return err
	}

	dbs.writeTimeout(ctx, time.Second*5, c, jsonPayload)

	for {
		select {
		case msg := <-s.msgs:
			//Loop to listen for messages from the client
			err := dbs.writeTimeout(ctx, time.Second*5, c, msg)
			if err != nil {
				return err
			}
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

func (dbs *dbServer) writeTimeout(ctx context.Context, timeout time.Duration, c *websocket.Conn, msg []byte) error {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	return c.Write(ctx, websocket.MessageText, msg)
}
