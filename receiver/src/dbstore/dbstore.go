package dbstore

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/joho/godotenv"

	"github.com/sekerez/polka/utils"
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

func GetPayment(ctx context.Context, paymnt *utils.Payment) error {
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
		getLatestPaymentQ,
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

	paymnt.Sender.Name = senBank
	paymnt.Receiver.Name = recBank
	paymnt.Sender.Account = senAcc
	paymnt.Receiver.Account = recAcc
	paymnt.Amount = amount
	paymnt.Time = time

	return err
}

func InsertPayment(ctx context.Context, paymnt *utils.Payment) error {
	_, err := db.conn.Exec(
		ctx,
		insertPaymentQ,
		paymnt.Sender.Name,
		paymnt.Receiver.Name,
		paymnt.Sender.Account,
		paymnt.Receiver.Account,
		paymnt.Amount,
		paymnt.Time,
	)
	return err
}

func DeletePayment(ctx context.Context, paymnt *utils.Payment) error {
	_, err := db.conn.Exec(
		ctx,
		deletePaymentQ,
		paymnt.Sender.Name,
		paymnt.Receiver.Name,
		paymnt.Sender.Account,
		paymnt.Receiver.Account,
		paymnt.Amount,
		paymnt.Time,
	)
	return err
}
