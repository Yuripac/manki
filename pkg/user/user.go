package user

import (
	"context"
	"database/sql"
)

type User struct {
	Id   int32  `json:"id"`
	Name string `json:"name"`
}

func Exists(ctx context.Context, pool *sql.DB, id int32) (bool, error) {
	q := `SELECT id FROM users WHERE id = ?`

	row := pool.QueryRowContext(ctx, q, id)

	var userId int
	switch err := row.Scan(&userId); err {
	case sql.ErrNoRows:
		return false, nil
	case nil:
		return true, nil
	default:
		return false, err
	}
}
