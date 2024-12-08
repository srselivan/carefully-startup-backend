package pg

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/jmoiron/sqlx"
	"github.com/samber/lo"
	"investment-game-backend/internal/models"
	"investment-game-backend/internal/repo"
)

type CompanySharesRepo struct {
	db *sqlx.DB
}

func NewCompanySharesRepo(db *sqlx.DB) *CompanySharesRepo {
	return &CompanySharesRepo{db: db}
}

type share struct {
	ID        int64 `db:"id"`
	CompanyID int64 `db:"company_id"`
	Round     int   `db:"round"`
	Price     int64 `db:"price"`
}

const companySharesQueryCreate = `
insert into backend.company_share (company_id, round, price) 
values (:company_id, :round, :price)
`

func (r *CompanySharesRepo) Create(ctx context.Context, share *models.CompanyShare) (int64, error) {
	result, err := r.db.NamedExecContext(
		ctx,
		companySharesQueryCreate,
		struct {
			CompanyID int64 `db:"company_id"`
			Round     int   `db:"round"`
			Price     int64 `db:"price"`
		}{
			CompanyID: share.CompanyID,
			Round:     share.Round,
			Price:     share.Price,
		},
	)
	if err != nil {
		return 0, fmt.Errorf("query error: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("get inserted id: %w", err)
	}
	return id, nil
}

const companySharesQueryUpdate = `
update backend.company_share 
set price = :price
where company_id = :company_id and round = :round
`

func (r *CompanySharesRepo) Update(ctx context.Context, share *models.CompanyShare) error {
	result, err := r.db.NamedExecContext(
		ctx,
		companySharesQueryUpdate,
		struct {
			CompanyID int64 `db:"company_id"`
			Round     int   `db:"round"`
			Price     int64 `db:"price"`
		}{
			CompanyID: share.CompanyID,
			Round:     share.Round,
			Price:     share.Price,
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

const companySharesQueryGetAllActual = `
select 
    cs.id, 
    cs.company_id, 
    cs.round, 
    cs.price
from backend.company_share cs
left join backend.company c on c.id = cs.company_id
where not c.archived or c.archived isnull 
`

func (r *CompanySharesRepo) GetAllActual(ctx context.Context) ([]models.CompanyShare, error) {
	var shares []share
	if err := r.db.SelectContext(ctx, &shares, companySharesQueryGetAllActual); err != nil {
		return nil, fmt.Errorf("query error: %w", err)
	}
	return lo.Map(
		shares,
		func(item share, _ int) models.CompanyShare {
			return models.CompanyShare{
				ID:        item.ID,
				CompanyID: item.CompanyID,
				Round:     item.Round,
				Price:     item.Price,
			}
		},
	), nil
}

const companySharesQueryGetListByIDs = `
select 
    cs.id, 
    cs.company_id, 
    cs.round, 
    cs.price
from backend.company_share cs
where cs.id in (?)
`

func (r *CompanySharesRepo) GetListByIDs(ctx context.Context, ids []int64) ([]models.CompanyShare, error) {
	query, args, err := sqlx.In(companySharesQueryGetListByIDs, ids)
	if err != nil {
		return nil, fmt.Errorf("sqlx.In: %w", err)
	}
	query = r.db.Rebind(query)

	var shares []share
	if err = r.db.SelectContext(ctx, &shares, query, args...); err != nil {
		return nil, fmt.Errorf("query error: %w", err)
	}

	return lo.Map(
		shares,
		func(item share, _ int) models.CompanyShare {
			return models.CompanyShare{
				ID:        item.ID,
				CompanyID: item.CompanyID,
				Round:     item.Round,
				Price:     item.Price,
			}
		},
	), nil
}

const companySharesQueryGetListByCompanyID = `
select 
    cs.id, 
    cs.company_id, 
    cs.round, 
    cs.price
from backend.company_share cs
where cs.company_id = $1
`

func (r *CompanySharesRepo) GetListByCompanyID(ctx context.Context, companyID int64) ([]models.CompanyShare, error) {
	var shares []share
	if err := r.db.SelectContext(ctx, &shares, companySharesQueryGetListByCompanyID, companyID); err != nil {
		return nil, fmt.Errorf("query error: %w", err)
	}
	return lo.Map(
		shares,
		func(item share, _ int) models.CompanyShare {
			return models.CompanyShare{
				ID:        item.ID,
				CompanyID: item.CompanyID,
				Round:     item.Round,
				Price:     item.Price,
			}
		},
	), nil
}

const companySharesQueryGetListByCompanyIDAndRound = `
select 
    cs.id, 
    cs.company_id, 
    cs.round, 
    cs.price
from backend.company_share cs
where cs.company_id in (?) and cs.round = ?
`

func (r *CompanySharesRepo) GetListByCompanyIDsAndRound(
	ctx context.Context,
	companyIDs []int64,
	round int,
) ([]models.CompanyShare, error) {
	query, args, err := sqlx.In(companySharesQueryGetListByIDs, companyIDs)
	if err != nil {
		return nil, fmt.Errorf("sqlx.In: %w", err)
	}
	query = r.db.Rebind(query)
	args = append(args, round)

	var shares []share
	if err = r.db.SelectContext(ctx, &shares, query, args...); err != nil {
		return nil, fmt.Errorf("query error: %w", err)
	}
	return lo.Map(
		shares,
		func(item share, _ int) models.CompanyShare {
			return models.CompanyShare{
				ID:        item.ID,
				CompanyID: item.CompanyID,
				Round:     item.Round,
				Price:     item.Price,
			}
		},
	), nil
}
