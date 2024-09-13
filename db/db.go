package db

import (
	"context"
	"log"

	"github.com/jackc/pgx/v5"
)

type DB struct {
	Conn *pgx.Conn
}

// LP Trigger: lp_row_update
// Vault Trigger: vault_update
// State Transition: state_transition(can be OR trigger on the state field)
// OB Trigger: ob_update
// OR Trigger:or_update
func (db *DB) Init(conninfo string) {
	connStr := "postgres://username:password@localhost:5432/mydb"
	conn, err := pgx.Connect(context.Background(), connStr)
	if err != nil {
		log.Fatal(err)
	}
	db.Conn = conn
	defer conn.Close(context.Background())

	// Set up listener for notifications

}
