package repositories

import (
	"context"
	"fmt"
	"pitchlake-backend/models"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// VaultRepository handles vault-related database operations
type VaultRepository struct {
	pool *pgxpool.Pool
}

// NewVaultRepository creates a new vault repository
func NewVaultRepository(pool *pgxpool.Pool) *VaultRepository {
	return &VaultRepository{pool: pool}
}

// GetVaultStateByID retrieves a VaultState record by its ID
func (r *VaultRepository) GetVaultStateByID(ctx context.Context, id string) (*models.VaultState, error) {
	if r.pool == nil {
		return nil, fmt.Errorf("database pool is nil")
	}

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	var vaultState models.VaultState
	query := `SELECT current_round, current_round_address, unlocked_balance, locked_balance, stashed_balance, address, latest_block, deployment_date, fossil_client_address, eth_address, option_round_class_hash, alpha, strike_level, auction_duration, round_duration, round_transition_period FROM public."VaultStates" WHERE address=$1`

	err := r.pool.QueryRow(ctx, query, id).Scan(
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
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("no vault state found with id %s", id)
		}
		return nil, fmt.Errorf("error scanning vault state: %w", err)
	}
	return &vaultState, nil
}

// GetAllVaultStates retrieves all VaultState records from the database
func (r *VaultRepository) GetAllVaultStates(ctx context.Context) ([]models.VaultState, error) {
	query := `SELECT current_round, current_round_address, unlocked_balance, locked_balance, stashed_balance, address, last_block FROM public."VaultStates"`
	rows, err := r.pool.Query(ctx, query)
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

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return vaultStates, nil
}

// GetVaultAddresses retrieves all vault addresses
func (r *VaultRepository) GetVaultAddresses(ctx context.Context) ([]string, error) {
	var vaultAddresses []string

	query := `SELECT address FROM "VaultStates"`

	rows, err := r.pool.Query(ctx, query)
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
