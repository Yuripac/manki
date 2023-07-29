package storage

import (
	"database/sql"
)

type Config struct {
	DriverName  string
	DataSrcName string
}

type CmdFunc func(pool *sql.DB) (*sql.Rows, error)

func Conn(c Config, cmdFunc CmdFunc) (*sql.Rows, error) {
	pool, err := sql.Open(c.DriverName, c.DataSrcName)
	if err != nil {
		return nil, err
	}
	defer pool.Close()

	return cmdFunc(pool)
}
