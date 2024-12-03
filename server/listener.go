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

	fmt.Println("Waiting for notifications...")

	for {
		// Wait for a notification
		notification, err := dbs.db.Conn.WaitForNotification(context.Background())
		if err != nil {
			log.Fatal(err)
		}
		//Process notification here
		switch notification.Channel {

		case "lp_update":
			var updatedRow models.LiquidityProviderState
			err := json.Unmarshal([]byte(notification.Payload), &updatedRow)
			if err != nil {
				log.Printf("Error parsing lp_update payload: %v", err)
			}
			for _, vaults := range dbs.subscribersVault {
				for _, s := range vaults {
					if s.address == updatedRow.Address && s.userType == "lp" {
						s.msgs <- []byte(notification.Payload)
					}
				}
			}
			fmt.Printf("Received an update on lp_row_update, %s", notification.Payload)
		case "vault_update":
			var updatedRow models.VaultState
			err := json.Unmarshal([]byte(notification.Payload), &updatedRow)
			if err != nil {
				log.Printf("Error parsing vault_update payload: %v", err)
			} else {
				for _, s := range dbs.subscribersVault[updatedRow.Address] {
					s.msgs <- []byte(notification.Payload)
				}
			}
			fmt.Println("Received an update on vault_update")
		case "ob_update":
			var updatedRow models.OptionBuyer
			err := json.Unmarshal([]byte(notification.Payload), &updatedRow)
			if err != nil {
				log.Printf("Error parsing ob_update payload: %v", err)
			} else {
				for _, vaults := range dbs.subscribersVault {
					for _, s := range vaults {
						if s.address == updatedRow.Address && s.userType == "ob" {
							s.msgs <- []byte(notification.Payload)
						}
					}
				}
			}
		case "or_update":
			fmt.Println("Received an update on or_update")
			// Parse the JSON payload
			var updatedRow models.OptionRound
			err := json.Unmarshal([]byte(notification.Payload), &updatedRow)
			if err != nil {
				log.Printf("Error parsing or_update payload: %v", err)
			} else {
				// Print the updated row
				fmt.Printf("Updated OptionRound: %+v\n", updatedRow)
				if dbs.subscribersVault[updatedRow.VaultAddress] != nil {

					for _, s := range dbs.subscribersVault[updatedRow.VaultAddress] {
						s.msgs <- []byte(notification.Payload)
					}
				}
			}
		}
	}
}
