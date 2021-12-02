package dbstore

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/jackc/pgx"
	"github.com/joho/godotenv"
)

const (
	envPath = "dbstore/db.env"
)

type DB struct {
	Path   string
	Ctx    context.Context
	Logger log.Logger
	Conn   *pgx.Conn
}

func (db *DB) GetPath() string {
	return db.Path
}

func (db *DB) GetCtx() context.Context {
	return db.Ctx
}

func (db *DB) GetConn() *pgx.Conn {
	return db.Conn
}

func New(ctx context.Context) *DB {

	// Get environment variables and format url
	if err := godotenv.Load(envPath); err != nil {
		log.Fatalf("DB environmental variables failed to load: %s", err)
	}

	// Write db url
	user := os.Getenv("DBUSER")
	pass := os.Getenv("DBPASS")
	host := os.Getenv("DBHOST")
	port := os.Getenv("DBPORT")
	name := os.Getenv("DBNAME")
	path := fmt.Sprintf("postgres://%s:%s@%s:%s/%s", user, pass, host, port, name)
	fmt.Println(path)

	// Parse connection url
	dbconfig, err := pgx.ParseConnectionString(path)
	if err != nil {
		log.Fatalf("Failed to parse connection url: %s", err)
	}

	// Connect to database
	conn, err := pgx.Connect(dbconfig)
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

// connConf := pgx.ConnConfig{
// 	Host:     os.Getenv("DBHOST"),
// 	Port:     uint16(os.Getenv("DBPORT")),
// 	User:     os.Getenv("DBUSER"),
// 	Password: os.Getenv("DBPASS"),
// 	Database: os.Getenv("DBNAME"),
// 	Logger:   log.New(os.Stderr, "[postgresql] "),
// }
