package repositories

import (
	"context"
	"pitchlake-backend/models"

	"github.com/jackc/pgx/v5/pgxpool"
)

// LiquidityRepository handles liquidity-related database operations
type LiquidityRepository struct {
	pool *pgxpool.Pool
}

// NewLiquidityRepository creates a new liquidity repository
func NewLiquidityRepository(pool *pgxpool.Pool) *LiquidityRepository {
	return &LiquidityRepository{pool: pool}
}

// GetLiquidityProviderStateByAddress retrieves a LiquidityProviderState record by its Address
func (r *LiquidityRepository) GetLiquidityProviderStateByAddress(ctx context.Context, address, vaultAddress string) (*models.LiquidityProviderState, error) {
	var liquidityProviderState models.LiquidityProviderState

	query := `SELECT address, vault_address, unlocked_balance, locked_balance, stashed_balance, latest_block FROM public."Liquidity_Providers" WHERE address=$1 AND vault_address=$2`
	err := r.pool.QueryRow(ctx, query, address, vaultAddress).Scan(
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
