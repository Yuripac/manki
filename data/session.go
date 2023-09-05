package data

import (
	"context"
	"database/sql"
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

type SignUp struct {
	Name  string `json:"name"`
	Email string `json:"email"`
	Psw   string `json:"password"`
}

type SignIn struct {
	Email string `json:"email"`
	Psw   string `json:"password"`
}

func (up SignUp) Save(ctx context.Context, pool *sql.DB) (*User, error) {
	if up.Psw == "" {
		return nil, fmt.Errorf("password is missing")
	}

	pswEncrypted, err := bcrypt.GenerateFromPassword([]byte(up.Psw), 14)
	if err != nil {
		return nil, err
	}

	q := `
	INSERT INTO users(name, email, password_encrypted)
	VALUES(?, ?, ?)
	`
	result, err := pool.ExecContext(ctx, q, up.Name, up.Email, pswEncrypted)
	if err != nil {
		return nil, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return nil, err
	}

	return FindUserById(ctx, pool, int32(id))
}

func (in SignIn) Validate(ctx context.Context, pool *sql.DB) (*User, error) {
	q := `SELECT id, name, email, password_encrypted FROM users WHERE email = ?`

	row := pool.QueryRowContext(ctx, q, in.Email)

	var user User
	err := row.Scan(&user.ID, &user.Name, &user.Email, &user.PswEncrypted)
	if err != nil {
		return nil, err
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.PswEncrypted), []byte(in.Psw))
	if err != nil {
		return nil, fmt.Errorf("password incorrect")
	}

	return &user, nil
}
