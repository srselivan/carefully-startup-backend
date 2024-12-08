package models

import (
	"errors"
	"time"
)

// TeamSharesState предоставляет информацию о текущих акциях, которыми владеет команда (Team).
// Ключ - ID компании (Company.ID), значение - количество акций.
type TeamSharesState map[int64]int64

type Team struct {
	ID              int64
	CreatedAt       time.Time
	UpdatedAt       *time.Time
	Name            string
	Members         []string
	Credentials     string
	BalanceID       int64
	Shares          TeamSharesState
	AdditionalInfos []int64
	RandomEventID   *int64
	GameID          int64
}

func (shares TeamSharesState) MergeChanges(changes map[int64]int64) error {
	for share, count := range changes {
		current, ok := shares[share]
		if !ok {
			if count < 0 {
				return errors.New("count of shares cannot be negative")
			}
			shares[share] = count
			continue
		}
		if current+count < 0 {
			return errors.New("count of shares cannot be negative")
		}
		shares[share] = current + count
	}
	return nil
}
