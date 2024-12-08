package games

import (
	"context"
	"fmt"
	"github.com/rs/zerolog"
	"investment-game-backend/internal/models"
	"investment-game-backend/internal/repo"
)

type Service struct {
	repo repo.GamesRepo
	log  *zerolog.Logger
}

func New(
	repo repo.GamesRepo,
	log *zerolog.Logger,
) *Service {
	return &Service{
		repo: repo,
		log:  log,
	}
}

func (s *Service) Get(ctx context.Context) (*models.Game, error) {
	game, err := s.repo.Get(ctx)
	if err != nil {
		return nil, fmt.Errorf("s.repo.Get: %w", err)
	}
	return game, nil
}

type UpdateParams struct {
	State        models.GameState
	CurrentRound int
	RoundState   models.RoundState
}

func (s *Service) Update(ctx context.Context, params UpdateParams) error {
	game, err := s.repo.Get(ctx)
	if err != nil {
		return fmt.Errorf("s.repo.Get: %w", err)
	}

	game.State = params.State
	game.CurrentRound = params.CurrentRound
	game.RoundState = params.RoundState

	if err = s.repo.Update(ctx, game); err != nil {
		return fmt.Errorf("s.repo.Update: %w", err)
	}
	return nil
}
