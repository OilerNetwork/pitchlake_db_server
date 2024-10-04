package models

import (
	"database/sql/driver"
	"fmt"
	"math/big"
	"strings"
)

type BigInt struct {
	*big.Int
}

var (
	maxUint256 = new(big.Int).Sub(new(big.Int).Lsh(big.NewInt(1), 256), big.NewInt(1))
)

// Scan implements the sql.Scanner interface for BigInt
func (b *BigInt) Scan(value interface{}) error {
	if b.Int == nil {
		b.Int = new(big.Int)
	}

	switch v := value.(type) {
	case []byte:
		return b.scanString(string(v))
	case string:
		return b.scanString(v)
	case int64:
		b.Int.SetInt64(v)
	case nil:
		b.Int.SetInt64(0)
	default:
		return fmt.Errorf("unsupported Scan, storing driver.Value type %T into type BigInt", value)
	}

	return b.validateUint256()
}

func (b *BigInt) scanString(s string) error {
	s = strings.TrimSpace(s)
	_, ok := b.Int.SetString(s, 10) // Parse as decimal
	if !ok {
		return fmt.Errorf("failed to scan BigInt: invalid value %q", s)
	}
	return b.validateUint256()
}

func (b *BigInt) validateUint256() error {
	if b.Int.Sign() < 0 {
		return fmt.Errorf("negative numbers are not allowed for uint256")
	}
	if b.Int.Cmp(maxUint256) > 0 {
		return fmt.Errorf("value exceeds maximum uint256")
	}
	return nil
}

// Value implements the driver.Valuer interface for BigInt
func (b BigInt) Value() (driver.Value, error) {
	if b.Int == nil {
		return "0", nil
	}
	return b.Int.String(), nil // Return as decimal string
}

// UnmarshalJSON implements the json.Unmarshaler interface
func (b *BigInt) UnmarshalJSON(data []byte) error {
	if string(data) == "null" {
		return nil // This allows for null values
	}
	var i big.Int
	err := i.UnmarshalJSON(data)
	if err != nil {
		return err
	}
	b.Int = &i
	return nil
}

// MarshalJSON implements the json.Marshaler interface.
func (b *BigInt) MarshalJSON() ([]byte, error) {
	if b == nil || b.Int == nil {
		return []byte("null"), nil
	}
	return b.Int.MarshalJSON()
}

// String returns a decimal string representation of BigInt
func (b BigInt) String() string {
	if b.Int == nil {
		return "0"
	}
	return b.Int.String()
}

type Vault struct {
	BlockNumber     BigInt `json:"block_number"`
	UnlockedBalance BigInt `json:"unlocked_balance"`
	LockedBalance   BigInt `json:"locked_balance"`
	StashedBalance  BigInt `json:"stashed_balance"`
}

type LiquidityProvider struct {
	Address         string `json:"address"`
	UnlockedBalance BigInt `json:"unlocked_balance"`
	LockedBalance   BigInt `json:"locked_balance"`
	StashedBalance  BigInt `json:"stashed_balance"`
}

type OptionBuyer struct {
	Address            string `json:"address"`
	RoundID            BigInt `json:"round_id"`
	TokenizableOptions BigInt `json:"tokenizable_options"`
	RefundableBalance  BigInt `json:"refundable_balance"`
}

type OptionRound struct {
	Address           *string `json:"address"`
	RoundID           *BigInt `json:"round_id"`
	CapLevel          *BigInt `json:"cap_level"`
	StartDate         *string `json:"start_date"`
	EndDate           *string `json:"end_date"`
	SettlementDate    *string `json:"settlement_date"`
	StartingLiquidity *BigInt `json:"starting_liquidity"`
	AvailableOptions  *BigInt `json:"available_options"`
	ClearingPrice     *BigInt `json:"clearing_price"`
	SettlementPrice   *BigInt `json:"settlement_price"`
	StrikePrice       *BigInt `json:"strike_price"`
	SoldOptions       *BigInt `json:"sold_options"`
	State             *string `json:"state"`
	Premiums          *BigInt `json:"premiums"`
	QueuedLiquidity   *BigInt `json:"queued_liquidity"`
	PayoutPerOption   *BigInt `json:"payout_per_option"`
	VaultAddress      *string `json:"vault_address"`
}

type VaultState struct {
	CurrentRound        BigInt `json:"current_round"`
	CurrentRoundAddress string `json:"current_round_address"`
	UnlockedBalance     BigInt `json:"unlocked_balance"`
	LockedBalance       BigInt `json:"locked_balance"`
	StashedBalance      BigInt `json:"stashed_balance"`
	Address             string `json:"address"`
	LatestBlock         BigInt `json:"last_block"`
}

type LiquidityProviderState struct {
	Address         string `json:"address"`
	UnlockedBalance BigInt `json:"unlocked_balance"`
	LockedBalance   BigInt `json:"locked_balance"`
	StashedBalance  BigInt `json:"stashed_balance"`
	QueuedBalance   BigInt `json:"queued_balance"`
	LastBlock       BigInt `json:"last_block"`
}

type Bid struct {
	Address   string `json:"address"`
	RoundID   BigInt `json:"round_id"`
	BidID     string `json:"bid_id"`
	TreeNonce string `json:"tree_nonce"`
	Amount    BigInt `json:"amount"`
	Price     BigInt `json:"price"`
}
type Position struct {
}

type VaultSubscription struct {
	LiquidityProviderState LiquidityProviderState `json:"liquidity_provider_state"`
	OptionBuyerState       OptionBuyer            `json:"option_buyer_state"`
	VaultState             VaultState             `json:"vault_state"`
	OptionRoundState       OptionRound            `json:"option_round_state"`
}
