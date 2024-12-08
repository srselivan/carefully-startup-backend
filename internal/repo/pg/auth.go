package pg

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/jmoiron/sqlx"
	"investment-game-backend/internal/repo"
)

type AuthRepo struct {
	db *sqlx.DB
}

func NewAuthRepo(db *sqlx.DB) *AuthRepo {
	return &AuthRepo{db: db}
}

const authQuerySetRefreshToken = `
insert into backend.team_refresh_token(team_id, token)
values ($1, $2)
on conflict (team_id) do update set token = $2
`

func (r *AuthRepo) SetRefreshToken(ctx context.Context, teamID int64, token string) error {
	if _, err := r.db.ExecContext(ctx, authQuerySetRefreshToken, teamID, token); err != nil {
		return fmt.Errorf("exec error: %w", err)
	}
	return nil
}

const authQueryVerifyRefreshToken = `
select 1
from backend.team_refresh_token
where team_id = $1 and token = $2
`

func (r *AuthRepo) VerifyRefreshToken(ctx context.Context, userID int64, token string) (bool, error) {
	var ok int8
	if err := r.db.GetContext(ctx, &ok, authQueryVerifyRefreshToken, userID, token); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, repo.ErrNotFound
		}
		return false, fmt.Errorf("query error: %w", err)
	}
	return true, nil
}
