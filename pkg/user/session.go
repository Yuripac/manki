package user

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

func (signup SignUp) Save(ctx context.Context, pool *sql.DB) (*User, error) {
	if signup.Psw == "" {
		return nil, fmt.Errorf("password is missing")
	}

	pswEncrypted, err := bcrypt.GenerateFromPassword([]byte(signup.Psw), 14)
	if err != nil {
		return nil, err
	}

	q := `
	INSERT INTO users(name, email, password_encrypted)
	VALUES(?, ?, ?)
	`
	result, err := pool.ExecContext(ctx, q, signup.Name, signup.Email, pswEncrypted)
	if err != nil {
		return nil, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return nil, err
	}

	return FindById(ctx, pool, int32(id))
}

func (signin SignIn) Validate(ctx context.Context, pool *sql.DB) (*User, error) {
	q := `SELECT id, name, email, password_encrypted FROM users WHERE email = ?`

	row := pool.QueryRowContext(ctx, q, signin.Email)

	var user User
	err := row.Scan(&user.Id, &user.Name, &user.Email, &user.PswEncrypted)
	if err != nil {
		return nil, err
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.PswEncrypted), []byte(signin.Psw))
	if err != nil {
		return nil, fmt.Errorf("password incorrect")
	}

	return &user, nil
}

// func SignIn(ctx context.Context, pool *sql.DB, email, password string) (*User, error) {
// 	q := `SELECT id, name, email, password_encrypted FROM users WHERE email = ?`

// 	row := pool.QueryRowContext(ctx, q, email)

// 	var user User
// 	err := row.Scan(&user.Id, &user.Name, &user.Email, &user.PswEncrypted)
// 	if err != nil {
// 		return nil, err
// 	}

// 	err = bcrypt.CompareHashAndPassword([]byte(user.PswEncrypted), []byte(password))
// 	if err != nil {
// 		return nil, fmt.Errorf("password incorrect")
// 	}

// 	return &user, nil
// }
