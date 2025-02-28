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
			var updatedData []models.Block
			err := json.Unmarshal([]byte(notification.Payload), &updatedData)
			if err != nil {
				log.Printf("Error parsing confirmed_insert payload: %v", err)
				return
			}
			response := NotificationPayloadGas{
				Type:   "confirmedBlocks",
				Blocks: updatedData,
			}
			jsonResponse, err := json.Marshal(response)
			if err != nil {
				log.Printf("Error parsing confirmed_insert payload: %v", err)
				return
			}
			for sub := range dbs.subscribersGas {
				sub.msgs <- []byte(jsonResponse)
			}
		case "unconfirmed_insert":
			var updatedData models.Block
			err := json.Unmarshal([]byte(notification.Payload), &updatedData)
			if err != nil {
				log.Printf("Error parsing unconfirmed_insert payload: %v", err)
				return
			}
			response := NotificationPayloadGas{
				Type:   "unconfirmedBlocks",
				Blocks: []models.Block{updatedData},
			}
			jsonResponse, err := json.Marshal(response)
			if err != nil {
				log.Printf("Error parsing unconfirmed_insert payload: %v", err)
				return
			}
			for sub := range dbs.subscribersGas {
				sub.msgs <- []byte(jsonResponse)
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
			for _, vaults := range dbs.subscribersVault {
				for _, s := range vaults {
					if s.address == updatedData.Payload.BuyerAddress {
						s.msgs <- []byte(response)
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
			for _, lp := range dbs.subscribersVault[updatedData.Payload.VaultAddress] {
				if lp.address == updatedData.Payload.Address {
					lp.msgs <- []byte(response)
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
			for _, s := range dbs.subscribersVault[updatedData.Payload.Address] {
				s.msgs <- []byte(response)
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
			if dbs.subscribersVault[updatedData.Payload.VaultAddress] != nil {

				for _, s := range dbs.subscribersVault[updatedData.Payload.VaultAddress] {
					s.msgs <- []byte(response)
				}
			}
		}
	}
}
