package db

import (
	"database/sql"
	"os"
)

const (
	DB = "mysql"
)

var pool *sql.DB

func Init() (err error) {
	pool, err = sql.Open(DB, os.Getenv("DB_DSN"))

	if err != nil {
		return err
	}

	return nil
}

func Pool() *sql.DB {
	return pool
}
