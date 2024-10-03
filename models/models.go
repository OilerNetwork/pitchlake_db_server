package models

import (
	"fmt"
	"math/big"
	"strings"
)

type BigInt struct {
	*big.Int
}

func (b *BigInt) UnmarshalJSON(data []byte) error {
	if b.Int == nil {
		b.Int = new(big.Int)
	}

	s := string(data)
	s = strings.Trim(s, "\"") // Remove quotes if present

	// Check if the string is hexadecimal
	if strings.HasPrefix(s, "0x") || strings.HasPrefix(s, "0X") {
		_, ok := b.Int.SetString(s[2:], 16)
		if !ok {
			return fmt.Errorf("failed to parse %s as hex big.Int", s)
		}
	} else {
		// Try to parse as decimal
		_, ok := b.Int.SetString(s, 10)
		if !ok {
			return fmt.Errorf("failed to parse %s as decimal big.Int", s)
		}
	}

	// Ensure the number is not negative and fits within uint256
	if b.Int.Sign() < 0 {
		return fmt.Errorf("negative numbers are not allowed")
	}
	if b.Int.BitLen() > 256 {
		return fmt.Errorf("number exceeds uint256 range")
	}

	return nil
}

// MarshalJSON implements the json.Marshaler interface
func (b BigInt) MarshalJSON() ([]byte, error) {
	if b.Int == nil {
		return []byte("null"), nil
	}
	return []byte(fmt.Sprintf("\"%s\"", b.Int.String())), nil
}

type Vault struct {
	BlockNumber     string `json:"block_number"`
	UnlockedBalance string `json:"unlocked_balance"`
	LockedBalance   string `json:"locked_balance"`
	StashedBalance  string `json:"stashed_balance"`
}

type LiquidityProvider struct {
	Address         string `json:"address"`
	UnlockedBalance string `json:"unlocked_balance"`
	LockedBalance   string `json:"locked_balance"`
	StashedBalance  string `json:"stashed_balance"`
	BlockNumber     string `json:"block_number"`
}

type OptionBuyer struct {
	Address            string `json:"address"`
	RoundID            uint64 `json:"round_id"`
	TokenizableOptions string `json:"tokenizable_options"`
	RefundableBalance  string `json:"refundable_balance"`
}

type OptionRound struct {
	Address           string `json:"address"`
	RoundID           uint64 `json:"round_id"`
	CapLevel          string `json:"cap_level"`
	StartDate         string `json:"start_date"`
	EndDate           string `json:"end_date"`
	SettlementDate    string `json:"settlement_date"`
	StartingLiquidity BigInt `json:"starting_liquidity"`
	AvailableOptions  BigInt `json:"available_options"`
	ClearingPrice     BigInt `json:"clearing_price"`
	SettlementPrice   BigInt `json:"settlement_price"`
	StrikePrice       BigInt `json:"strike_price"`
	SoldOptions       BigInt `json:"sold_options"`
	State             string `json:"state"`
	Premiums          BigInt `json:"premiums"`
	QueuedLiquidity   BigInt `json:"queued_liquidity"`
	PayoutPerOption   BigInt `json:"payout_per_option"`
	VaultAddress      string `json:"vault_address"`
}

type VaultState struct {
	CurrentRound        string `json:"current_round"`
	CurrentRoundAddress string `json:"current_round_address"`
	UnlockedBalance     string `json:"unlocked_balance"`
	LockedBalance       string `json:"locked_balance"`
	StashedBalance      string `json:"stashed_balance"`
	Address             string `json:"address"`
	LastBlock           string `json:"last_block"`
}

type LiquidityProviderState struct {
	Address         string `json:"address"`
	UnlockedBalance string `json:"unlocked_balance"`
	LockedBalance   string `json:"locked_balance"`
	StashedBalance  string `json:"stashed_balance"`
	QueuedBalance   string `json:"queued_balance"`
	LastBlock       string `json:"last_block"`
}

type QueuedLiquidity struct {
	Address        string `json:"address"`
	RoundID        uint64 `json:"round_id"`
	StartingAmount string `json:"starting_amount"`
	QueuedAmount   string `json:"amount"`
}

type Bid struct {
	Address   string `json:"address"`
	RoundID   uint64 `json:"round_id"`
	BidID     string `json:"bid_id"`
	TreeNonce string `json:"tree_nonce"`
	Amount    string `json:"amount"`
	Price     string `json:"price"`
}
type Position struct {
}

type VaultSubscription struct {
	LiquidityProviderState LiquidityProviderState `json:"liquidity_provider_state"`
	OptionBuyerState       OptionBuyer            `json:"option_buyer_state"`
	VaultState             VaultState             `json:"vault_state"`
	OptionRoundState       OptionRound            `json:"option_round_state"`
}
