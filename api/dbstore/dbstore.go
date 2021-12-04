package dbstore

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/joho/godotenv"

	"github.com/IsakJones/polka/api/utils"
)

const (
	envPath        = "dbstore/db.env"
	insertTransSQL = `
	INSERT INTO transactions (
		sending_bank_id,
		receiving_bank_id,
		sending_account,
		receiving_account,
		dollar_amount,
		time
	) VALUES (
		(SELECT id FROM banks WHERE name=$1),
		(SELECT id FROM banks WHERE name=$2),
		$3,
		$4,
		$5,
		$6
	);
	`
)

// Create singleton DB
var db *DB

type DB struct {
	Path   string
	Ctx    context.Context
	Logger *log.Logger
	Conn   *pgxpool.Pool
}

func (db *DB) GetPath() string {
	return db.Path
}

func (db *DB) GetCtx() context.Context {
	return db.Ctx
}

func (db *DB) GetConn() *pgxpool.Pool {
	return db.Conn
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

	// Insert variables inside object
	db = &DB{
		Path:   path,
		Ctx:    ctx,
		Conn:   conn,
		Logger: logger,
	}

	return nil
}

func InsertTransaction(ctx context.Context, transaction utils.Transaction) error {

	_, err := db.Conn.Exec(
		ctx,
		insertTransSQL,
		transaction.GetSenBank(),
		transaction.GetRecBank(),
		transaction.GetSenAcc(),
		transaction.GetRecAcc(),
		transaction.GetAmount(),
		transaction.GetTime(),
	)
	return err
}
