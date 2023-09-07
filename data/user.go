package data

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type User struct {
	ID           int32
	Name         string `json:"name"`
	Email        string `json:"email"`
	PswEncrypted string
	JWT          string `json:"jwt"`
}

const jwtDuration = 15 * 24 * time.Hour

func (u User) GenJWT() (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"id":         u.ID,
		"name":       u.Name,
		"email":      u.Email,
		"expires_at": time.Now().Add(jwtDuration).Unix(),
	})
	tokenStr, err := token.SignedString([]byte(os.Getenv("JWT_SECRET")))
	if err != nil {
		return "", err
	}
	return tokenStr, nil
}

func FindUserByJWT(ctx context.Context, pool *sql.DB, tokenStr string) (*User, error) {
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
		expiresAt, _ := claims["expires_at"].(float64)

		if time.Now().Unix() > int64(expiresAt) {
			return nil, fmt.Errorf("token is expired")
		}

		if id, ok := claims["id"].(float64); ok {
			return FindUserById(ctx, pool, int32(id))
		} else {
			return nil, fmt.Errorf("User ID was not found in JWT")
		}
	}

	return nil, err
}

func FindUserById(ctx context.Context, pool *sql.DB, id int32) (*User, error) {
	q := `SELECT id, name, email FROM users WHERE id = ?`

	row := pool.QueryRowContext(ctx, q, id)

	var user User
	switch err := row.Scan(&user.ID, &user.Name, &user.Email); err {
	case sql.ErrNoRows:
		return nil, err
	case nil:
		return &user, nil
	default:
		return nil, err
	}
}

func UserExists(ctx context.Context, pool *sql.DB, id int32) (bool, error) {
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
