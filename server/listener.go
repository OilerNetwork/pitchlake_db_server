package server

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"pitchlake-backend/db/repositories"
	"pitchlake-backend/models"
	"pitchlake-backend/server/types"
)

type confirmedUpdate struct {
	StartTimestamp uint64 `json:"start_timestamp"`
	EndTimestamp   uint64 `json:"end_timestamp"`
}

type NotificationPayloadGas struct {
	Type   string                `json:"type"`
	Blocks []types.BlockResponse `json:"blocks"`
}

type NotificationPayloadVault[T types.AllowedPayload] struct {
	Operation string `json:"operation"`
	Type      string `json:"type"`
	Payload   T      `json:"payload"`
}
type InitialPayloadVault struct {
	PayloadType            string                        `json:"payloadType"`
	LiquidityProviderState models.LiquidityProviderState `json:"liquidityProviderState"`
	OptionBuyerStates      []*models.OptionBuyer         `json:"optionBuyerStates"`
	VaultState             models.VaultState             `json:"vaultState"`
	OptionRoundStates      []*models.OptionRound         `json:"optionRoundStates"`
}

type InitialPayloadGas struct {
	UnconfirmedBlocks []models.Block `json:"unconfirmedBlocks"`
	ConfirmedBlocks   []models.Block `json:"confirmedBlocks"`
}

