package settings

import (
	"context"
	"fmt"
	"github.com/rs/zerolog"
	"investment-game-backend/internal/models"
	"investment-game-backend/internal/repo"
	"time"
)

type Service struct {
	repo                      repo.SettingsRepo
	updateTradePeriodCallback func(time.Duration)
	teamsRepo                 repo.TeamsRepo
	gameRepo                  repo.GamesRepo
	balanceRepo               repo.BalancesRepo
	additionalInfoRepo        repo.AdditionalInfosRepo
	log                       *zerolog.Logger
}

func New(
	repo repo.SettingsRepo,
	updateTradePeriodCallback func(time.Duration),
	teamsRepo repo.TeamsRepo,
	balanceRepo repo.BalancesRepo,
	gameRepo repo.GamesRepo,
	log *zerolog.Logger,
) *Service {
	return &Service{
		repo:                      repo,
		balanceRepo:               balanceRepo,
		teamsRepo:                 teamsRepo,
		updateTradePeriodCallback: updateTradePeriodCallback,
		gameRepo:                  gameRepo,
		log:                       log,
	}
}

func (s *Service) Get(ctx context.Context) (*models.Settings, error) {
	settings, err := s.repo.Get(ctx)
	if err != nil {
		return nil, fmt.Errorf("s.repo.Get: %w", err)
	}
	return settings, nil
}

type UpdateParams struct {
	RoundsCount               int
	RoundsDuration            time.Duration
	LinkToPDF                 string
	EnableRandomEvents        bool
	DefaultBalance            int64
	DefaultAdditionalInfoCost int64
}

func (s *Service) Update(ctx context.Context, params UpdateParams) error {
	settings, err := s.repo.Get(ctx)
	if err != nil {
		return fmt.Errorf("s.repo.Get: %w", err)
	}

	settings.RoundsCount = params.RoundsCount
	if settings.RoundsDuration != params.RoundsDuration {
		s.updateTradePeriodCallback(params.RoundsDuration)
	}
	settings.RoundsDuration = params.RoundsDuration
	settings.EnableRandomEvents = params.EnableRandomEvents
	settings.LinkToPDF = params.LinkToPDF

	if settings.DefaultBalanceAmount != params.DefaultBalance {
		if err = s.updateDefaultBalanceForActiveTeams(ctx, params.DefaultBalance); err != nil {
			return fmt.Errorf("s.updateDefaultBalanceForActiveTeams: %w", err)
		}
	}
	settings.DefaultBalanceAmount = params.DefaultBalance

	if settings.DefaultAdditionalInfoCost != params.DefaultAdditionalInfoCost {
		if err = s.updateDefaultCostForActiveAddInfos(ctx, params.DefaultAdditionalInfoCost); err != nil {
			return fmt.Errorf("s.updateDefaultCostForActiveAddInfos: %w", err)
		}
	}
	settings.DefaultAdditionalInfoCost = params.DefaultAdditionalInfoCost

	if err = s.repo.Update(ctx, settings); err != nil {
		return fmt.Errorf("s.repo.Update: %w", err)
	}
	return nil
}

func (s *Service) updateDefaultBalanceForActiveTeams(ctx context.Context, defaultBalance int64) error {
	game, err := s.gameRepo.Get(ctx)
	if err != nil {
		return fmt.Errorf("s.gameRepo.Get: %w", err)
	}
	if game.State == models.GameStateStarted {
		return nil
	}

	teams, err := s.teamsRepo.GetAllByGameID(ctx, game.CurrentGame)
	if err != nil {
		return fmt.Errorf("s.teamsRepo.GetAllByGameID: %w", err)
	}
	if len(teams) == 0 {
		return nil
	}

	for _, team := range teams {
		if err = s.balanceRepo.Update(ctx, &models.Balance{
			ID:     team.BalanceID,
			Amount: defaultBalance,
		}); err != nil {
			return fmt.Errorf("s.balanceRepo.Update: %w", err)
		}
	}
	return nil
}

func (s *Service) updateDefaultCostForActiveAddInfos(ctx context.Context, defaultCost int64) error {
	additionalInfos, err := s.additionalInfoRepo.GetAllActualWithType(ctx, models.AdditionalInfoTypeCompanyInfo)
	if err != nil {
		return fmt.Errorf("s.additionalInfoRepo.GetAllActualWithType: %w", err)
	}

	for _, additionalInfo := range additionalInfos {
		additionalInfo.Cost = defaultCost
		if err = s.additionalInfoRepo.Update(ctx, &additionalInfo); err != nil {
			return fmt.Errorf("s.additionalInfoRepo.Update: %w", err)
		}
	}
	return nil
}
