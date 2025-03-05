package models

type AllowedPayload interface {
	IsAllowedPayload() // Dummy method
}
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
	AuctionStartDate   uint64 `json:"auctionStartDate"`
	AuctionEndDate     uint64 `json:"auctionEndDate"`
	OptionSettleDate   uint64 `json:"optionSettleDate"`
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

type QueuedLiquidity struct {
	Address         string `json:"address"`
	RoundAddress    string `json:"roundAddress"`
	Bps             BigInt `json:"bps"`
	QueuedLiquidity BigInt `json:"queuedLiquidity"`
}
type Bid struct {
	BuyerAddress string `json:"address"`
	RoundAddress string `json:"roundAddress"`
	BidID        string `json:"bidId"`
	TreeNonce    string `json:"treeNonce"`
	Amount       BigInt `json:"amount"`
	Price        BigInt `json:"price"`
}

type Block struct {
	BlockNumber   uint64 `json:"blockNumber"`
	Timestamp     uint64 `json:"timestamp"`
	BaseFee       string `json:"baseFee"`
	IsConfirmed   bool   `json:"isConfirmed"`
	TwelveMinTwap string `json:"twelveMinTwap"`
	ThreeHourTwap string `json:"threeHourTwap"`
	ThirtyDayTwap string `json:"thirtyDayTwap"`
}

type TwapWindowType string

const (
	TwapWindowTwelveMin TwapWindowType = "twelve_min"
	TwapWindowThreeHour TwapWindowType = "three_hour"
	TwapWindowThirtyDay TwapWindowType = "thirty_day"
)

type TwapState struct {
	WindowType         TwapWindowType `json:"windowType"`
	WeightedSum        string         `json:"weightedSum"`
	TotalSeconds       BigInt         `json:"totalSeconds"`
	IsConfirmed        bool           `json:"isConfirmed"`
	TwapValue          string         `json:"twapValue"`
	LastBlockNumber    uint64         `json:"lastBlockNumber"`
	LastBlockTimestamp uint64         `json:"lastBlockTimestamp"`
}

func (Bid) IsAllowedPayload()                    {}
func (VaultState) IsAllowedPayload()             {}
func (LiquidityProviderState) IsAllowedPayload() {}
func (OptionRound) IsAllowedPayload()            {}
func (OptionBuyer) IsAllowedPayload()            {}
func (QueuedLiquidity) IsAllowedPayload()        {}
func (Block) IsAllowedPayload()                  {}
func (TwapState) IsAllowedPayload()              {}
