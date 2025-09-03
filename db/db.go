package db

import (
	"context"
	"fmt"
	"os"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// DB manages the database connection pool
type DB struct {
	Pool *pgxpool.Pool
	Conn *pgx.Conn
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
		Pool: pool,
		Conn: conn,
	}, nil
}

// GetPool returns the connection pool for use by repositories
func (db *DB) GetPool() *pgxpool.Pool {
	return db.Pool
}

// Close closes all database connections
func (db *DB) Close() error {
	if db.Pool != nil {
		db.Pool.Close()
	}
	if db.Conn != nil {
		db.Conn.Close(context.Background())
	}
	return nil
}
