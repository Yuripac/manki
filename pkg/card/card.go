package card

import (
	"context"
	"database/sql"
)

type Card struct {
	Id       int32
	UserId   int32
	Sentence string
	Meaning  string
}

func All(ctx context.Context, pool *sql.DB) ([]Card, error) {
	q := `
	SELECT id, sentence, meaning
	FROM cards;
	`
	rows, err := pool.QueryContext(ctx, q)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var cards []Card
	for rows.Next() {
		var card Card
		rows.Scan(&card.Id, &card.Sentence, &card.Meaning)
		cards = append(cards, card)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return cards, nil
}

func Save(ctx context.Context, pool *sql.DB, c *Card) error {
	q := `
	INSERT INTO cards(sentence, meaning, user_id)
	VALUES(?, ?, ?)
	`
	result, err := pool.ExecContext(ctx, q, c.Sentence, c.Meaning, c.UserId)
	if err != nil {
		return err
	}
	id, err := result.LastInsertId()
	c.Id = int32(id)

	if err != nil {
		return err
	}
	return nil
}
