package repo

import (
	"context"
	"investment-game-backend/internal/models"
)

type SettingsRepo interface {
	Update(ctx context.Context, settings *models.Settings) error
	Get(ctx context.Context) (*models.Settings, error)
}

type GamesRepo interface {
	Update(ctx context.Context, settings *models.Game) error
	Get(ctx context.Context) (*models.Game, error)
}

type CompaniesRepo interface {
	Update(ctx context.Context, company *models.Company) error
	Create(ctx context.Context, company *models.Company) (int64, error)
	GetByID(ctx context.Context, id int64) (*models.Company, error)
	GetAllNotArchived(ctx context.Context) ([]models.Company, error)
}

type CompanySharesRepo interface {
	Create(ctx context.Context, share *models.CompanyShare) (int64, error)
	Update(ctx context.Context, share *models.CompanyShare) error
	GetAllActual(ctx context.Context) ([]models.CompanyShare, error)
	GetListByIDs(ctx context.Context, ids []int64) ([]models.CompanyShare, error)
	GetListByCompanyID(ctx context.Context, companyID int64) ([]models.CompanyShare, error)
	GetListByCompanyIDsAndRound(ctx context.Context, companyIDs []int64, round int) ([]models.CompanyShare, error)
}

type AdditionalInfosRepo interface {
	Create(ctx context.Context, info *models.AdditionalInfo) (int64, error)
	Update(ctx context.Context, info *models.AdditionalInfo) error
	GetAllActualWithType(ctx context.Context, infoType models.AdditionalInfoType) ([]models.AdditionalInfo, error)
	GetByID(ctx context.Context, id int64) (*models.AdditionalInfo, error)
	GetByIDs(ctx context.Context, ids []int64) ([]models.AdditionalInfo, error)
}

type BalancesRepo interface {
	Create(ctx context.Context, balance *models.Balance) (int64, error)
	Update(ctx context.Context, balance *models.Balance) error
	GetByID(ctx context.Context, id int64) (*models.Balance, error)
}

type TeamsRepo interface {
	Create(ctx context.Context, team *models.Team) (int64, error)
	Update(ctx context.Context, team *models.Team) error
	DeleteBulk(ctx context.Context, ids []int64) error
	GetByCredentials(ctx context.Context, credentials string, gameID int64) (*models.Team, error)
	GetByID(ctx context.Context, id int64) (*models.Team, error)
	GetAllByGameID(ctx context.Context, gameID int64) ([]models.Team, error)
}

type BalanceTransactionsRepo interface {
	Create(ctx context.Context, tr *models.BalanceTransaction) (int64, error)
	Update(ctx context.Context, tr *models.BalanceTransaction) error
	Delete(ctx context.Context, balanceID int64, round int) error
	Get(ctx context.Context, balanceID int64, round int) (*models.BalanceTransaction, error)
}

type AuthRepo interface {
	SetRefreshToken(ctx context.Context, teamID int64, token string) error
	VerifyRefreshToken(ctx context.Context, userID int64, token string) (bool, error)
}
