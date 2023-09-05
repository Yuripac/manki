package data

import "time"

type Memo struct {
	*Card
}

func (memo Memo) CalcEfactor(score float64) float64 {
	efactor := memo.Card.Efactor
	if efactor <= 1.3 {
		return 1.3
	}

	return efactor - 0.8 + (0.28 * score) - (0.02 * score * score)
}

func (memo Memo) NextRepetition() time.Time {
	card := memo.Card
	return time.Now().Add(calcDaysDuration(card.Repetitions, card.Efactor))
}

func calcDaysDuration(repetition int32, efactor float64) time.Duration {
	var days time.Duration
	switch repetition {
	case 1:
		days = 1
	case 2:
		days = 6
	default:
		days = calcDaysDuration(repetition-1, efactor) * time.Duration(efactor)
	}

	return days * 24 * time.Hour
}
