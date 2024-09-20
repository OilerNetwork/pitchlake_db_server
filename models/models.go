package models

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
	RoundID            string `json:"round_id"`
	TokenizableOptions string `json:"tokenizable_options"`
	RefundableBalance  string `json:"refundable_balance"`
}

type OptionRound struct {
	Address           string `json:"address"`
	RoundID           string `json:"round_id"`
	Bids              string `json:"bids"` // Store bids as JSON in PostgreSQL
	CapLevel          string `json:"cap_level"`
	StartingBlock     string `json:"starting_block"`
	EndingBlock       string `json:"ending_block"`
	SettlementDate    string `json:"settlement_date"`
	StartingLiquidity string `json:"starting_liquidity"`
	QueuedLiquidity   string `json:"queued_liquidity"`
	AvailableOptions  string `json:"available_options"`
	SettlementPrice   string `json:"settlement_price"`
	StrikePrice       string `json:"strike_price"`
	SoldOptions       string `json:"sold_options"`
	ClearingPrice     string `json:"clearing_price"`
	State             string `json:"state"`
	Premiums          string `json:"premiums"`
	PayoutPerOption   string `json:"payout_per_option"`
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
	RoundID        string `json:"round_id"`
	StartingAmount string `json:"starting_amount"`
	QueuedAmount   string `json:"amount"`
}

type Bid struct {
	Address   string `json:"address"`
	RoundID   string `json:"round_id"`
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
