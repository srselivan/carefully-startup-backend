package pg

import (
	"context"
	"fmt"
	"github.com/jmoiron/sqlx"
	"investment-game-backend/internal/models"
)

type BalancesRepo struct {
	db *sqlx.DB
}

func NewBalancesRepo(db *sqlx.DB) *BalancesRepo {
	return &BalancesRepo{db: db}
}

type balance struct {
	ID     int64 `db:"id"`
	Amount int64 `db:"amount"`
}

const balancesQueryCreate = `
insert into backend.balance(amount) 
values($1)
`

func (r *BalancesRepo) Create(ctx context.Context, balance *models.Balance) (int64, error) {
	result, err := r.db.ExecContext(ctx, balancesQueryCreate, balance.Amount)
	if err != nil {
		return 0, fmt.Errorf("query error: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("get inserted id: %w", err)
	}
	return id, nil
}

const balancesQueryUpdate = `
update backend.balance 
set amount = $1 
where id = $2
`

func (r *BalancesRepo) Update(ctx context.Context, balance *models.Balance) error {
	if _, err := r.db.ExecContext(ctx, balancesQueryUpdate, balance.Amount); err != nil {
		return fmt.Errorf("query error: %w", err)
	}
	return nil
}

const balancesQueryGet = `
select id, amount 
from backend.balance 
where id = $1
`

func (r *BalancesRepo) GetByID(ctx context.Context, id int64) (*models.Balance, error) {
	var b balance
	if err := r.db.GetContext(ctx, &b, balancesQueryGet, id); err != nil {
		return nil, fmt.Errorf("query error: %w", err)
	}
	return &models.Balance{
		ID:     b.ID,
		Amount: b.Amount,
	}, nil
}
