package dbstore

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/joho/godotenv"

	"github.com/sekerez/polka/cache/utils"
)

const envPath = "db.env"

var db *DB // Create singleton DB

type DB struct {
	path    string
	ctx     context.Context
	logger  *log.Logger
	conn    *pgxpool.Pool
	quit    <-chan bool
	memChan <-chan utils.BankBalance
}

func New(ctx context.Context, quit <-chan bool, memChan <-chan utils.BankBalance) error {

	logger := log.New(os.Stderr, "[postgres] ", log.LstdFlags)

	// Get environment variables and format url
	if err := godotenv.Load(envPath); err != nil {
		return err
	}

	// Write db url
	path := fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s",
		os.Getenv("DBUSER"),
		os.Getenv("DBPASS"),
		os.Getenv("DBHOST"),
		os.Getenv("DBPORT"),
		os.Getenv("DBNAME"),
	)

	// Connect to database
	conn, err := pgxpool.Connect(ctx, path) // ConnPool?
	if err != nil {
		return err
	}
	log.Printf("Max Connections: %d", conn.Stat().MaxConns())

	// Insert variables inside object
	db = &DB{
		path:    path,
		ctx:     ctx,
		conn:    conn,
		quit:    quit,
		logger:  logger,
		memChan: memChan,
	}
	// Set up periodic update
	go periodicallyUpdateDues()

	return nil
}

func periodicallyUpdateDues() {
	for {
		select {
		case <-db.quit:
			return
		case current := <-db.memChan:
			_, err := db.conn.Exec(
				db.ctx,
				addDues,
				current.Name,
				current.Balance,
			)
			if err != nil {
				db.logger.Printf("Error updating database: %s", err)
			}
		}
	}
}
