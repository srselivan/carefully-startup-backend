package services

import (
	"context"
	"investment-game-backend/internal/models"
	additionalinfos "investment-game-backend/internal/services/additional_infos"
	"investment-game-backend/internal/services/companies"
	"investment-game-backend/internal/services/games"
	"investment-game-backend/internal/services/settings"
	"investment-game-backend/internal/services/teams"
	"time"
)

type Settings interface {
	Get(ctx context.Context) (*models.Settings, error)
	Update(ctx context.Context, params settings.UpdateParams) error
}

type Games interface {
	Get(ctx context.Context) (*models.Game, error)
	Update(ctx context.Context, params games.UpdateParams) error
	UpdateTradePeriod(period time.Duration)
	CreateNewGame(ctx context.Context) error
	StartGame(ctx context.Context) error
	StopGame(ctx context.Context) error
	StartRegistration(ctx context.Context) error
	StopRegistration(ctx context.Context) error
	StartRound(ctx context.Context) error
	StartTrade(ctx context.Context) error
	StopTrade(_ context.Context)
}

type Companies interface {
	CreateWithShares(ctx context.Context, params companies.CreateWithSharesParams) (int64, error)
	Update(ctx context.Context, params companies.UpdateParams) error
	Archive(ctx context.Context, id int64) error
	GetAllWithShares(ctx context.Context, onlyCurrentRound bool) ([]models.CompanyWithShares, error)
}

type Teams interface {
	Create(ctx context.Context, params teams.CreateParams) (int64, error)
	Update(ctx context.Context, params teams.UpdateParams) error
	Purchase(ctx context.Context, params teams.PurchaseParams) error
	GetDetailedByID(ctx context.Context, id int64) (teams.DetailedTeam, error)
	NotifyTradePeriodUpdated(isTrade bool)
	NotifyGameRegistrationPeriodUpdated(idRegistration bool)
}

type Auth interface {
	Login(ctx context.Context, credentials string) (models.JWTPair, error)
	Refresh(ctx context.Context, refreshToken string) (models.JWTPair, error)
	RefreshTokenExpTime() time.Duration
}

type AdditionalInfos interface {
	Create(ctx context.Context, params additionalinfos.CreateParams) (*models.AdditionalInfo, error)
	Update(ctx context.Context, params additionalinfos.UpdateParams) error
	GetActualListByType(ctx context.Context, infoType models.AdditionalInfoType) ([]models.AdditionalInfo, error)
}
