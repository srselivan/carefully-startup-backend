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

type CompaniesRepo struct {
	db *sqlx.DB
}

func NewCompaniesRepo(db *sqlx.DB) *CompaniesRepo {
	return &CompaniesRepo{db: db}
}

type company struct {
	ID       int64  `db:"id"`
	Name     string `db:"name"`
	Archived *bool  `db:"archived"`
}

const companiesRepoQueryCreate = `
insert into backend.company (name, archived) values (:name, :archived)
`

func (r *CompaniesRepo) Create(ctx context.Context, company *models.Company) (int64, error) {
	result, err := r.db.NamedExecContext(
		ctx,
		companiesRepoQueryCreate,
		struct {
			Name     string `db:"name"`
			Archived *bool  `db:"archived"`
		}{
			Name:     company.Name,
			Archived: company.Archived,
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

const companiesRepoQueryUpdate = `
update backend.company
set (
    name, 
    archived
) = (
    :name, 
    :archived
)
where id = :id
`

func (r *CompaniesRepo) Update(ctx context.Context, company *models.Company) error {
	result, err := r.db.NamedExecContext(
		ctx,
		companiesRepoQueryUpdate,
		struct {
			ID       int64  `db:"id"`
			Name     string `db:"name"`
			Archived *bool  `db:"archived"`
		}{
			ID:       company.ID,
			Name:     company.Name,
			Archived: company.Archived,
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

const companiesRepoQueryGetByID = `
select 
    id, 
    name,
    archived
from backend.company
where id = $1
`

func (r *CompaniesRepo) GetByID(ctx context.Context, id int64) (*models.Company, error) {
	var c company
	if err := r.db.GetContext(ctx, &c, companiesRepoQueryGetByID, id); err != nil {
		return nil, fmt.Errorf("query error: %w", err)
	}
	return &models.Company{
		ID:       c.ID,
		Name:     c.Name,
		Archived: c.Archived,
	}, nil
}

const companiesRepoQueryGetAll = `
select 
    id, 
    name,
    archived
from backend.company
where not archived or archived isnull
`

func (r *CompaniesRepo) GetAllNotArchived(ctx context.Context) ([]models.Company, error) {
	var companies []company
	if err := r.db.SelectContext(ctx, &companies, companiesRepoQueryGetAll); err != nil {
		return nil, fmt.Errorf("query error: %w", err)
	}
	return lo.Map(
		companies,
		func(item company, _ int) models.Company {
			return models.Company{
				ID:       item.ID,
				Name:     item.Name,
				Archived: item.Archived,
			}
		},
	), nil
}
