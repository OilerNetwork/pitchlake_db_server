package models

type Vault struct {
	BlockNumber     BigInt `json:"blockNumber"`
	UnlockedBalance BigInt `json:"unlockedBalance"`
	LockedBalance   BigInt `json:"lockedBalance"`
	StashedBalance  BigInt `json:"stashedBalance"`
}

type LiquidityProvider struct {
	VaultAddress    string `json:"vaultAddress"`
	Address         string `json:"address"`
	BlockNumber     BigInt `json:"blockNumber"`
	UnlockedBalance BigInt `json:"unlockedBalance"`
	LockedBalance   BigInt `json:"lockedBalance"`
	StashedBalance  BigInt `json:"stashedBalance"`
}

type OptionBuyer struct {
	Address           string `json:"address"`
	RoundAddress      string `json:"roundAddress"`
	MintableOptions   BigInt `json:"mintableOptions"`
	HasMinted         bool   `json:"hasMinted"`
	HasRefunded       bool   `json:"hasRefunded"`
	RefundableOptions BigInt `json:"refundableOptions"`
	Bids              []*Bid `json:"bids"`
}

type OptionRound struct {
	VaultAddress       string `json:"vaultAddress"`
	Address            string `json:"address"`
	RoundID            BigInt `json:"roundId"`
	CapLevel           BigInt `json:"capLevel"`
	AuctionStartDate   string `json:"auctionStartDate"`
	AuctionEndDate     string `json:"auctionEndDate"`
	OptionSettleDate   string `json:"optionSettleDate"`
	StartingLiquidity  BigInt `json:"startingLiquidity"`
	QueuedLiquidity    BigInt `json:"queuedLiquidity"`
	RemainingLiquidity BigInt `json:"remainingLiquidity"`
	AvailableOptions   BigInt `json:"availableOptions"`
	ClearingPrice      BigInt `json:"clearingPrice"`
	SettlementPrice    BigInt `json:"settlementPrice"`
	ReservePrice       BigInt `json:"reservePrice"`
	StrikePrice        BigInt `json:"strikePrice"`
	OptionsSold        BigInt `json:"optionsSold"`
	UnsoldLiquidity    BigInt `json:"unsoldLiquidity"`
	RoundState         string `json:"roundState"`
	Premiums           BigInt `json:"premiums"`
	PayoutPerOption    BigInt `json:"payoutPerOption"`
	DeploymentDate     uint64 `json:"deploymentDate"`
}

type VaultState struct {
	CurrentRound          BigInt `json:"currentRoundId"`
	CurrentRoundAddress   string `json:"currentRoundAddress"`
	UnlockedBalance       BigInt `json:"unlockedBalance"`
	LockedBalance         BigInt `json:"lockedBalance"`
	StashedBalance        BigInt `json:"stashedBalance"`
	Address               string `json:"address"`
	LatestBlock           BigInt `json:"latestBlock"`
	DeploymentDate        uint64 `json:"deploymentDate"`
	FossilClientAddress   string `json:"fossilClientAddress"`
	EthAddress            string `json:"ethAddress"`
	OptionRoundClassHash  string `json:"optionRoundClassHash"`
	Alpha                 BigInt `json:"alpha"`
	StrikeLevel           BigInt `json:"strikeLevel"`
	AuctionRunTime        uint64 `json:"auctionRunTime"`
	OptionRunTime         uint64 `json:"optionRunTime"`
	RoundTransitionPeriod uint64 `json:"roundTransitionPeriod"`
}

type LiquidityProviderState struct {
	VaultAddress    string `json:"vaultAddress"`
	Address         string `json:"address"`
	UnlockedBalance BigInt `json:"unlockedBalance"`
	LockedBalance   BigInt `json:"lockedBalance"`
	StashedBalance  BigInt `json:"stashedBalance"`
	LatestBlock     BigInt `json:"latestBlock"`
}

type Bid struct {
	BuyerAddress string `json:"address"`
	RoundAddress string `json:"roundAddress"`
	BidID        string `json:"bidId"`
	TreeNonce    string `json:"treeNonce"`
	Amount       BigInt `json:"amount"`
	Price        BigInt `json:"price"`
}
