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
	repo repo.SettingsRepo
	log  *zerolog.Logger
}

func New(
	repo repo.SettingsRepo,
	log *zerolog.Logger,
) *Service {
	return &Service{
		repo: repo,
		log:  log,
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
	RoundsCount        int
	RoundsDuration     time.Duration
	LinkToPDF          string
	EnableRandomEvents bool
}

func (s *Service) Update(ctx context.Context, params UpdateParams) error {
	settings, err := s.repo.Get(ctx)
	if err != nil {
		return fmt.Errorf("s.repo.Get: %w", err)
	}

	settings.RoundsCount = params.RoundsCount
	settings.RoundsDuration = params.RoundsDuration
	settings.EnableRandomEvents = params.EnableRandomEvents
	settings.LinkToPDF = params.LinkToPDF

	if err = s.repo.Update(ctx, settings); err != nil {
		return fmt.Errorf("s.repo.Update: %w", err)
	}
	return nil
}
