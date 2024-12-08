package pg

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/jmoiron/sqlx"
	jsoniter "github.com/json-iterator/go"
	"investment-game-backend/internal/models"
	"investment-game-backend/internal/repo"
)

type BalanceTransactionsRepo struct {
	db *sqlx.DB
}

func NewBalanceTransactionsRepo(db *sqlx.DB) *BalanceTransactionsRepo {
	return &BalanceTransactionsRepo{db: db}
}

type balanceTransaction struct {
	ID               int64  `db:"id"`
	BalanceID        int64  `db:"balance_id"`
	Round            int    `db:"round"`
	Amount           int64  `db:"amount"`
	Details          []byte `db:"details"`
	AdditionalInfoID *int64 `db:"additional_info_id"`
	RandomEventID    *int64 `db:"random_event_id"`
}

const balanceTransactionsQueryCreate = `
insert into backend.balance_transaction (balance_id, round, amount, details, additional_info_id, random_event_id) 
values (:balance_id, :round, :amount, :details, :additional_info_id, :random_event_id)
`

func (r *BalanceTransactionsRepo) Create(ctx context.Context, tr *models.BalanceTransaction) (int64, error) {
	result, err := r.db.NamedExecContext(
		ctx,
		balanceTransactionsQueryCreate,
		struct {
			BalanceID        int64  `db:"balance_id"`
			Round            int    `db:"round"`
			Amount           int64  `db:"amount"`
			Details          any    `db:"details"`
			AdditionalInfoID *int64 `db:"additional_info_id"`
			RandomEventID    *int64 `db:"random_event_id"`
		}{
			BalanceID:        tr.BalanceID,
			Round:            tr.Round,
			Amount:           tr.Amount,
			Details:          tr.Details,
			AdditionalInfoID: tr.AdditionalInfoID,
			RandomEventID:    tr.RandomEventID,
		},
	)
	if err != nil {
		return 0, fmt.Errorf("exec error: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("get inserted id: %w", err)
	}
	return id, nil
}

const balanceTransactionsQueryUpdate = `
update backend.balance_transaction
set (
     amount, 
     details, 
     additional_info_id, 
     random_event_id
) = (
     :amount, 
     :details, 
     :additional_info_id, 
     :random_event_id
)
where balance_id = :balance_id and round = :round and (additional_info_id isnull and random_event_id isnull)
`

func (r *BalanceTransactionsRepo) Update(ctx context.Context, tr *models.BalanceTransaction) error {
	result, err := r.db.NamedExecContext(
		ctx,
		balanceTransactionsQueryUpdate,
		struct {
			BalanceID        int64  `db:"balance_id"`
			Round            int    `db:"round"`
			Amount           int64  `db:"amount"`
			Details          any    `db:"details"`
			AdditionalInfoID *int64 `db:"additional_info_id"`
			RandomEventID    *int64 `db:"random_event_id"`
		}{
			BalanceID:        tr.BalanceID,
			Round:            tr.Round,
			Amount:           tr.Amount,
			Details:          tr.Details,
			AdditionalInfoID: tr.AdditionalInfoID,
			RandomEventID:    tr.RandomEventID,
		},
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return repo.ErrNotFound
		}
		return fmt.Errorf("query error: %w", err)
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("get affected rows: %w", err)
	}
	if affected == 0 {
		return repo.ErrNothingUpdated
	}
	return nil
}

const balanceTransactionsQueryGet = `
select
    id, 
    balance_id, 
    round, 
    amount, 
    details, 
    additional_info_id, 
    random_event_id
from backend.balance_transaction
where balance_id = $1 and round = $2 and (additional_info_id isnull and random_event_id isnull)
`

func (r *BalanceTransactionsRepo) Get(
	ctx context.Context,
	balanceID int64,
	round int,
) (*models.BalanceTransaction, error) {
	var tr balanceTransaction
	if err := r.db.GetContext(ctx, &tr, balanceTransactionsQueryGet, balanceID, round); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, repo.ErrNotFound
		}
		return nil, fmt.Errorf("query error: %w", err)
	}
	model := &models.BalanceTransaction{
		ID:               tr.ID,
		BalanceID:        tr.BalanceID,
		Round:            tr.Round,
		Amount:           tr.Amount,
		Details:          nil,
		AdditionalInfoID: tr.AdditionalInfoID,
		RandomEventID:    tr.RandomEventID,
	}
	if len(tr.Details) != 0 {
		if err := jsoniter.Unmarshal(tr.Details, &model.Details); err != nil {
			return nil, fmt.Errorf("unmarshal json: %T:%w", model.Details, err)
		}
	}
	return model, nil
}
