package general

import (
	"context"
	"encoding/json"
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

func (router *GeneralRouter) subscribeGasData(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	var mu sync.Mutex
	var c *websocket.Conn
	var closed bool

	log.Printf("Subscribing to gas data")

	c2, err := websocket.Accept(w, r, &websocket.AcceptOptions{
		InsecureSkipVerify: true,
	})
	if err != nil {
		return err
	}
	defer c2.Close(websocket.StatusInternalError, "Internal server error")

	// Create a context that we can cancel
	readerCtx, cancelReader := context.WithCancel(ctx)
	defer cancelReader()

	s := &types.SubscriberGas{
		Msgs:           make(chan []byte, router.subscriberMessageBuffer),
		StartTimestamp: 0,
		EndTimestamp:   0,
		RoundDuration:  0,
		CloseSlow: func() {
			mu.Lock()
			defer mu.Unlock()
			closed = true
			if c != nil {
				c.Close(websocket.StatusPolicyViolation, "connection too slow to keep up with messages")
			}
			cancelReader() // Cancel the reader goroutine
		},
	}

	router.addSubscriberGas(s)
	defer router.deleteSubscriberGas(s)

	mu.Lock()
	if closed {
		mu.Unlock()
		return net.ErrClosed
	}
	c = c2
	mu.Unlock()

	defer c.CloseNow()

	// Create error channel to handle goroutine errors
	errChan := make(chan error, 1)

	go func() {
		defer close(errChan)
		for {
			select {
			case <-readerCtx.Done():
				return
			default:
				var request types.SubscriberGasRequest
				_, msg, err := c.Read(ctx)
				if err != nil {
					log.Printf("Error reading message: %v", err)
					errChan <- err
					return
				}
				log.Printf("Received message from client: %s", msg)
				err = json.Unmarshal(msg, &request)
				if err != nil {
					log.Printf("Incorrect message format: %v", err)
					errChan <- err
					return
				}

				// Validate gas request
				if err := validations.ValidateGasRequest(request); err != nil {
					log.Printf("Invalid gas request: %v", err)
					// Send error response to client
					errorResponse := map[string]string{
						"error": "Invalid request",
						"details": err.Error(),
					}
					errorJson, _ := json.Marshal(errorResponse)
					c.Write(ctx, websocket.MessageText, errorJson)
					errChan <- err
					return
				}

				s.StartTimestamp = request.StartTimestamp
				s.EndTimestamp = request.EndTimestamp
				s.RoundDuration = request.RoundDuration
				blockRepo := repositories.NewBlockRepository(&router.pool)
				blocks, err := blockRepo.GetBlocks(ctx, request.StartTimestamp, request.EndTimestamp, request.RoundDuration)
				if err != nil {
					log.Printf("Error fetching blocks: %v", err)
					errChan <- err
					return
				}
				var confirmedBlocks, unconfirmedBlocks []types.BlockResponse
				for _, block := range blocks {
					var twap string
					switch request.RoundDuration {
					case 960:
						twap = block.TwelveMinTwap
					case 13200:
						twap = block.ThreeHourTwap
					case 2631600:
						twap = block.ThirtyDayTwap
					}
					if block.IsConfirmed {
						confirmedBlocks = append(confirmedBlocks, types.BlockResponse{
							BlockNumber: block.BlockNumber,
							Timestamp:   block.Timestamp,
							BaseFee:     block.BaseFee,
							IsConfirmed: block.IsConfirmed,
							Twap:        twap,
						})
					} else {
						unconfirmedBlocks = append(unconfirmedBlocks, types.BlockResponse{
							BlockNumber: block.BlockNumber,
							Timestamp:   block.Timestamp,
							BaseFee:     block.BaseFee,
							IsConfirmed: block.IsConfirmed,
							Twap:        twap,
						})
					}
				}
				response := struct {
					ConfirmedBlocks   []types.BlockResponse `json:"confirmedBlocks"`
					UnconfirmedBlocks []types.BlockResponse `json:"unconfirmedBlocks"`
				}{
					ConfirmedBlocks:   confirmedBlocks,
					UnconfirmedBlocks: unconfirmedBlocks,
				}
				jsonPayload, err := json.Marshal(response)
				if err != nil {
					log.Printf("Incorrect response generated: %v", err)
					errChan <- err
					return
				}
				select {
				case s.Msgs <- []byte(jsonPayload):
				case <-readerCtx.Done():
					return
				}
			}
		}
	}()

	for {
		select {
		case err := <-errChan:
			return err
		case msg := <-s.Msgs:
			err := utils.WriteTimeout(ctx, time.Second*5, c, msg)
			if err != nil {
				return err
			}
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

func (router *GeneralRouter) addSubscriberGas(s *types.SubscriberGas) {
	router.Subscribers.mux.Lock()
	router.Subscribers.List[s] = struct{}{}
	router.Subscribers.mux.Unlock()
}

func (router *GeneralRouter) deleteSubscriberGas(s *types.SubscriberGas) {
	router.Subscribers.mux.Lock()
	delete(router.Subscribers.List, s)
	router.Subscribers.mux.Unlock()
}
