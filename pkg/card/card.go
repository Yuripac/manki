package card

import (
	"context"
	"database/sql"
	"fmt"
	"time"
)

type Card struct {
	Id             int32      `json:"id"`
	UserId         int32      `json:"user_id"`
	Sentence       string     `json:"sentence"`
	Meaning        string     `json:"meaning"`
	Efactor        float64    `json:"efactor"`
	Repetitions    int32      `json:"repetitions"`
	NextRepetition *time.Time `json:"next_repetition_at"`
}

func Next(ctx context.Context, pool *sql.DB) (*Card, error) {
	q := `
	SELECT id, repetitions, efactor, next_repetition_at
	FROM cards
	WHERE next_repetition_at IS NULL 
		OR DATE(next_repetition_at) <= DATE('now')
	LIMIT 1
	`

	var card Card
	err := pool.QueryRowContext(ctx, q).Scan(&card.Id, &card.Repetitions, &card.Efactor, &card.NextRepetition)
	if err != nil {
		fmt.Printf("error searching the next card: %s\n", err)
		return nil, err
	}

	return &card, nil
}

func All(ctx context.Context, pool *sql.DB) ([]Card, error) {
	q := `
	SELECT id, sentence, meaning
	FROM cards
	ORDER BY id DESC;
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

func UpdateMemo(ctx context.Context, pool *sql.DB, c *Card, score float64) error {
	q := `
	UPDATE cards
	SET repetitions = ?, efactor = ?, next_repetition_at = ?
	WHERE id = ?
	`

	c.Repetitions++
	memo := Memo{Card: c}
	c.Efactor = memo.CalcEfactor(score)
	nextRep := memo.NextRepetition()
	c.NextRepetition = &nextRep

	result, err := pool.ExecContext(ctx, q, c.Repetitions, c.Efactor, c.NextRepetition, c.Id)
	if err != nil {
		return err
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return fmt.Errorf("card %d not updated", c.Id)
	}
	return nil
}

func Add(ctx context.Context, pool *sql.DB, c *Card) error {
	if !c.isValid() {
		return fmt.Errorf("card attribute is missing")
	}

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

func (c Card) isValid() bool {
	return c.Sentence != "" && c.Meaning != "" && c.UserId != 0
}
