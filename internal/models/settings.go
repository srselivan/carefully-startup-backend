package models

import "time"

const DefaultRoundsCount = 3

type Settings struct {
	RoundsCount          int
	RoundsDuration       time.Duration
	LinkToPDF            string
	EnableRandomEvents   bool
	DefaultBalanceAmount int64
}
