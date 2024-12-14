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

type AdditionalInfosRepo struct {
	db *sqlx.DB
}

func NewAdditionalInfosRepo(db *sqlx.DB) *AdditionalInfosRepo {
	return &AdditionalInfosRepo{db: db}
}

type additionalInfo struct {
	ID          int64  `db:"id"`
	Name        string `db:"name"`
	Description string `db:"description"`
	Type        int8   `db:"type"`
	Cost        int64  `db:"cost"`
	CompanyID   *int64 `db:"company_id"`
}

const additionalInfosQueryCreate = `
insert into backend.additional_info (name, description, type, cost, company_id)
values (:name, :description, :type, :cost, :company_id)
returning id
`

func (r *AdditionalInfosRepo) Create(ctx context.Context, info *models.AdditionalInfo) (int64, error) {
	rows, err := r.db.NamedQueryContext(
		ctx,
		additionalInfosQueryCreate,
		struct {
			Name        string `db:"name"`
			Description string `db:"description"`
			Type        int8   `db:"type"`
			Cost        int64  `db:"cost"`
			CompanyID   *int64 `db:"company_id"`
		}{
			Name:        info.Name,
			Description: info.Description,
			Type:        int8(info.Type),
			Cost:        info.Cost,
			CompanyID:   info.CompanyID,
		},
	)
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

const additionalInfosQueryUpdate = `
update backend.additional_info
set (
    name,
    description,
    cost,
    company_id
) = (
    :name,
    :description,
    :cost,
    :company_id
)
where id = :id
`

func (r *AdditionalInfosRepo) Update(ctx context.Context, info *models.AdditionalInfo) error {
	result, err := r.db.NamedExecContext(
		ctx,
		additionalInfosQueryUpdate,
		struct {
			ID          int64  `db:"id"`
			Name        string `db:"name"`
			Description string `db:"description"`
			Cost        int64  `db:"cost"`
			CompanyID   *int64 `db:"company_id"`
		}{
			ID:          info.ID,
			Name:        info.Name,
			Description: info.Description,
			Cost:        info.Cost,
			CompanyID:   info.CompanyID,
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

const additionalInfosQueryGetAllActual = `
select 
    ai.id,
    ai.name,
    ai.description,
    ai.type,
    ai.company_id,
    ai.cost
from backend.additional_info ai
left join backend.company c on c.id = ai.company_id
where ai.type = $1 and (not c.archived or c.archived isnull)
`

func (r *AdditionalInfosRepo) GetAllActualWithType(
	ctx context.Context,
	infoType models.AdditionalInfoType,
) ([]models.AdditionalInfo, error) {
	var infos []additionalInfo
	if err := r.db.SelectContext(ctx, &infos, additionalInfosQueryGetAllActual, infoType); err != nil {
		return nil, fmt.Errorf("query error: %w", err)
	}
	return lo.Map(
		infos,
		func(item additionalInfo, _ int) models.AdditionalInfo {
			return models.AdditionalInfo{
				ID:          item.ID,
				Name:        item.Name,
				Description: item.Description,
				Type:        models.AdditionalInfoType(item.Type),
				Cost:        item.Cost,
				CompanyID:   item.CompanyID,
			}
		},
	), nil
}

const additionalInfosQueryGetByID = `
select 
    ai.id,
    ai.name,
    ai.description,
    ai.type,
    ai.company_id,
    ai.cost
from backend.additional_info ai
where id = $1
`

func (r *AdditionalInfosRepo) GetByID(ctx context.Context, id int64) (*models.AdditionalInfo, error) {
	var info additionalInfo
	if err := r.db.GetContext(ctx, &info, additionalInfosQueryGetByID, id); err != nil {
		return nil, fmt.Errorf("query error: %w", err)
	}
	return &models.AdditionalInfo{
		ID:          info.ID,
		Name:        info.Name,
		Description: info.Description,
		Type:        models.AdditionalInfoType(info.Type),
		Cost:        info.Cost,
		CompanyID:   info.CompanyID,
	}, nil
}

const additionalInfosQueryGetByIDs = `
select 
    ai.id,
    ai.name,
    ai.description,
    ai.type,
    ai.company_id,
    ai.cost
from backend.additional_info ai
where id in (?)
`

func (r *AdditionalInfosRepo) GetByIDs(ctx context.Context, ids []int64) ([]models.AdditionalInfo, error) {
	query, args, err := sqlx.In(additionalInfosQueryGetByIDs, ids)
	if err != nil {
		return nil, fmt.Errorf("sqlx.In: %w", err)
	}
	query = r.db.Rebind(query)

	var infos []additionalInfo
	if err = r.db.SelectContext(ctx, &infos, query, args...); err != nil {
		return nil, fmt.Errorf("query error: %w", err)
	}
	return lo.Map(
		infos,
		func(item additionalInfo, _ int) models.AdditionalInfo {
			return models.AdditionalInfo{
				ID:          item.ID,
				Name:        item.Name,
				Description: item.Description,
				Type:        models.AdditionalInfoType(item.Type),
				Cost:        item.Cost,
				CompanyID:   item.CompanyID,
			}
		},
	), nil
}
