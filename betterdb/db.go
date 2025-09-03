package betterdb

import (
	"context"
	"fmt"
	"os"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// DB manages the database connection pool
type DB struct {
	pool *pgxpool.Pool
	conn *pgx.Conn
}

// NewDB creates a new database connection
func NewDB() (*DB, error) {
	connStr := os.Getenv("DB_URL")
	if connStr == "" {
		return nil, fmt.Errorf("DB_URL environment variable not set")
	}

	config, err := pgxpool.ParseConfig(connStr)
	if err != nil {
		return nil, fmt.Errorf("unable to parse connection string: %w", err)
	}

	pool, err := pgxpool.NewWithConfig(context.Background(), config)
	if err != nil {
		return nil, fmt.Errorf("unable to create connection pool: %w", err)
	}

	conn, err := pgx.Connect(context.Background(), connStr)
	if err != nil {
		pool.Close()
		return nil, fmt.Errorf("unable to connect to database: %w", err)
	}

	return &DB{
		pool: pool,
		conn: conn,
	}, nil
}

// GetPool returns the connection pool for use by repositories
func (db *DB) GetPool() *pgxpool.Pool {
	return db.pool
}

// Close closes all database connections
func (db *DB) Close() error {
	if db.pool != nil {
		db.pool.Close()
	}
	if db.conn != nil {
		db.conn.Close(context.Background())
	}
	return nil
}
