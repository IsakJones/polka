package dbstore

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/joho/godotenv"

	"github.com/sekerez/polka/apis/src/utils"
)

const (
	envPath = "db.env"
)

// Create singleton DB
var db *DB

type DB struct {
	path   string
	ctx    context.Context
	logger *log.Logger
	conn   *pgxpool.Pool
}

func (db *DB) GetPath() string {
	return db.path
}

func (db *DB) GetCtx() context.Context {
	return db.ctx
}

func (db *DB) GetConn() *pgxpool.Pool {
	return db.conn
}

func New(ctx context.Context) error {

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
		path:   path,
		ctx:    ctx,
		conn:   conn,
		logger: logger,
	}

	return nil
}

func GetTransaction(ctx context.Context, destTransaction utils.Transaction) error {
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

func InsertTransaction(ctx context.Context, transaction utils.Transaction) error {
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

func DeleteTransaction(ctx context.Context, transaction utils.Transaction) error {
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
