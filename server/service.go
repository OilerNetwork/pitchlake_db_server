package server

import (
	"pitchlake-backend/models"
)

// dbServer enables broadcasting to a set of subscribers.

type NotificationPayloadGas struct {
	Type   string          `json:"type"`
	Blocks []BlockResponse `json:"blocks"`
}

type NotificationPayloadVault[T AllowedPayload] struct {
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

func (dbs *dbServer) addSubscriberVault(s *subscriberVault) {

	dbs.subscribersVaultMu.Lock()
	defer dbs.subscribersVaultMu.Unlock()

	// Initialize the slice if it doesn't exist
	if _, exists := dbs.subscribersVault[s.vaultAddress]; !exists {
		dbs.subscribersVault[s.vaultAddress] = make([]*subscriberVault, 0)
	}

	dbs.subscribersVault[s.vaultAddress] = append(dbs.subscribersVault[s.vaultAddress], s)

}
func (dbs *dbServer) addSubscriberHome(s *subscriberHome) {

	dbs.subscribersHomeMu.Lock()
	dbs.subscribersHome[s] = struct{}{}
	dbs.subscribersHomeMu.Unlock()
}

func (dbs *dbServer) addSubscriberGas(s *subscriberGas) {
	dbs.subscribersGasMu.Lock()
	dbs.subscribersGas[s] = struct{}{}
	dbs.subscribersGasMu.Unlock()
}

func (dbs *dbServer) deleteSubscriberHome(s *subscriberHome) {

	dbs.subscribersHomeMu.Lock()
	delete(dbs.subscribersHome, s)
	dbs.subscribersHomeMu.Unlock()
}

func (dbs *dbServer) deleteSubscriberGas(s *subscriberGas) {
	dbs.subscribersGasMu.Lock()
	delete(dbs.subscribersGas, s)
	dbs.subscribersGasMu.Unlock()
}

// deleteSubscriber deletes the given subscriber.
func (dbs *dbServer) deleteSubscriberVault(s *subscriberVault) {

	dbs.subscribersVaultMu.Lock()
	defer dbs.subscribersVaultMu.Unlock()

	subscribers, exists := dbs.subscribersVault[s.vaultAddress]
	if !exists {
		return // Nothing to delete
	}

	for i, subscriber := range subscribers {
		if subscriber == s {
			// Replace the element to be deleted with the last element
			subscribers[i] = subscribers[len(subscribers)-1]
			// Truncate the slice
			dbs.subscribersVault[s.vaultAddress] = subscribers[:len(subscribers)-1]
			break
		}
	}

	// If the slice is empty after deletion, remove the key from the map
	if len(dbs.subscribersVault[s.vaultAddress]) == 0 {
		delete(dbs.subscribersVault, s.vaultAddress)
	}
}
