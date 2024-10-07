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
	BlockNumber     BigInt `json:"blockNumber"`
	UnlockedBalance BigInt `json:"unlockedBalance"`
	LockedBalance   BigInt `json:"lockedBalance"`
	StashedBalance  BigInt `json:"stashedBalance"`
}

type LiquidityProvider struct {
	Address         string `json:"address"`
	UnlockedBalance BigInt `json:"unlockedBalance"`
	LockedBalance   BigInt `json:"lockedBalance"`
	StashedBalance  BigInt `json:"stashedBalance"`
}

type OptionBuyer struct {
	Address            string `json:"address"`
	RoundID            BigInt `json:"roundId"`
	TokenizableOptions BigInt `json:"tokenizableOptions"`
	RefundableBalance  BigInt `json:"refundableBalance"`
}

type OptionRound struct {
	Address           *string `json:"address"`
	RoundID           *BigInt `json:"roundId"`
	CapLevel          *BigInt `json:"capLevel"`
	AuctionStartDate  *string `json:"auctionStartDate"`
	AuctionEndDate    *string `json:"auctionEndDate"`
	OptionSettleDate  *string `json:"optionSettleDate"`
	StartingLiquidity *BigInt `json:"startingLiquidity"`
	AvailableOptions  *BigInt `json:"availableOptions"`
	ClearingPrice     *BigInt `json:"clearingPrice"`
	SettlementPrice   *BigInt `json:"settlementPrice"`
	StrikePrice       *BigInt `json:"strikePrice"`
	OptionsSold       *BigInt `json:"optionsSold"`
	RoundState        *string `json:"roundState"`
	Premiums          *BigInt `json:"premiums"`
	QueuedLiquidity   *BigInt `json:"queuedLiquidity"`
	PayoutPerOption   *BigInt `json:"payoutPerOption"`
	VaultAddress      *string `json:"vaultAddress"`
}

type VaultState struct {
	CurrentRound          BigInt `json:"currentRound"`
	CurrentRoundAddress   string `json:"currentRoundAddress"`
	UnlockedBalance       BigInt `json:"unlockedBalance"`
	LockedBalance         BigInt `json:"lockedBalance"`
	StashedBalance        BigInt `json:"stashedBalance"`
	Address               string `json:"address"`
	LatestBlock           BigInt `json:"lastBlock"`
	AuctionRunTime        BigInt `json:"auctionRunTime"`
	OptionRunTime         BigInt `json:"optionRunTime"`
	RoundTransitionPeriod BigInt `json:"roundTransitionPeriod"`
}

type LiquidityProviderState struct {
	Address         string `json:"address"`
	UnlockedBalance BigInt `json:"unlockedBalance"`
	LockedBalance   BigInt `json:"lockedBalance"`
	StashedBalance  BigInt `json:"stashedBalance"`
	LatestBlock     BigInt `json:"lastBlock"`
}

type Bid struct {
	Address   string `json:"address"`
	RoundID   BigInt `json:"roundId"`
	BidID     string `json:"bidId"`
	TreeNonce string `json:"treeNonce"`
	Amount    BigInt `json:"amount"`
	Price     BigInt `json:"price"`
}
type Position struct {
}
