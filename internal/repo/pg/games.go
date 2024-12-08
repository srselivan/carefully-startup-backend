package pg

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/jmoiron/sqlx"
	"investment-game-backend/internal/models"
	"investment-game-backend/internal/repo"
)

type GamesRepo struct {
	db *sqlx.DB
}

func NewGamesRepo(db *sqlx.DB) *GamesRepo {
	return &GamesRepo{db: db}
}

type game struct {
	State        int8 `db:"state"`
	CurrentRound int  `db:"current_round"`
	RoundState   int8 `db:"round_state"`
}

const gamesRepoUpdateQuery = `
update backend.game
set (
    state,
    current_round,
    round_state
) = (
    :state,
    :current_round,
    :round_state
)
where id = 1
`

func (r *GamesRepo) Update(ctx context.Context, game *models.Game) error {
	result, err := r.db.NamedExecContext(
		ctx,
		gamesRepoUpdateQuery,
		struct {
			State        int8 `db:"state"`
			CurrentRound int  `db:"current_round"`
			RoundState   int8 `db:"round_state"`
		}{
			State:        int8(game.State),
			CurrentRound: game.CurrentRound,
			RoundState:   int8(game.RoundState),
		},
	)
	if err != nil {
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

const gamesRepoGetQuery = `
select 
    state,
    current_round,
    round_state
from backend.game
where id = 1
`

func (r *GamesRepo) Get(ctx context.Context) (*models.Game, error) {
	var g game
	if err := r.db.GetContext(ctx, &g, gamesRepoGetQuery); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, repo.ErrNotFound
		}
		return nil, fmt.Errorf("query error: %w", err)
	}
	return &models.Game{
		State:        models.GameState(g.State),
		CurrentRound: g.CurrentRound,
		RoundState:   models.RoundState(g.RoundState),
	}, nil
}
