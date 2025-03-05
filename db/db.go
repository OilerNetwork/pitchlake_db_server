package db

import (
	"context"
	"fmt"
	"log"
	"os"
	"pitchlake-backend/models"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type DB struct {
	Conn *pgx.Conn
	Pool *pgxpool.Pool
}

func (db *DB) Init() error {
	connStr := os.Getenv("DB_URL")
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
	query := `SELECT current_round, current_round_address, unlocked_balance, locked_balance, stashed_balance, address, latest_block, deployment_date, fossil_client_address, eth_address, option_round_class_hash, alpha, strike_level, auction_duration, round_duration, round_transition_period FROM public."VaultStates" WHERE address=$1`

	err := db.Pool.QueryRow(ctx, query, id).Scan(
		&vaultState.CurrentRound,
		&vaultState.CurrentRoundAddress,
		&vaultState.UnlockedBalance,
		&vaultState.LockedBalance,
		&vaultState.StashedBalance,
		&vaultState.Address,
		&vaultState.LatestBlock,
		&vaultState.DeploymentDate,
		&vaultState.FossilClientAddress,
		&vaultState.EthAddress,
		&vaultState.OptionRoundClassHash,
		&vaultState.Alpha,
		&vaultState.StrikeLevel,
		&vaultState.AuctionRunTime,
		&vaultState.OptionRunTime,
		&vaultState.RoundTransitionPeriod,
	)

	if err != nil {
		fmt.Printf("Error getting vault state by id %s", err.Error())
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
    address, vault_address, round_id, cap_level, start_date, end_date, settlement_date, 
    starting_liquidity, queued_liquidity,remaining_liquidity, unsold_liquidity, available_options, reserve_price, 
    settlement_price, strike_price, sold_options, clearing_price, state, 
    premiums, payout_per_option, deployment_date
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
			&optionRound.VaultAddress,
			&optionRound.RoundID,
			&optionRound.CapLevel,
			&optionRound.AuctionStartDate,
			&optionRound.AuctionEndDate,
			&optionRound.OptionSettleDate,
			&optionRound.StartingLiquidity,
			&optionRound.QueuedLiquidity,
			&optionRound.RemainingLiquidity,
			&optionRound.UnsoldLiquidity,
			&optionRound.AvailableOptions,
			&optionRound.ReservePrice,
			&optionRound.SettlementPrice,
			&optionRound.StrikePrice,
			&optionRound.OptionsSold,
			&optionRound.ClearingPrice,
			&optionRound.RoundState,
			&optionRound.Premiums,
			&optionRound.PayoutPerOption,
			&optionRound.DeploymentDate,
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
func (db *DB) GetBlocks(startTimestamp, endTimestamp, roundDuration uint64) ([]models.Block, error) {
	query := `SELECT block_number, timestamp, basefee, is_confirmed, twelve_min_twap,three_hour_twap,thirty_day_twap 
	FROM public."blocks" 
	WHERE timestamp BETWEEN $1 AND $2
	AND block_number % 4 = 0
	ORDER BY block_number ASC
	`

	var blocks []models.Block
	rows, err := db.Pool.Query(context.Background(), query, startTimestamp, endTimestamp)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var block models.Block
		err := rows.Scan(
			&block.BlockNumber,
			&block.Timestamp,
			&block.BaseFee,
			&block.IsConfirmed,
			&block.TwelveMinTwap,
			&block.ThreeHourTwap,
			&block.ThirtyDayTwap,
		)
		if err != nil {
			return nil, err
		}
		blocks = append(blocks, block)
	}

	if rows.Err() != nil {
		return nil, rows.Err()
	}

	return blocks, nil
}

// GetAllVaultStates retrieves all VaultState records from the database
func (db *DB) GetAllVaultStates() ([]models.VaultState, error) {
	query := `SELECT current_round, current_round_address, unlocked_balance, locked_balance, stashed_balance, address, last_block FROM public."VaultStates"`
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

func (db *DB) GetOptionRoundByAddress(address string) (*models.OptionRound, error) {
	var optionRound models.OptionRound
	query := `SELECT address, round_id, bids, cap_level, starting_block, ending_block, settlement_date, starting_liquidity, queued_liquidity,remaining_liquidity, unsold_liquidity, available_options, settlement_price, strike_price, sold_options, clearing_price, state, premiums, payout_per_option, deployment_date FROM public."Option_Rounds" WHERE address=$1`
	err := db.Pool.QueryRow(context.Background(), query, address).Scan(
		&optionRound.Address,
		&optionRound.RoundID,
		&optionRound.CapLevel,
		&optionRound.AuctionStartDate,
		&optionRound.AuctionEndDate,
		&optionRound.OptionSettleDate,
		&optionRound.StartingLiquidity,
		&optionRound.QueuedLiquidity,
		&optionRound.RemainingLiquidity,
		&optionRound.UnsoldLiquidity,
		&optionRound.AvailableOptions,
		&optionRound.SettlementPrice,
		&optionRound.StrikePrice,
		&optionRound.OptionsSold,
		&optionRound.ClearingPrice,
		&optionRound.RoundState,
		&optionRound.Premiums,
		&optionRound.PayoutPerOption,
		&optionRound.DeploymentDate,
	)
	if err != nil {
		return nil, err
	}
	return &optionRound, nil
}
func (db *DB) GetVaultAddresses() ([]string, error) {
	var vaultAddresses []string

	query := `
	SELECT address 
	FROM "VaultStates" ;`

	rows, err := db.Pool.Query(context.Background(), query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var address string
		if err := rows.Scan(&address); err != nil {
			return nil, err
		}
		vaultAddresses = append(vaultAddresses, address)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return vaultAddresses, nil
}

// GetLiquidityProviderStateByID retrieves a LiquidityProviderState record by its Address
func (db *DB) GetLiquidityProviderStateByAddress(address, vaultAddress string) (*models.LiquidityProviderState, error) {
	var liquidityProviderState models.LiquidityProviderState

	query := `SELECT address, vault_address, unlocked_balance, locked_balance, stashed_balance, latest_block FROM public."Liquidity_Providers" WHERE address=$1 AND vault_address=$2`
	err := db.Pool.QueryRow(context.Background(), query, address, vaultAddress).Scan(
		&liquidityProviderState.Address,
		&liquidityProviderState.VaultAddress,
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

// GetOptionBuyerByID retrieves an OptionBuyer record by its Address
func (db *DB) GetOptionBuyerByAddress(address string) ([]*models.OptionBuyer, error) {
	var optionBuyers []*models.OptionBuyer
	query := `SELECT address, round_address, mintable_options, refundable_amount, has_minted, has_refunded 
	          FROM public."Option_Buyers" WHERE address=$1`

	rows, err := db.Pool.Query(context.Background(), query, address)
	if err != nil {
		if err == pgx.ErrNoRows {
			// Return an empty slice if no option buyers are found
			return []*models.OptionBuyer{}, nil
		}
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var optionBuyer models.OptionBuyer
		err := rows.Scan(
			&optionBuyer.Address,
			&optionBuyer.RoundAddress,
			&optionBuyer.MintableOptions,
			&optionBuyer.RefundableOptions,
			&optionBuyer.HasMinted,
			&optionBuyer.HasRefunded,
		)
		if err != nil {
			return nil, err
		}

		// Fetch associated bids for this optionBuyer
		bidQuery := `SELECT buyer_address, round_address, bid_id, tree_nonce, amount, price 
		             FROM public."Bids" WHERE buyer_address=$1 AND round_address=$2`
		bidRows, err := db.Pool.Query(context.Background(), bidQuery, optionBuyer.Address, optionBuyer.RoundAddress)

		if err != nil {
			if err == pgx.ErrNoRows {
				// If no rows are found, initialize an empty slice for bids
				optionBuyer.Bids = []*models.Bid{}
			} else {
				return nil, err
			}
		} else {
			defer bidRows.Close()

			var bids []*models.Bid
			for bidRows.Next() {
				var bid models.Bid
				err := bidRows.Scan(
					&bid.BuyerAddress,
					&bid.RoundAddress,
					&bid.BidID,
					&bid.TreeNonce,
					&bid.Amount,
					&bid.Price,
				)
				if err != nil {
					return nil, err
				}
				bids = append(bids, &bid)
			}

			// Check for errors after finishing iteration
			if err = bidRows.Err(); err != nil {
				return nil, err
			}

			// Attach bids to the optionBuyer
			optionBuyer.Bids = bids
		}

		optionBuyers = append(optionBuyers, &optionBuyer)
	}

	// Check for errors after finishing iteration
	if err = rows.Err(); err != nil {
		return nil, err
	}

	return optionBuyers, nil
}

// GetAllOptionBuyers retrieves all OptionBuyer records from the database
