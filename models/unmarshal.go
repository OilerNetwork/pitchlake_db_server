package models

import "encoding/json"

func (lps *LiquidityProviderState) UnmarshalJSON(data []byte) error {
	// Auxiliary struct to map JSON keys
	aux := struct {
		VaultAddress    string `json:"vault_address"`
		Address         string `json:"address"`
		UnlockedBalance BigInt `json:"unlocked_balance"`
		LockedBalance   BigInt `json:"locked_balance"`
		StashedBalance  BigInt `json:"stashed_balance"`
		LatestBlock     BigInt `json:"latest_block"`
	}{}

	// Unmarshal into the auxiliary struct
	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}

	// Copy data from aux to the original struct
	lps.VaultAddress = aux.VaultAddress
	lps.Address = aux.Address
	lps.UnlockedBalance = aux.UnlockedBalance
	lps.LockedBalance = aux.LockedBalance
	lps.StashedBalance = aux.StashedBalance
	lps.LatestBlock = aux.LatestBlock

	return nil
}
func (vs *VaultState) UnmarshalJSON(data []byte) error {
	// Auxiliary struct to map JSON keys
	aux := struct {
		CurrentRound          BigInt `json:"current_round"`
		CurrentRoundAddress   string `json:"current_round_address"`
		UnlockedBalance       BigInt `json:"unlocked_balance"`
		LockedBalance         BigInt `json:"locked_balance"`
		StashedBalance        BigInt `json:"stashed_balance"`
		Address               string `json:"address"`
		LatestBlock           BigInt `json:"latest_block"`
		DeploymentDate        uint64 `json:"deployment_date"`
		FossilClientAddress   string `json:"fossil_client_address"`
		EthAddress            string `json:"eth_address"`
		OptionRoundClassHash  string `json:"option_round_class_hash"`
		Alpha                 BigInt `json:"alpha"`
		StrikeLevel           BigInt `json:"strike_level"`
		AuctionRunTime        uint64 `json:"auction_duration"`
		OptionRunTime         uint64 `json:"round_duration"`
		RoundTransitionPeriod uint64 `json:"round_transition_period"`
	}{}

	// Unmarshal into the auxiliary struct
	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}

	// Copy data from aux to the original struct
	vs.CurrentRound = aux.CurrentRound
	vs.CurrentRoundAddress = aux.CurrentRoundAddress
	vs.UnlockedBalance = aux.UnlockedBalance
	vs.LockedBalance = aux.LockedBalance
	vs.StashedBalance = aux.StashedBalance
	vs.Address = aux.Address
	vs.LatestBlock = aux.LatestBlock
	vs.DeploymentDate = aux.DeploymentDate
	vs.FossilClientAddress = aux.FossilClientAddress
	vs.EthAddress = aux.EthAddress
	vs.OptionRoundClassHash = aux.OptionRoundClassHash
	vs.Alpha = aux.Alpha
	vs.StrikeLevel = aux.StrikeLevel
	vs.AuctionRunTime = aux.AuctionRunTime
	vs.OptionRunTime = aux.OptionRunTime
	vs.RoundTransitionPeriod = aux.RoundTransitionPeriod

	return nil
}
func (b *Bid) UnmarshalJSON(data []byte) error {
	// Auxiliary struct to map JSON keys
	aux := struct {
		BuyerAddress string `json:"address"`
		RoundAddress string `json:"round_address"`
		BidID        string `json:"bid_id"`
		TreeNonce    string `json:"tree_nonce"`
		Amount       BigInt `json:"amount"`
		Price        BigInt `json:"price"`
	}{}

	// Unmarshal into the auxiliary struct
	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}

	// Copy data from aux to the original struct
	b.BuyerAddress = aux.BuyerAddress
	b.RoundAddress = aux.RoundAddress
	b.BidID = aux.BidID
	b.TreeNonce = aux.TreeNonce
	b.Amount = aux.Amount
	b.Price = aux.Price

	return nil
}

func (ob *OptionBuyer) UnmarshalJSON(data []byte) error {
	// Auxiliary struct to map JSON keys
	aux := struct {
		Address           string `json:"address"`
		RoundAddress      string `json:"round_address"`
		MintableOptions   BigInt `json:"mintable_options"`
		HasMinted         bool   `json:"has_minted"`
		HasRefunded       bool   `json:"has_refunded"`
		RefundableOptions BigInt `json:"refundable_amount"`
		Bids              []*Bid `json:"bids"`
	}{}

	// Unmarshal into the auxiliary struct
	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}

	// Copy data from aux to the original struct
	ob.Address = aux.Address
	ob.RoundAddress = aux.RoundAddress
	ob.MintableOptions = aux.MintableOptions
	ob.HasMinted = aux.HasMinted
	ob.HasRefunded = aux.HasRefunded
	ob.RefundableOptions = aux.RefundableOptions
	ob.Bids = aux.Bids

	return nil
}

func (or *OptionRound) UnmarshalJSON(data []byte) error {
	// Auxiliary struct to map JSON keys
	aux := struct {
		VaultAddress       string `json:"vault_address"`
		Address            string `json:"address"`
		RoundID            BigInt `json:"round_id"`
		CapLevel           BigInt `json:"cap_level"`
		AuctionStartDate   uint64 `json:"start_date"`
		AuctionEndDate     uint64 `json:"end_date"`
		OptionSettleDate   uint64 `json:"settle_date"`
		StartingLiquidity  BigInt `json:"starting_liquidity"`
		QueuedLiquidity    BigInt `json:"queued_liquidity"`
		RemainingLiquidity BigInt `json:"remaining_liquidity"`
		AvailableOptions   BigInt `json:"available_options"`
		ClearingPrice      BigInt `json:"clearing_price"`
		SettlementPrice    BigInt `json:"settlement_price"`
		ReservePrice       BigInt `json:"reserve_price"`
		StrikePrice        BigInt `json:"strike_price"`
		OptionsSold        BigInt `json:"sold_options"`
		UnsoldLiquidity    BigInt `json:"unsold_liquidity"`
		RoundState         string `json:"state"`
		Premiums           BigInt `json:"premiums"`
		PayoutPerOption    BigInt `json:"payout_per_option"`
		DeploymentDate     uint64 `json:"deployment_date"`
	}{}

	// Unmarshal into the auxiliary struct
	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}

	// Copy data from aux to the original struct
	or.VaultAddress = aux.VaultAddress
	or.Address = aux.Address
	or.RoundID = aux.RoundID
	or.CapLevel = aux.CapLevel
	or.AuctionStartDate = aux.AuctionStartDate
	or.AuctionEndDate = aux.AuctionEndDate
	or.OptionSettleDate = aux.OptionSettleDate
	or.StartingLiquidity = aux.StartingLiquidity
	or.QueuedLiquidity = aux.QueuedLiquidity
	or.RemainingLiquidity = aux.RemainingLiquidity
	or.AvailableOptions = aux.AvailableOptions
	or.ClearingPrice = aux.ClearingPrice
	or.SettlementPrice = aux.SettlementPrice
	or.ReservePrice = aux.ReservePrice
	or.StrikePrice = aux.StrikePrice
	or.OptionsSold = aux.OptionsSold
	or.UnsoldLiquidity = aux.UnsoldLiquidity
	or.RoundState = aux.RoundState
	or.Premiums = aux.Premiums
	or.PayoutPerOption = aux.PayoutPerOption
	or.DeploymentDate = aux.DeploymentDate

	return nil
}
