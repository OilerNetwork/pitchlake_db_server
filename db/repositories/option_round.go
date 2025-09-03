package repositories

import (
	"context"
	"pitchlake-backend/models"

	"github.com/jackc/pgx/v5/pgxpool"
)

// OptionRepository handles option-related database operations
type OptionRepository struct {
	pool *pgxpool.Pool
}

// NewOptionRepository creates a new option repository
func NewOptionRepository(pool *pgxpool.Pool) *OptionRepository {
	return &OptionRepository{pool: pool}
}

// GetOptionRoundsByVaultAddress retrieves option rounds for a specific vault
func (r *OptionRepository) GetOptionRoundsByVaultAddress(ctx context.Context, vaultAddress string) ([]*models.OptionRound, error) {
	var optionRounds []*models.OptionRound
	query := `
	SELECT 
		address, vault_address, round_id, cap_level, start_date, end_date, settlement_date, 
		starting_liquidity, queued_liquidity, remaining_liquidity, unsold_liquidity, available_options, reserve_price, 
		settlement_price, strike_price, sold_options, clearing_price, state, 
		premiums, payout_per_option, deployment_date
	FROM 
		public."Option_Rounds" 
	WHERE 
		vault_address = $1 
	ORDER BY 
		round_id ASC;`

	rows, err := r.pool.Query(ctx, query, vaultAddress)
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
		optionRounds = append(optionRounds, optionRound)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return optionRounds, nil
}

// GetOptionRoundByAddress retrieves an option round by its address
func (r *OptionRepository) GetOptionRoundByAddress(ctx context.Context, address string) (*models.OptionRound, error) {
	var optionRound models.OptionRound
	query := `SELECT address, round_id, bids, cap_level, starting_block, ending_block, settlement_date, starting_liquidity, queued_liquidity, remaining_liquidity, unsold_liquidity, available_options, settlement_price, strike_price, sold_options, clearing_price, state, premiums, payout_per_option, deployment_date FROM public."Option_Rounds" WHERE address=$1`

	err := r.pool.QueryRow(ctx, query, address).Scan(
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