func (dbs *dbServer) listener(ctx context.Context, sv map[string][]*types.SubscriberVault, sh map[*types.SubscriberHome]struct{}, sg map[*types.SubscriberGas]struct{}) {
	_, err := dbs.db.Conn.Exec(context.Background(), "LISTEN lp_update")
	if err != nil {
		log.Fatal(err)
	}

	_, err = dbs.db.Conn.Exec(context.Background(), "LISTEN vault_update")
	if err != nil {
		log.Fatal(err)
	}

	_, err = dbs.db.Conn.Exec(context.Background(), "LISTEN ob_update")
	if err != nil {
		log.Fatal(err)
	}

	_, err = dbs.db.Conn.Exec(context.Background(), "LISTEN or_update")
	if err != nil {
		log.Fatal(err)
	}
	_, err = dbs.db.Conn.Exec(context.Background(), "LISTEN bids_update")
	if err != nil {
		log.Fatal(err)
	}

	_, err = dbs.db.Conn.Exec(context.Background(), "LISTEN unconfirmed_insert")
	if err != nil {
		log.Fatal(err)
	}

	_, err = dbs.db.Conn.Exec(context.Background(), "LISTEN confirmed_insert")
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Waiting for notifications...")

	for {
		// Wait for a notification
		notification, err := dbs.db.Conn.WaitForNotification(context.Background())
		if err != nil {
			log.Fatal(err)
		}
		//Process notification here
		switch notification.Channel {

		case "confirmed_insert":
			fmt.Println("Received a confirmed insert")
			var updatedData confirmedUpdate
			err := json.Unmarshal([]byte(notification.Payload), &updatedData)
			if err != nil {
				log.Printf("Error parsing confirmed_insert payload: %v", err)
				return
			}
			blockRepo := repositories.NewBlockRepository(dbs.db.Pool)
			blocks, err := blockRepo.GetBlocks(ctx, updatedData.StartTimestamp, updatedData.EndTimestamp, 0)
			if err != nil {
				log.Printf("Error parsing confirmed_insert payload: %v", err)
				return
			}
			log.Printf("Blocks: %v", blocks)

			var twelveMinResponse, threeHourResponse, thirtyDayResponse []types.BlockResponse

			for _, block := range blocks {
				twelveMinResponse = append(twelveMinResponse, types.BlockResponse{
					BlockNumber: block.BlockNumber,
					Timestamp:   block.Timestamp,
					BaseFee:     block.BaseFee,
					IsConfirmed: block.IsConfirmed,
					Twap:        block.TwelveMinTwap,
				})
				threeHourResponse = append(threeHourResponse, types.BlockResponse{
					BlockNumber: block.BlockNumber,
					Timestamp:   block.Timestamp,
					BaseFee:     block.BaseFee,
					IsConfirmed: block.IsConfirmed,
					Twap:        block.ThreeHourTwap,
				})
				thirtyDayResponse = append(thirtyDayResponse, types.BlockResponse{
					BlockNumber: block.BlockNumber,
					Timestamp:   block.Timestamp,
					BaseFee:     block.BaseFee,
					IsConfirmed: block.IsConfirmed,
					Twap:        block.ThirtyDayTwap,
				})
			}
			responseTwelveMin := NotificationPayloadGas{
				Type:   "confirmedBlocks",
				Blocks: twelveMinResponse,
			}
			responseThreeHour := NotificationPayloadGas{
				Type:   "confirmedBlocks",
				Blocks: threeHourResponse,
			}
			responseThirtyDay := NotificationPayloadGas{
				Type:   "confirmedBlocks",
				Blocks: thirtyDayResponse,
			}
			jsonResponseTwelveMin, err := json.Marshal(responseTwelveMin)
			if err != nil {
				log.Printf("Error parsing confirmed_insert payload: %v", err)
				return
			}
			jsonResponseThreeHour, err := json.Marshal(responseThreeHour)
			if err != nil {
				log.Printf("Error parsing confirmed_insert payload: %v", err)
				return
			}
			jsonResponseThirtyDay, err := json.Marshal(responseThirtyDay)
			if err != nil {
				log.Printf("Error parsing confirmed_insert payload: %v", err)
				return
			}
			for sub := range sg {
				log.Print("Sending payload")
				switch sub.RoundDuration {
				case 960:
					sub.Msgs <- []byte(jsonResponseTwelveMin)
				case 13200:
					sub.Msgs <- []byte(jsonResponseThreeHour)
				case 2631600:
					sub.Msgs <- []byte(jsonResponseThirtyDay)
				}
			}
		case "unconfirmed_insert":
			log.Printf("Received an unconfirmed insert")
			var updatedData models.Block
			err := json.Unmarshal([]byte(notification.Payload), &updatedData)
			if err != nil {
				log.Printf("Error parsing unconfirmed_insert payload: %v", err)
				return
			}
			twelveMinResponse := types.BlockResponse{
				BlockNumber: updatedData.BlockNumber,
				Timestamp:   updatedData.Timestamp,
				BaseFee:     updatedData.BaseFee,
				IsConfirmed: updatedData.IsConfirmed,
				Twap:        updatedData.TwelveMinTwap,
			}
			threeHourResponse := types.BlockResponse{
				BlockNumber: updatedData.BlockNumber,
				Timestamp:   updatedData.Timestamp,
				BaseFee:     updatedData.BaseFee,
				IsConfirmed: updatedData.IsConfirmed,
				Twap:        updatedData.ThreeHourTwap,
			}
			thirtyDayResponse := types.BlockResponse{
				BlockNumber: updatedData.BlockNumber,
				Timestamp:   updatedData.Timestamp,
				BaseFee:     updatedData.BaseFee,
				IsConfirmed: updatedData.IsConfirmed,
				Twap:        updatedData.ThirtyDayTwap,
			}
			responseTwelveMin := NotificationPayloadGas{
				Type:   "unconfirmedBlocks",
				Blocks: []types.BlockResponse{twelveMinResponse},
			}
			responseThreeHour := NotificationPayloadGas{
				Type:   "unconfirmedBlocks",
				Blocks: []types.BlockResponse{threeHourResponse},
			}
			responseThirtyDay := NotificationPayloadGas{
				Type:   "unconfirmedBlocks",
				Blocks: []types.BlockResponse{thirtyDayResponse},
			}
			jsonResponseTwelveMin, err := json.Marshal(responseTwelveMin)
			if err != nil {
				log.Printf("Error parsing unconfirmed_insert payload: %v", err)
				return
			}
			jsonResponseThreeHour, err := json.Marshal(responseThreeHour)
			if err != nil {
				log.Printf("Error parsing unconfirmed_insert payload: %v", err)
				return
			}
			jsonResponseThirtyDay, err := json.Marshal(responseThirtyDay)
			if err != nil {
				log.Printf("Error parsing unconfirmed_insert payload: %v", err)
				return
			}
			for sub := range sg {
				switch sub.RoundDuration {
				case 960:
					sub.Msgs <- []byte(jsonResponseTwelveMin)
				case 13200:
					sub.Msgs <- []byte(jsonResponseThreeHour)
				case 2631600:
					sub.Msgs <- []byte(jsonResponseThirtyDay)
				}
			}
		case "bids_update":
			var updatedData NotificationPayloadVault[models.Bid]
			err := json.Unmarshal([]byte(notification.Payload), &updatedData)
			if err != nil {
				log.Printf("Error parsing ob_update payload: %v", err)
				return
			}
			updatedData.Type = "bid"
			response, err := json.Marshal(updatedData)

			if err != nil {
				log.Printf("Error parsing ob_update payload: %v", err)
				return
			}
			for _, vaults := range sv {
				for _, s := range vaults {
					if s.Address == updatedData.Payload.BuyerAddress {
						s.Msgs <- []byte(response)
					}
				}

			}
		case "lp_update":
			var updatedData NotificationPayloadVault[models.LiquidityProviderState]
			err := json.Unmarshal([]byte(notification.Payload), &updatedData)
			if err != nil {
				log.Printf("Error parsing lp_update payload: %v", err)
				return
			}
			updatedData.Type = "lpState"
			response, err := json.Marshal(updatedData)
			if err != nil {
				log.Printf("Error parsing lp_update payload: %v", err)
				return
			}
			for _, lp := range sv[updatedData.Payload.VaultAddress] {
				if lp.Address == updatedData.Payload.Address {
					lp.Msgs <- []byte(response)
				}
			}
			fmt.Printf("Received an update on lp_row_update, %s", notification.Payload)
		case "vault_update":
			var updatedData NotificationPayloadVault[models.VaultState]
			err := json.Unmarshal([]byte(notification.Payload), &updatedData)
			if err != nil {
				log.Printf("Error parsing vault_update payload: %v", err)
				return
			}
			updatedData.Type = "vaultState"
			response, err := json.Marshal(updatedData)
			if err != nil {
				log.Printf("Marshalling error %v", err)
				return
			}
			for _, s := range sv[updatedData.Payload.Address] {
				s.Msgs <- []byte(response)
			}
			fmt.Println("Received an update on vault_update")
		case "ob_update":
			var updatedData NotificationPayloadVault[models.OptionBuyer]
			var newOptionBuyer models.OptionBuyer
			err := json.Unmarshal([]byte(notification.Payload), &updatedData)
			if err != nil {
				log.Printf("Error parsing ob_update payload: %v", err)
				return
			}
			updatedData.Type = "optionBuyerState"
			response, err := json.Marshal(updatedData)

			if err != nil {
				log.Printf("Error parsing ob_update payload: %v", err)
				return
			}
			for _, vaults := range sv {
				for _, s := range vaults {
					if s.Address == newOptionBuyer.Address && s.UserType == "ob" {
						s.Msgs <- []byte(response)
					}
				}
			}
		case "or_update":
			fmt.Println("Received an update on or_update")
			// Parse the JSON payload
			var updatedData NotificationPayloadVault[models.OptionRound]
			err := json.Unmarshal([]byte(notification.Payload), &updatedData)
			if err != nil {
				log.Printf("Error parsing or_update payload: %v", err)
				return
			}
			updatedData.Type = "optionRoundState"
			response, err := json.Marshal(updatedData)
			if err != nil {
				log.Printf("Error parsing or_update payload: %v", err)
				return
			}
			// Print the updated row
			fmt.Printf("Updated OptionRound: %+v\n", updatedData.Payload.Address)
			if sv[updatedData.Payload.VaultAddress] != nil {

				for _, s := range sv[updatedData.Payload.VaultAddress] {
					s.Msgs <- []byte(response)
				}
			}
		}
	}
}
