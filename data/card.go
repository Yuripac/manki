package data

import (
	"context"
	"fmt"
	"manki/db"
	"time"
)

type Card struct {
	ID             int32      `json:"id"`
	UserId         int32      `json:"user_id"`
	Sentence       string     `json:"sentence"`
	Meaning        string     `json:"meaning"`
	Efactor        float64    `json:"efactor"`
	Repetitions    int32      `json:"repetitions"`
	NextRepetition *time.Time `json:"next_repetition_at"`
}

func NextCard(ctx context.Context, userId int32) (*Card, error) {
	q := `
	SELECT id, sentence, meaning, repetitions, efactor, next_repetition_at
	FROM cards
	WHERE (next_repetition_at IS NULL OR DATE(next_repetition_at) <= DATE('now'))
		AND user_id = ?
	LIMIT 1
	`

	var card Card
	err := db.Pool().QueryRowContext(ctx, q, userId).Scan(&card.ID, &card.Sentence, &card.Meaning,
		&card.Repetitions, &card.Efactor, &card.NextRepetition)

	if err != nil {
		return nil, err
	}

	return &card, nil
}

func Cards(ctx context.Context, userId int32) ([]Card, error) {
	q := `
	SELECT id, sentence, meaning
	FROM cards
	WHERE user_id = ?
	ORDER BY id DESC;
	`
	rows, err := db.Pool().QueryContext(ctx, q, userId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var cards []Card
	for rows.Next() {
		var card Card
		rows.Scan(&card.ID, &card.Sentence, &card.Meaning)
		cards = append(cards, card)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return cards, nil
}

func UpdateMemo(ctx context.Context, c *Card, score int8) error {
	q := `
	UPDATE cards
	SET repetitions = ?, efactor = ?, next_repetition_at = ?
	WHERE id = ?
	`

	CalcCardMemo(c, score)

	result, err := db.Pool().ExecContext(ctx, q, c.Repetitions, c.Efactor, c.NextRepetition, c.ID)
	if err != nil {
		return err
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return fmt.Errorf("card %d not updated", c.ID)
	}
	return nil
}

func AddCard(ctx context.Context, c *Card) error {
	if !c.isValid() {
		return fmt.Errorf("card attribute is missing")
	}

	q := `
	INSERT INTO cards(sentence, meaning, user_id)
	VALUES(?, ?, ?)
	`
	result, err := db.Pool().ExecContext(ctx, q, c.Sentence, c.Meaning, c.UserId)
	if err != nil {
		return err
	}
	id, err := result.LastInsertId()
	c.ID = int32(id)

	if err != nil {
		return err
	}
	return nil
}

func (c Card) isValid() bool {
	return c.Sentence != "" && c.Meaning != "" && c.UserId != 0
}
