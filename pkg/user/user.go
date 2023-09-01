package user

import (
	"context"
	"database/sql"
	"fmt"
	"os"

	"github.com/golang-jwt/jwt/v5"
)

type User struct {
	Id    int32  `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
	JWT   string `json:"jwt"`
}

func New(name, email, psw string) User {
	return User{Name: name, Email: email}
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
