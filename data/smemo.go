package data

import (
	"math"
	"time"
)

type SMemo struct{}

func (sm SMemo) PrepareNext(c *Card, score int8) {
	c.Repetitions++

	prepareCardEfactor(c, score)
	prepareCardNextRep(c)
}

func prepareCardEfactor(c *Card, score int8) {
	scoreF := float64(score)
	c.Efactor = c.Efactor - 0.8 + (0.28 * scoreF) - (0.02 * scoreF * scoreF)

	if c.Efactor < 1.3 {
		c.Efactor = 1.3
		return
	}
}

func prepareCardNextRep(c *Card) {
	days := calcDays(c.Repetitions, c.Efactor)
	days = math.Round(days)
	nextRep := time.Now().Add(time.Duration(days) * 24 * time.Hour)

	c.NextRepetition = &nextRep
}

func calcDays(repetition int32, efactor float64) (days float64) {
	switch repetition {
	case 1:
		days = 1
	case 2:
		days = 6
	default:
		return calcDays(repetition-1, efactor) * efactor
	}

	return days
}
