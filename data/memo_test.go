package data

import (
	"fmt"
	"testing"
	"time"
)

type testMemoCalculation struct {
	Card
	Score              int8
	Repetitions        int32
	Efactor            float64
	NextRepetitionDays time.Duration
}

func TestMemoCalculationSetup(t *testing.T) {
	tests := []testMemoCalculation{
		{
			Card:               Card{Efactor: 2.5},
			Score:              1,
			Repetitions:        1,
			Efactor:            1.96,
			NextRepetitionDays: 1,
		},
		{
			Card:               Card{Efactor: 1.96, Repetitions: 1},
			Score:              1,
			Repetitions:        2,
			Efactor:            1.42,
			NextRepetitionDays: 6,
		},
		{
			Card:               Card{Efactor: 1.42, Repetitions: 2},
			Score:              1,
			Repetitions:        3,
			Efactor:            1.3,
			NextRepetitionDays: 8,
		},
		{
			Card:               Card{Efactor: 1.3, Repetitions: 3},
			Score:              1,
			Repetitions:        4,
			Efactor:            1.3,
			NextRepetitionDays: 10,
		},
		{
			Card:               Card{Efactor: 2.5},
			Score:              3,
			Repetitions:        1,
			Efactor:            2.36,
			NextRepetitionDays: 1,
		},
		{
			Card:               Card{Efactor: 2.36, Repetitions: 1},
			Score:              3,
			Repetitions:        2,
			Efactor:            2.22,
			NextRepetitionDays: 6,
		},
		{
			Card:               Card{Efactor: 2.22, Repetitions: 2},
			Score:              3,
			Repetitions:        3,
			Efactor:            2.08,
			NextRepetitionDays: 12,
		},
		{
			Card:               Card{Efactor: 2.08, Repetitions: 3},
			Score:              3,
			Repetitions:        4,
			Efactor:            1.94,
			NextRepetitionDays: 23,
		},
	}

	for _, test := range tests {
		func(t *testing.T) {
			c := test.Card
			CalcCardMemo(&c, test.Score)

			if c.Repetitions != test.Repetitions {
				t.Fatalf(`Card.Repetitions = %d, expected %d`, c.Repetitions, test.Repetitions)
			}
			if fmt.Sprintf("%.2f", c.Efactor) != fmt.Sprintf("%.2f", test.Efactor) {
				t.Fatalf(`Card.Efactor = %f, expected %f`, c.Efactor, test.Efactor)
			}

			format := "2006-01-02 15:04"
			cardNextRep := c.NextRepetition.Format(format)

			expNextRepetionDays := time.Now().Add(test.NextRepetitionDays * 24 * time.Hour)
			expCardNextRep := expNextRepetionDays.Format(format)

			if cardNextRep != expCardNextRep {
				t.Fatalf(`Card.NextRepetion = %s, expected %s`, cardNextRep, expCardNextRep)
			}
		}(t)
	}
}
