package db

import (
	"context"
	"fmt"
	"log"
	"pitchlake-backend/models"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
)

var envFile, _ = godotenv.Read(".env")

type DB struct {
	Conn *pgx.Conn
	Pool *pgxpool.Pool
}

func (db *DB) Init() error {
	connStr := envFile["DB_URL"]
	config, err := pgxpool.ParseConfig(connStr)
	if err != nil {
		return fmt.Errorf("unable to parse connection string: %w", err)
	}

	conn, err := pgx.Connect(context.Background(), connStr)
	if err != nil {
		log.Fatal(err)
	}

	pool, err := pgxpool.NewWithConfig(context.Background(), config)
	if err != nil {
		return fmt.Errorf("unable to create connection pool: %w", err)
	}

	db.Conn = conn
	db.Pool = pool
	return nil
}

// GetVaultStateByID retrieves a VaultState record by its ID
func (db *DB) GetVaultStateByID(id string) (*models.VaultState, error) {
	if db.Pool == nil {
		return nil, fmt.Errorf("database pool is nil")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var vaultState models.VaultState
	query := `SELECT current_round, current_round_address, unlocked_balance, locked_balance, stashed_balance, address, latest_block FROM public."VaultStates" WHERE address=$1`

	err := db.Pool.QueryRow(ctx, query, id).Scan(
		&vaultState.CurrentRound,
		&vaultState.CurrentRoundAddress,
		&vaultState.UnlockedBalance,
		&vaultState.LockedBalance,
		&vaultState.StashedBalance,
		&vaultState.Address,
		&vaultState.LatestBlock,
	)

	if err != nil {
		fmt.Println("Error getting vault state by id %s", err.Error())
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("no vault state found with id %s", id)
		}
		return nil, fmt.Errorf("error scanning vault state: %w", err)
	}
	fmt.Printf("vaultState %+v", vaultState)
	return &vaultState, nil
}

func (db *DB) GetOptionRoundsByVaultAddress(vaultAddress string) ([]*models.OptionRound, error) {

	var optionRounds []*models.OptionRound
	query :=
		`
	SELECT 
    address, round_id, cap_level, start_date, end_date, settlement_date, 
    starting_liquidity, queued_liquidity, available_options, reserve_price, 
    settlement_price, strike_price, sold_options, clearing_price, state, 
    premiums, payout_per_option 
	FROM 
		public."Option_Rounds" 
	WHERE 
		vault_address = $1 
	ORDER BY 
		round_id ASC;`

	rows, err := db.Pool.Query(context.Background(), query, vaultAddress)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		optionRound := &models.OptionRound{}
		err := rows.Scan(
			&optionRound.Address,
			&optionRound.RoundID,
			&optionRound.CapLevel,
			&optionRound.AuctionStartDate,
			&optionRound.AuctionEndDate,
			&optionRound.OptionSettleDate,
			&optionRound.StartingLiquidity,
			&optionRound.QueuedLiquidity,
			&optionRound.AvailableOptions,
			&optionRound.ReservePrice,
			&optionRound.SettlementPrice,
			&optionRound.StrikePrice,
			&optionRound.OptionsSold,
			&optionRound.ClearingPrice,
			&optionRound.RoundState,
			&optionRound.Premiums,
			&optionRound.PayoutPerOption,
		)
		if err != nil {
			return nil, err
		}
		fmt.Printf("ROUND %v", optionRound)
		optionRounds = append(optionRounds, optionRound)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return optionRounds, nil
}

