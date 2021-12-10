package dbstore

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	// "github.com/georgysavva/scany/pgxscan"
	"github.com/georgysavva/scany/pgxscan"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/joho/godotenv"

	"github.com/IsakJones/polka/api/utils"
)

const (
	envPath         = "dbstore/db.env"
	getTransByTrans = `
	SELECT 
		sending_bank_id AND
		receiving_bank_id AND
		sending_account AND
		receiving_account AND
		dollar_amount AND
		time 
	FROM transactions WHERE 
	sending_bank_id=(SELECT id FROM banks WHERE name=$1) AND
	receiving_bank_id=(SELECT id FROM banks WHERE name=$2) AND
	sending_account=$3 AND
	receiving_account=$4 AND
	dollar_amount=$5 AND
	time=$6;
	`
	getLatestTrans = `
	SELECT * FROM transactions
	ORDER BY time
	LIMIT 1;
	`
	getBankNames = `
	SELECT sender.name, receiver.name
	FROM banks sender
	INNER JOIN transactions ON sender.id=$1
	INNER JOIN banks receiver ON receiver.id=$2
	LIMIT 1;
	`
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

// Define transaction struct

type dbTransaction struct {
	ID                int
	Sending_account   int
	Receiving_account int
	Dollar_amount     int
	Time              time.Time
	Sending_bank_id   int
	Receiving_bank_id int
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

func GetTransaction(ctx context.Context, destTransaction utils.Transaction) error {
	var (
		transactionRows []*dbTransaction
		err             error
		senBank         string
		recBank         string
	)

	err = pgxscan.Select(ctx, db.Conn, &transactionRows, getLatestTrans)
	if err != nil {
		return err
	}

	err = db.Conn.QueryRow(
		ctx,
		getBankNames,
		transactionRows[0].Sending_bank_id,
		transactionRows[0].Receiving_bank_id,
	).Scan(
		&senBank,
		&recBank,
	)
	if err != nil {
		return err
	}

	// err = db.Conn.QueryRow(
	// 	ctx,
	// 	"select name from banks where id=$1;",
	// 	transactionRows[0].Receiving_bank_id,
	// ).Scan(&recBank)
	// if err != nil {
	// 	return err
	// }

	destTransaction.SetSenBank(senBank)
	destTransaction.SetRecBank(recBank)
	destTransaction.SetSenAcc(transactionRows[0].Sending_account)
	destTransaction.SetRecAcc(transactionRows[0].Receiving_account)
	destTransaction.SetAmount(transactionRows[0].Dollar_amount)
	destTransaction.SetTime(transactionRows[0].Time)

	return err
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
