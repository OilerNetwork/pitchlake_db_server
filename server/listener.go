package server

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"pitchlake-backend/models"
)

func (dbs *dbServer) listener() {
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

	fmt.Println("Waiting for notifications...")

	for {
		// Wait for a notification
		notification, err := dbs.db.Conn.WaitForNotification(context.Background())
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("PAYLOAD %v", notification.Payload)
		//Process notification here
		switch notification.Channel {

		case "bids_channel":
			var updatedData webSocketPayload
			var bidData BidData
			err := json.Unmarshal([]byte(notification.Payload), &bidData)
			if err != nil {
				log.Printf("Error parsing ob_update payload: %v", err)
				return
			}
			updatedData.PayloadType = "bid_update"
			updatedData.BidData = bidData
			response, err := json.Marshal(updatedData)

			if err != nil {
				log.Printf("Error parsing ob_update payload: %v", err)
				return
			}
			for _, vaults := range dbs.subscribersVault {
				for _, s := range vaults {
					if s.address == bidData.Bid.BuyerAddress {
						s.msgs <- []byte(response)
					}
				}

			}
		case "lp_update":
			var updatedData webSocketPayload
			var updatedRow models.LiquidityProviderState
			err := json.Unmarshal([]byte(notification.Payload), &updatedRow)
			if err != nil {
				log.Printf("Error parsing lp_update payload: %v", err)
				return
			}
			updatedData.LiquidityProviderState = updatedRow
			updatedData.PayloadType = "lp_update"
			response, err := json.Marshal(updatedData)
			if err != nil {
				log.Printf("Error parsing lp_update payload: %v", err)
				return
			}
			for _, lp := range dbs.subscribersVault[updatedRow.VaultAddress] {
				if lp.address == updatedRow.Address {
					lp.msgs <- []byte(response)
				}
			}
			fmt.Printf("Received an update on lp_row_update, %s", notification.Payload)
		case "vault_update":
			var updatedData webSocketPayload
			var updatedRow models.VaultState
			err := json.Unmarshal([]byte(notification.Payload), &updatedRow)
			updatedData.VaultState = updatedRow
			if err != nil {
				log.Printf("Error parsing vault_update payload: %v", err)
				return
			}
			fmt.Printf("REACHED %v ", updatedRow.Address)
			updatedData.PayloadType = "vault_update"
			response, err := json.Marshal(updatedData)
			if err != nil {
				log.Printf("Marshalling error %v", err)
				return
			}
			for _, s := range dbs.subscribersVault[updatedRow.Address] {
				s.msgs <- []byte(response)
			}
			fmt.Println("Received an update on vault_update")
		case "ob_update":
			var updatedData webSocketPayload
			var newOptionBuyer models.OptionBuyer
			err := json.Unmarshal([]byte(notification.Payload), &newOptionBuyer)
			if err != nil {
				log.Printf("Error parsing ob_update payload: %v", err)
				return
			}
			updatedData.PayloadType = "ob_update"
			updatedData.OptionBuyerStates = append(updatedData.OptionBuyerStates, &newOptionBuyer)
			response, err := json.Marshal(updatedData)

			if err != nil {
				log.Printf("Error parsing ob_update payload: %v", err)
				return
			}
			for _, vaults := range dbs.subscribersVault {
				for _, s := range vaults {
					if s.address == newOptionBuyer.Address && s.userType == "ob" {
						s.msgs <- []byte(response)
					}
				}

			}
		case "or_update":
			fmt.Println("Received an update on or_update")
			// Parse the JSON payload
			var updatedData webSocketPayload
			var updatedRow models.OptionRound
			err := json.Unmarshal([]byte(notification.Payload), &updatedRow)
			if err != nil {
				log.Printf("Error parsing or_update payload: %v", err)
				return
			}
			updatedData.PayloadType = "or_update"
			updatedData.OptionRoundStates = append(updatedData.OptionRoundStates, &updatedRow)
			response, err := json.Marshal(updatedData)
			if err != nil {
				log.Printf("Error parsing or_update payload: %v", err)
				return
			}
			// Print the updated row
			fmt.Printf("Updated OptionRound: %+v\n", updatedRow)
			if dbs.subscribersVault[updatedRow.VaultAddress] != nil {

				for _, s := range dbs.subscribersVault[updatedRow.VaultAddress] {
					s.msgs <- []byte(response)
				}
			}
		}
	}
}
