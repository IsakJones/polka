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
	envPath = "dbstore/db.env"
)

type DB struct {
	Path   string
	Ctx    context.Context
	Logger log.Logger
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

func New(ctx context.Context) *DB {

	// Get environment variables and format url
	if err := godotenv.Load(envPath); err != nil {
		log.Fatalf("DB environmental variables failed to load: %s", err)
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
		log.Fatalf("Failed to establish a connection with the database server: %s", err)
	}

	// Insert variables inside object
	dbs := &DB{
		Path: path,
		Ctx:  ctx,
		Conn: conn,
	}

	return dbs
}

func (db *DB) InsertTransaction(trans *utils.Transaction) { //error {

}

// connConf := pgx.ConnConfig{
// 	Host:     os.Getenv("DBHOST"),
// 	Port:     uint16(os.Getenv("DBPORT")),
// 	User:     os.Getenv("DBUSER"),
// 	Password: os.Getenv("DBPASS"),
// 	Database: os.Getenv("DBNAME"),
// 	Logger:   log.New(os.Stderr, "[postgresql] "),
// }
