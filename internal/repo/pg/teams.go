package pg

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jmoiron/sqlx"
	jsoniter "github.com/json-iterator/go"
	"investment-game-backend/internal/models"
	"investment-game-backend/internal/repo"
	"time"
)

type TeamsRepo struct {
	db *sqlx.DB
}

func NewTeamsRepo(db *sqlx.DB) *TeamsRepo {
	return &TeamsRepo{db: db}
}

type team struct {
	ID              int64                `db:"id"`
	CreatedAt       time.Time            `db:"created_at"`
	UpdatedAt       *time.Time           `db:"updated_at"`
	Name            string               `db:"name"`
	Members         pgtype.Array[string] `db:"members"`
	Credentials     string               `db:"credentials"`
	BalanceID       int64                `db:"balance_id"`
	Shares          []byte               `db:"shares"`
	AdditionalInfos []byte               `db:"additional_info_ids"`
	RandomEventID   *int64               `db:"random_event_id"`
	GameID          int64                `db:"game_id"`
}

const teamsRepoQueryCreate = `
insert into backend.team 
    (
     created_at, 
     updated_at, 
     name, 
     members, 
     credentials, 
     balance_id, 
     shares, 
     additional_info_ids, 
     random_event_id,
     game_id
    ) 
values 
    (
     default, 
     null, 
     :name, 
     :members, 
     :credentials, 
     :balance_id, 
     :shares, 
     :additional_info_ids, 
     :random_event_id,
     :game_id
    )
returning id
`

