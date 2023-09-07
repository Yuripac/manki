package data

import "time"

func CalcCardMemo(c *Card, score int8) {
	CalcCardEfactor(c, score)
	CalcCardNextRep(c)
}

func CalcCardEfactor(c *Card, score int8) {
	if c.Efactor <= 1.3 {
		return
	}

	scoreF := float64(score)
	c.Efactor = c.Efactor - 0.8 + (0.28 * scoreF) - (0.02 * scoreF * scoreF)
}

func CalcCardNextRep(c *Card) {
	newDuration := calcDaysDuration(c.Repetitions, c.Efactor)
	nextRep := time.Now().Add(newDuration)

	c.NextRepetition = &nextRep
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
