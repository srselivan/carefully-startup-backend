package pg

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/jmoiron/sqlx"
	"investment-game-backend/internal/models"
	"investment-game-backend/internal/repo"
	"time"
)

type SettingsRepo struct {
	db *sqlx.DB
}

func NewSettingsRepo(db *sqlx.DB) *SettingsRepo {
	return &SettingsRepo{db: db}
}

type settings struct {
	RoundsCount               int    `db:"rounds_count"`
	RoundsDuration            int64  `db:"rounds_duration"`
	LinkToPDF                 string `db:"link_to_pdf"`
	EnableRandomEvents        bool   `db:"enable_random_events"`
	DefaultBalanceAmount      int64  `db:"default_balance_amount"`
	DefaultAdditionalInfoCost int64  `db:"default_additional_info_cost"`
}

const settingsRepoUpdateQuery = `
update backend.settings
set (
    rounds_count,
    rounds_duration,
    link_to_pdf,
    enable_random_events,
    default_balance_amount,
    default_additional_info_cost
) = (
    :rounds_count,
    :rounds_duration,
    :link_to_pdf,
    :enable_random_events,
    :default_balance_amount,
    :default_additional_info_cost
)
where id = 1
`

func (r *SettingsRepo) Update(ctx context.Context, settings *models.Settings) error {
	result, err := r.db.NamedExecContext(
		ctx,
		settingsRepoUpdateQuery,
		struct {
			RoundsCount               int    `db:"rounds_count"`
			RoundsDuration            int64  `db:"rounds_duration"`
			LinkToPDF                 string `db:"link_to_pdf"`
			EnableRandomEvents        bool   `db:"enable_random_events"`
			DefaultBalanceAmount      int64  `db:"default_balance_amount"`
			DefaultAdditionalInfoCost int64  `db:"default_additional_info_cost"`
		}{
			RoundsCount:               settings.RoundsCount,
			RoundsDuration:            int64(settings.RoundsDuration),
			LinkToPDF:                 settings.LinkToPDF,
			EnableRandomEvents:        settings.EnableRandomEvents,
			DefaultBalanceAmount:      settings.DefaultBalanceAmount,
			DefaultAdditionalInfoCost: settings.DefaultAdditionalInfoCost,
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

const settingsRepoGetQuery = `
select 
    rounds_count,
    rounds_duration,
    link_to_pdf,
    enable_random_events,
    default_balance_amount,
    default_additional_info_cost
from backend.settings
where id = 1
`

func (r *SettingsRepo) Get(ctx context.Context) (*models.Settings, error) {
	var s settings
	if err := r.db.GetContext(ctx, &s, settingsRepoGetQuery); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, repo.ErrNotFound
		}
		return nil, fmt.Errorf("query error: %w", err)
	}
	return &models.Settings{
		RoundsCount:               s.RoundsCount,
		RoundsDuration:            time.Duration(s.RoundsDuration),
		LinkToPDF:                 s.LinkToPDF,
		EnableRandomEvents:        s.EnableRandomEvents,
		DefaultBalanceAmount:      s.DefaultBalanceAmount,
		DefaultAdditionalInfoCost: s.DefaultAdditionalInfoCost,
	}, nil
}
