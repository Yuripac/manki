package user

import (
	"context"
	"database/sql"
	"fmt"
	"os"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

type User struct {
	Id           int32
	Name         string `json:"name"`
	Email        string `json:"email"`
	PswEncrypted string
	JWT          string `json:"jwt"`
}

type SignUp struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (signup SignUp) Save(ctx context.Context, pool *sql.DB) (*User, error) {
	if signup.Password == "" {
		return nil, fmt.Errorf("password is missing")
	}

	pswEncrypted, err := bcrypt.GenerateFromPassword([]byte(signup.Password), 14)
	if err != nil {
		return nil, err
	}

	// Save User and returns
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

func SignIn(ctx context.Context, pool *sql.DB, email, password string) (*User, error) {
	q := `SELECT id, name, email, password_encrypted FROM users WHERE email = ?`

	row := pool.QueryRowContext(ctx, q, email)

	var user User
	err := row.Scan(&user.Id, &user.Name, &user.Email, &user.PswEncrypted)
	if err != nil {
		return nil, err
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.PswEncrypted), []byte(password))
	if err != nil {
		return nil, fmt.Errorf("password incorrect")
	}

	return &user, nil
}

func FindByJWT(ctx context.Context, pool *sql.DB, tokenStr string) (*User, error) {
	token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
		}

		return []byte(os.Getenv("JWT_SECRET")), nil
	})

	if !token.Valid {
		return nil, fmt.Errorf("JWT is not valid")
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok {
		if id, ok := claims["id"].(float64); ok {
			return FindById(ctx, pool, int32(id))
		} else {
			return nil, fmt.Errorf("User ID was not found in JWT")
		}
	}

	return nil, err
}

func (u User) GenJWT() (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"id":    u.Id,
		"name":  u.Name,
		"email": u.Email,
	})
	tokenStr, err := token.SignedString([]byte(os.Getenv("JWT_SECRET")))
	if err != nil {
		return "", err
	}
	return tokenStr, nil
}

func FindById(ctx context.Context, pool *sql.DB, id int32) (*User, error) {
	q := `SELECT id, name, email FROM users WHERE id = ?`

	row := pool.QueryRowContext(ctx, q, id)

	var user User
	switch err := row.Scan(&user.Id, &user.Name, &user.Email); err {
	case sql.ErrNoRows:
		return nil, err
	case nil:
		return &user, nil
	default:
		return nil, err
	}
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
