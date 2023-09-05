package db

import (
	"database/sql"
	"os"
)

const (
	DB = "mysql"
)

func Open() (*sql.DB, error) {
	return sql.Open(DB, os.Getenv("DB_DSN"))
}
