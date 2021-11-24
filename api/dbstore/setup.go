package dbstore

import (
	"fmt"
	"database/sql"

	"github.com/jackc/pgx"
)

const (
	host = "localhost"
	port = 5432
	user = "polka"
	password = "12345678"
	dbname = "payments"
)

