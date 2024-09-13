package db

import (
	"context"
	"database/sql"
	"fmt"
	"log"

	"github.com/jackc/pgx/v5"
)

type DB struct {
	conn *sql.DB
}

func (db *DB) Init(conninfo string) {
	connStr := "postgres://username:password@localhost:5432/mydb"
	conn, err := pgx.Connect(context.Background(), connStr)
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close(context.Background())

	// Set up listener for notifications
	_, err = conn.Exec(context.Background(), "LISTEN multi_row_change")
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Waiting for notifications...")

	for {
		// Wait for a notification
		notification, err := conn.WaitForNotification(context.Background())
		if err != nil {
			log.Fatal(err)
		}
		//Process notification here
		fmt.Printf("Received notification: %s\n", notification.Payload)
	}
}
