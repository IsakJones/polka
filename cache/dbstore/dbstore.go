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
	path     string
	ctx      context.Context
	logger   *log.Logger
	conn     *pgxpool.Pool
	quit     chan bool
	bankChan <-chan *utils.BankBalance
	accChan  <-chan *utils.Balance
}

func New(
	ctx context.Context,
	bankChan <-chan *utils.BankBalance,
	accChan <-chan *utils.Balance,
	bankRetChan chan<- *utils.BankBalance,
	accRetChan chan<- *utils.Balance,
) error {

	var (
		bank        string
		account     uint16
		bankBalance int64
		accBalance  int
	)

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
		ctx:      ctx,
		path:     path,
		conn:     conn,
		logger:   logger,
		accChan:  accChan,
		bankChan: bankChan,
		quit:     make(chan bool),
	}

	// Restore bank balances
	rows, err := db.conn.Query(
		db.ctx,
		bankRetrieveQ,
	)
	if err != nil {
		log.Fatalf("Could not retrieve bank balances: %s", err)
	}
	for rows.Next() {
		err = rows.Scan(&bank, &bankBalance)
		if err != nil {
			log.Printf("Could not retrieve bank balance row: %s", err)
		}

		bankRetChan <- &utils.BankBalance{
			Name:    bank,
			Balance: bankBalance,
		}
	}
	close(bankRetChan)

	// Restore account balances
	rows, err = db.conn.Query(
		db.ctx,
		accRetrieveQ,
	)
	if err != nil {
		log.Fatalf("Could not retrieve account balances: %s", err)
	}
	for rows.Next() {
		err = rows.Scan(&bank, &account, &accBalance)
		if err != nil {
			log.Printf("Could not retrieve account balance row: %s", err)
		}

		accRetChan <- &utils.Balance{
			Bank:    bank,
			Account: account,
			Balance: accBalance,
		}
	}
	close(accRetChan)

	// Set up periodic update
	go updateDatabase()

	return nil
}

func updateDatabase() {
	for {
		select {
		case <-db.quit:
			return
		case bankBalance := <-db.bankChan:
			_, err := db.conn.Exec(
				db.ctx,
				updateBankBalanceQ,
				bankBalance.Name,
				bankBalance.Balance,
			)
			if err != nil {
				db.logger.Printf("Error updating database: %s", err)
			}
		case accBalance := <-db.accChan:
			_, err := db.conn.Exec(
				db.ctx,
				updateAccBalanceQ,
				accBalance.Bank,
				accBalance.Account,
				accBalance.Balance,
			)
			if err != nil {
				db.logger.Printf("Error updating database: %s", err)
			}
		}
	}
}

func Close() (err error) {
	db.quit <- true
	db.conn.Close()
	return
}
