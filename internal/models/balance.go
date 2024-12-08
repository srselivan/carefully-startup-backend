package models

type Balance struct {
	ID     int64
	Amount int64
}

type BalanceTransaction struct {
	ID               int64
	BalanceID        int64
	Round            int
	Amount           int64
	Details          map[int64]int64
	AdditionalInfoID *int64
	RandomEventID    *int64
}
