package dbstore

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/joho/godotenv"

	"github.com/sekerez/polka/receiver/src/utils"
)

const (
	envPath = "postgres.env"
)

// Create singleton DB
var db *DB

type DB struct {
	ctx    context.Context
	conn   *pgxpool.Pool
	logger *log.Logger
}

func New(ctx context.Context) error {

	logger := log.New(os.Stderr, "[postgres] ", log.LstdFlags)

	// Get environment variables and format url
	if err := godotenv.Load(envPath); err != nil {
		return err
	}

	// Write db url
	uri := fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s",
		os.Getenv("POSTGRESUSER"),
		os.Getenv("POSTGRESPASS"),
		os.Getenv("POSTGRESHOST"),
		os.Getenv("POSTGRESPORT"),
		os.Getenv("POSTGRESNAME"),
	)

	// Connect to database
	conn, err := pgxpool.Connect(ctx, uri)
	if err != nil {
		return err
	}
	// log.Printf("Max Connections: %d", conn.Stat().MaxConns())

	// Insert variables inside object
	db = &DB{
		ctx:    ctx,
		conn:   conn,
		logger: logger,
	}

	return nil
}

func GetTransaction(ctx context.Context, destTransaction utils.Payment) error {
	var (
		senBank string
		recBank string
		senAcc  int
		recAcc  int
		amount  int
		time    time.Time
		err     error
	)

	err = db.conn.QueryRow(
		ctx,
		getLatestTransactionQ,
	).Scan(
		&senBank,
		&recBank,
		&senAcc,
		&recAcc,
		&amount,
		&time,
	)
	if err != nil {
		return err
	}

	destTransaction.SetSenBank(senBank)
	destTransaction.SetRecBank(recBank)
	destTransaction.SetSenAcc(senAcc)
	destTransaction.SetRecAcc(recAcc)
	destTransaction.SetAmount(amount)
	destTransaction.SetTime(time)

	return err

}

func InsertPayment(ctx context.Context, transaction utils.Payment) error {
	_, err := db.conn.Exec(
		ctx,
		insertTransactionQ,
		transaction.GetSenBank(),
		transaction.GetRecBank(),
		transaction.GetSenAcc(),
		transaction.GetRecAcc(),
		transaction.GetAmount(),
		transaction.GetTime(),
	)
	return err
}

func DeletePayment(ctx context.Context, transaction utils.Payment) error {
	_, err := db.conn.Exec(
		ctx,
		deleteTransactionQ,
		transaction.GetSenBank(),
		transaction.GetRecBank(),
		transaction.GetSenAcc(),
		transaction.GetRecAcc(),
		transaction.GetAmount(),
		transaction.GetTime(),
	)
	return err
}