// GetAllVaultStates retrieves all VaultState records from the database
func (db *DB) GetAllVaultStates() ([]models.VaultState, error) {
	query := `SELECT current_round, current_round_address, unlocked_balance, locked_balance, stashed_balance, address, last_block FROM vault_states`
	rows, err := db.Pool.Query(context.Background(), query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var vaultStates []models.VaultState
	for rows.Next() {
		var vaultState models.VaultState
		err := rows.Scan(
			&vaultState.CurrentRound,
			&vaultState.CurrentRoundAddress,
			&vaultState.UnlockedBalance,
			&vaultState.LockedBalance,
			&vaultState.StashedBalance,
			&vaultState.Address,
			&vaultState.LatestBlock,
		)
		if err != nil {
			return nil, err
		}
		vaultStates = append(vaultStates, vaultState)
	}

	if rows.Err() != nil {
		return nil, rows.Err()
	}

	return vaultStates, nil
}

// GetOptionRoundByID retrieves an OptionRound record by its ID
func (db *DB) GetOptionRoundByID(id uint64) (*models.OptionRound, error) {
	var optionRound models.OptionRound
	query := `SELECT address, round_id, bids, cap_level, starting_block, ending_block, settlement_date, starting_liquidity, queued_liquidity, available_options, settlement_price, strike_price, sold_options, clearing_price, state, premiums, payout_per_option FROM option_rounds WHERE id=$1`
	err := db.Pool.QueryRow(context.Background(), query, id).Scan(
		&optionRound.Address,
		&optionRound.RoundID,
		&optionRound.CapLevel,
		&optionRound.AuctionStartDate,
		&optionRound.AuctionEndDate,
		&optionRound.OptionSettleDate,
		&optionRound.StartingLiquidity,
		&optionRound.QueuedLiquidity,
		&optionRound.AvailableOptions,
		&optionRound.SettlementPrice,
		&optionRound.StrikePrice,
		&optionRound.OptionsSold,
		&optionRound.ClearingPrice,
		&optionRound.RoundState,
		&optionRound.Premiums,
		&optionRound.PayoutPerOption,
	)
	if err != nil {
		return nil, err
	}
	return &optionRound, nil
}

func (db *DB) GetOptionRoundByAddress(address string) (*models.OptionRound, error) {
	var optionRound models.OptionRound
	query := `SELECT address, round_id, bids, cap_level, starting_block, ending_block, settlement_date, starting_liquidity, queued_liquidity, available_options, settlement_price, strike_price, sold_options, clearing_price, state, premiums, payout_per_option FROM option_rounds WHERE address=$1`
	err := db.Pool.QueryRow(context.Background(), query, address).Scan(
		&optionRound.Address,
		&optionRound.RoundID,
		&optionRound.CapLevel,
		&optionRound.AuctionStartDate,
		&optionRound.AuctionEndDate,
		&optionRound.OptionSettleDate,
		&optionRound.StartingLiquidity,
		&optionRound.QueuedLiquidity,
		&optionRound.AvailableOptions,
		&optionRound.SettlementPrice,
		&optionRound.StrikePrice,
		&optionRound.OptionsSold,
		&optionRound.ClearingPrice,
		&optionRound.RoundState,
		&optionRound.Premiums,
		&optionRound.PayoutPerOption,
	)
	if err != nil {
		return nil, err
	}
	return &optionRound, nil
}

// GetAllOptionRounds retrieves all OptionRound records from the database
func (db *DB) GetAllOptionRounds() ([]models.OptionRound, error) {
	query := `SELECT address, round_id, bids, cap_level, starting_block, ending_block, settlement_date, starting_liquidity, queued_liquidity, available_options, settlement_price, strike_price, sold_options, clearing_price, state, premiums, payout_per_option FROM option_rounds`
	rows, err := db.Pool.Query(context.Background(), query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var optionRounds []models.OptionRound
	for rows.Next() {
		var optionRound models.OptionRound
		err := rows.Scan(
			&optionRound.Address,
			&optionRound.RoundID,
			&optionRound.CapLevel,
			&optionRound.AuctionStartDate,
			&optionRound.AuctionEndDate,
			&optionRound.OptionSettleDate,
			&optionRound.StartingLiquidity,
			&optionRound.QueuedLiquidity,
			&optionRound.AvailableOptions,
			&optionRound.SettlementPrice,
			&optionRound.StrikePrice,
			&optionRound.OptionsSold,
			&optionRound.ClearingPrice,
			&optionRound.RoundState,
			&optionRound.Premiums,
			&optionRound.PayoutPerOption,
		)
		if err != nil {
			return nil, err
		}
		optionRounds = append(optionRounds, optionRound)
	}

	if rows.Err() != nil {
		return nil, rows.Err()
	}

	return optionRounds, nil
}

// GetLiquidityProviderStateByID retrieves a LiquidityProviderState record by its Address
func (db *DB) GetLiquidityProviderStateByAddress(address string) (*models.LiquidityProviderState, error) {
	var liquidityProviderState models.LiquidityProviderState

	query := `SELECT address, unlocked_balance, locked_balance, stashed_balance, latest_block FROM public."Liquidity_Providers" WHERE address=$1`
	err := db.Pool.QueryRow(context.Background(), query, address).Scan(
		&liquidityProviderState.Address,
		&liquidityProviderState.UnlockedBalance,
		&liquidityProviderState.LockedBalance,
		&liquidityProviderState.StashedBalance,
		&liquidityProviderState.LatestBlock,
	)
	if err != nil {
		return nil, err
	}
	return &liquidityProviderState, nil
}

// GetAllLiquidityProviderStates retrieves all LiquidityProviderState records from the database
func (db *DB) GetAllLiquidityProviderStates() ([]models.LiquidityProviderState, error) {
	query := `SELECT address, unlocked_balance, locked_balance, stashed_balance, last_block FROM liquidity_provider_states`
	rows, err := db.Pool.Query(context.Background(), query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var liquidityProviderStates []models.LiquidityProviderState
	for rows.Next() {
		var liquidityProviderState models.LiquidityProviderState
		err := rows.Scan(
			&liquidityProviderState.Address,
			&liquidityProviderState.UnlockedBalance,
			&liquidityProviderState.LockedBalance,
			&liquidityProviderState.StashedBalance,
			&liquidityProviderState.LatestBlock,
		)
		if err != nil {
			return nil, err
		}
		liquidityProviderStates = append(liquidityProviderStates, liquidityProviderState)
	}

	if rows.Err() != nil {
		return nil, rows.Err()
	}

	return liquidityProviderStates, nil
}

// GetOptionBuyerByID retrieves an OptionBuyer record by its Address
func (db *DB) GetOptionBuyerByAddress(address string) (*models.OptionBuyer, error) {
	var optionBuyer models.OptionBuyer
	query := `SELECT address, round_id, tokenizable_options, refundable_balance FROM option_buyers WHERE address=$1`
	err := db.Pool.QueryRow(context.Background(), query, address).Scan(
		&optionBuyer.Address,
		&optionBuyer.RoundID,
		&optionBuyer.TokenizableOptions,
		&optionBuyer.RefundableBalance,
	)
	if err != nil {
		return nil, err
	}
	return &optionBuyer, nil
}

// GetAllOptionBuyers retrieves all OptionBuyer records from the database
func (db *DB) GetAllOptionBuyers() ([]models.OptionBuyer, error) {
	query := `SELECT address, round_id, tokenizable_options, refundable_balance FROM option_buyers`
	rows, err := db.Pool.Query(context.Background(), query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var optionBuyers []models.OptionBuyer
	for rows.Next() {
		var optionBuyer models.OptionBuyer
		err := rows.Scan(
			&optionBuyer.Address,
			&optionBuyer.RoundID,
			&optionBuyer.TokenizableOptions,
			&optionBuyer.RefundableBalance,
		)
		if err != nil {
			return nil, err
		}
		optionBuyers = append(optionBuyers, optionBuyer)
	}

	if rows.Err() != nil {
		return nil, rows.Err()
	}

	return optionBuyers, nil
}
