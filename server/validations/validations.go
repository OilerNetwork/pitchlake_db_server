package validations

import (
	"fmt"
	"pitchlake-backend/server/types"
	"strings"
)

// ValidateSubscriptionMessage validates the incoming subscription message
func ValidateSubscriptionMessage(sm types.SubscriberMessage) error {
	// Validate required fields
	if sm.Address == "" {
		return fmt.Errorf("address is required")
	}
	if sm.VaultAddress == "" {
		return fmt.Errorf("vault address is required")
	}
	if sm.UserType == "" {
		return fmt.Errorf("user type is required")
	}

	// Validate user type
	if sm.UserType != "lp" && sm.UserType != "ob" {
		return fmt.Errorf("invalid user type: %s, must be 'lp' or 'ob'", sm.UserType)
	}

	// Validate Ethereum address format (basic check - should start with 0x and be 42 chars)
	if !strings.HasPrefix(sm.Address, "0x") || len(sm.Address) != 42 {
		return fmt.Errorf("invalid address format: %s", sm.Address)
	}
	if !strings.HasPrefix(sm.VaultAddress, "0x") || len(sm.VaultAddress) != 42 {
		return fmt.Errorf("invalid vault address format: %s", sm.VaultAddress)
	}

	return nil
}

// ValidateGasRequest validates the gas data request
func ValidateGasRequest(req types.SubscriberGasRequest) error {
	// Validate timestamps
	if req.StartTimestamp == 0 {
		return fmt.Errorf("start timestamp is required")
	}
	if req.EndTimestamp == 0 {
		return fmt.Errorf("end timestamp is required")
	}
	if req.StartTimestamp >= req.EndTimestamp {
		return fmt.Errorf("start timestamp must be before end timestamp")
	}

	// Validate round duration
	validDurations := map[uint64]bool{960: true, 13200: true, 2631600: true}
	if !validDurations[req.RoundDuration] {
		return fmt.Errorf("invalid round duration: %d, must be 960, 13200, or 2631600", req.RoundDuration)
	}

	return nil
}

// ValidateVaultRequest validates vault update requests
func ValidateVaultRequest(req types.SubscriberVaultRequest) error {
	// Validate required fields
	if req.UpdatedField == "" {
		return fmt.Errorf("updated field is required")
	}
	if req.UpdatedValue == "" {
		return fmt.Errorf("updated value is required")
	}

	// Validate field names
	validFields := map[string]bool{"address": true}
	if !validFields[req.UpdatedField] {
		return fmt.Errorf("invalid field: %s, must be 'address'", req.UpdatedField)
	}

	// Validate address format if field is address
	if req.UpdatedField == "address" {
		if !strings.HasPrefix(req.UpdatedValue, "0x") || len(req.UpdatedValue) != 42 {
			return fmt.Errorf("invalid address format: %s", req.UpdatedValue)
		}
	}

	return nil
}
