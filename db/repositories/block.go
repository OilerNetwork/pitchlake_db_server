package repositories

import (
	"context"
	"pitchlake-backend/models"

	"github.com/jackc/pgx/v5/pgxpool"
)

// BlockRepository handles block-related database operations
type BlockRepository struct {
	pool *pgxpool.Pool
}

// NewBlockRepository creates a new block repository
func NewBlockRepository(pool *pgxpool.Pool) *BlockRepository {
	return &BlockRepository{pool: pool}
}

// GetBlocks retrieves blocks within a time range with sampling
func (r *BlockRepository) GetBlocks(ctx context.Context, startTimestamp, endTimestamp, roundDuration uint64) ([]models.Block, error) {
	var samplingRate uint64
	switch roundDuration {
	case 960:
		samplingRate = 4
	case 13200:
		samplingRate = 5
	case 2631600:
		samplingRate = 40
	default:
		samplingRate = 1
	}

	query := `SELECT block_number, timestamp, basefee, is_confirmed, twelve_min_twap, three_hour_twap, thirty_day_twap 
	FROM public."blocks" 
	WHERE timestamp BETWEEN $1 AND $2
	AND block_number % $3 = 0
	ORDER BY block_number ASC`

	var blocks []models.Block
	rows, err := r.pool.Query(ctx, query, startTimestamp, endTimestamp, samplingRate)
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
