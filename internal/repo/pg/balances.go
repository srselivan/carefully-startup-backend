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
returning id
`

func (r *BalancesRepo) Create(ctx context.Context, balance *models.Balance) (int64, error) {
	rows, err := r.db.QueryxContext(ctx, balancesQueryCreate, balance.Amount)
	if err != nil {
		return 0, fmt.Errorf("query error: %w", err)
	}
	defer rows.Close()

	var id int64
	for rows.Next() {
		if err = rows.Scan(&id); err != nil {
			return 0, fmt.Errorf("scan error: %w", err)
		}
	}
	if err = rows.Err(); err != nil {
		return 0, fmt.Errorf("rows error: %w", err)
	}

	return id, nil
}

const balancesQueryUpdate = `
update backend.balance 
set amount = $1 
where id = $2
`

func (r *BalancesRepo) Update(ctx context.Context, balance *models.Balance) error {
	if _, err := r.db.ExecContext(ctx, balancesQueryUpdate, balance.Amount, balance.ID); err != nil {
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