func (r *TeamsRepo) Create(ctx context.Context, team *models.Team) (int64, error) {
	rows, err := r.db.NamedQueryContext(
		ctx,
		teamsRepoQueryCreate,
		struct {
			Name            string               `db:"name"`
			Members         pgtype.Array[string] `db:"members"`
			Credentials     string               `db:"credentials"`
			BalanceID       int64                `db:"balance_id"`
			Shares          any                  `db:"shares"`
			AdditionalInfos any                  `db:"additional_info_ids"`
			RandomEventID   *int64               `db:"random_event_id"`
			GameID          int64                `db:"game_id"`
		}{
			Name: team.Name,
			Members: pgtype.Array[string]{
				Elements: team.Members,
			},
			Credentials:     team.Credentials,
			BalanceID:       team.BalanceID,
			Shares:          team.Shares,
			AdditionalInfos: team.AdditionalInfos,
			RandomEventID:   team.RandomEventID,
			GameID:          team.GameID,
		},
	)
	if err != nil {
		return 0, fmt.Errorf("exec err: %w", err)
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

const teamsRepoQueryUpdate = `
update backend.team
set (
    updated_at,
    name,
    members,
    shares,
    additional_info_ids,
    random_event_id
) = (
    now(),
    :name,
    :members,
    :shares,
    :additional_info_ids,
    :random_event_id
)
where id = :id
`

func (r *TeamsRepo) Update(ctx context.Context, team *models.Team) error {
	result, err := r.db.NamedExecContext(
		ctx,
		teamsRepoQueryUpdate,
		struct {
			Name            string               `db:"name"`
			Members         pgtype.Array[string] `db:"members"`
			Shares          any                  `db:"shares"`
			AdditionalInfos any                  `db:"additional_infos"`
			RandomEventID   *int64               `db:"random_event_id"`
		}{
			Name: team.Name,
			Members: pgtype.Array[string]{
				Elements: team.Members,
			},
			Shares:          team.Shares,
			AdditionalInfos: team.AdditionalInfos,
			RandomEventID:   team.RandomEventID,
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

const teamsRepoQueryDeleteBulk = `
delete from backend.team
where id in(?)
`

func (r *TeamsRepo) DeleteBulk(ctx context.Context, ids []int64) error {
	if len(ids) == 0 {
		return nil
	}
	query, args, err := sqlx.In(teamsRepoQueryDeleteBulk, ids)
	if err != nil {
		return fmt.Errorf("build query: %w", err)
	}
	query = r.db.Rebind(query)
	if _, err = r.db.ExecContext(ctx, query, args...); err != nil {
		return fmt.Errorf("query error: %w", err)
	}
	return nil
}

const teamsRepoQueryGetByCredentials = `
select 
    id, 
    created_at,
    updated_at,
    name, 
    members,
    credentials,
    balance_id,
    shares, 
    additional_info_ids, 
    random_event_id,
    game_id
from backend.team
where credentials = $1
`

func (r *TeamsRepo) GetByCredentials(ctx context.Context, credentials string) (*models.Team, error) {
	var t team
	if err := r.db.GetContext(ctx, &t, teamsRepoQueryGetByCredentials, credentials); err != nil {
		return nil, fmt.Errorf("query error: %w", err)
	}
	model := &models.Team{
		ID:              t.ID,
		CreatedAt:       t.CreatedAt,
		UpdatedAt:       t.UpdatedAt,
		Name:            t.Name,
		Members:         t.Members.Elements,
		Credentials:     t.Credentials,
		BalanceID:       t.BalanceID,
		Shares:          nil,
		AdditionalInfos: nil,
		RandomEventID:   t.RandomEventID,
		GameID:          t.GameID,
	}
	if len(t.Shares) != 0 {
		if err := jsoniter.Unmarshal(t.Shares, &model.Shares); err != nil {
			return nil, fmt.Errorf("unmarshal json: %T:%w", model.Shares, err)
		}
	}
	if len(t.AdditionalInfos) != 0 {
		if err := jsoniter.Unmarshal(t.AdditionalInfos, &model.AdditionalInfos); err != nil {
			return nil, fmt.Errorf("unmarshal json: %T:%w", model.AdditionalInfos, err)
		}
	}
	return model, nil
}

const teamsRepoQueryGetByID = `
select 
    id, 
    created_at,
    updated_at,
    name, 
    members,
    credentials,
    balance_id,
    shares, 
    additional_info_ids, 
    random_event_id,
    game_id
from backend.team
where id = $1
`

func (r *TeamsRepo) GetByID(ctx context.Context, id int64) (*models.Team, error) {
	var t team
	if err := r.db.GetContext(ctx, &t, teamsRepoQueryGetByID, id); err != nil {
		return nil, fmt.Errorf("query error: %w", err)
	}
	model := &models.Team{
		ID:              t.ID,
		CreatedAt:       t.CreatedAt,
		UpdatedAt:       t.UpdatedAt,
		Name:            t.Name,
		Members:         t.Members.Elements,
		Credentials:     t.Credentials,
		BalanceID:       t.BalanceID,
		Shares:          nil,
		AdditionalInfos: nil,
		RandomEventID:   t.RandomEventID,
		GameID:          t.GameID,
	}
	if len(t.Shares) != 0 {
		if err := jsoniter.Unmarshal(t.Shares, &model.Shares); err != nil {
			return nil, fmt.Errorf("unmarshal json: %T:%w", model.Shares, err)
		}
	}
	if len(t.AdditionalInfos) != 0 {
		if err := jsoniter.Unmarshal(t.AdditionalInfos, &model.AdditionalInfos); err != nil {
			return nil, fmt.Errorf("unmarshal json: %T:%w", model.AdditionalInfos, err)
		}
	}
	return model, nil
}

const teamsRepoQueryGetAllByGameID = `
select 
    id, 
    created_at,
    updated_at,
    name, 
    members,
    credentials,
    balance_id,
    shares, 
    additional_info_ids, 
    random_event_id,
    game_id
from backend.team
where game_id = $1
`

func (r *TeamsRepo) GetAllByGameID(ctx context.Context, gameID int64) ([]models.Team, error) {
	var teams []team
	if err := r.db.SelectContext(ctx, &teams, teamsRepoQueryGetAllByGameID, gameID); err != nil {
		return nil, fmt.Errorf("query error: %w", err)
	}
	result := make([]models.Team, 0, len(teams))
	for _, t := range teams {
		model := models.Team{
			ID:              t.ID,
			CreatedAt:       t.CreatedAt,
			UpdatedAt:       t.UpdatedAt,
			Name:            t.Name,
			Members:         t.Members.Elements,
			Credentials:     t.Credentials,
			BalanceID:       t.BalanceID,
			Shares:          nil,
			AdditionalInfos: nil,
			RandomEventID:   t.RandomEventID,
			GameID:          t.GameID,
		}
		if len(t.Shares) != 0 {
			if err := jsoniter.Unmarshal(t.Shares, &model.Shares); err != nil {
				return nil, fmt.Errorf("unmarshal json: %T:%w", model.Shares, err)
			}
		}
		if len(t.AdditionalInfos) != 0 {
			if err := jsoniter.Unmarshal(t.AdditionalInfos, &model.AdditionalInfos); err != nil {
				return nil, fmt.Errorf("unmarshal json: %T:%w", model.AdditionalInfos, err)
			}
		}
		result = append(result, model)
	}
	return result, nil
}
