package repositories

import (
	"context"
	"pitchlake-backend/models"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// OptionBuyerRepository handles option buyer-related database operations
type OptionBuyerRepository struct {
	pool *pgxpool.Pool
}

// NewOptionBuyerRepository creates a new option buyer repository
func NewOptionBuyerRepository(pool *pgxpool.Pool) *OptionBuyerRepository {
	return &OptionBuyerRepository{pool: pool}
}

// GetOptionBuyerByAddress retrieves an OptionBuyer record by its Address
func (r *OptionBuyerRepository) GetOptionBuyerByAddress(ctx context.Context, address string) ([]*models.OptionBuyer, error) {
	var optionBuyers []*models.OptionBuyer
	query := `SELECT address, round_address, mintable_options, refundable_amount, has_minted, has_refunded 
	          FROM public."Option_Buyers" WHERE address=$1`

	rows, err := r.pool.Query(ctx, query, address)
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
		bidRows, err := r.pool.Query(ctx, bidQuery, optionBuyer.Address, optionBuyer.RoundAddress)

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
