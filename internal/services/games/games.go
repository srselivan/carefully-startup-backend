package games

import (
	"context"
	"fmt"
	"github.com/rs/zerolog"
	"investment-game-backend/internal/models"
	"investment-game-backend/internal/repo"
	"time"
)

type Service struct {
	repo            repo.GamesRepo
	tradeController *TradeController
	gameController  *GameController
	log             *zerolog.Logger
}

func New(
	repo repo.GamesRepo,
	tradeController *TradeController,
	gameController *GameController,
	log *zerolog.Logger,
) *Service {
	return &Service{
		repo:            repo,
		tradeController: tradeController,
		gameController:  gameController,
		log:             log,
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
	TradeState   models.TradeState
}

func (s *Service) Update(ctx context.Context, params UpdateParams) error {
	game, err := s.repo.Get(ctx)
	if err != nil {
		return fmt.Errorf("s.repo.Get: %w", err)
	}

	game.CurrentRound = params.CurrentRound

	tradeStateChanged := game.TradeState != params.TradeState
	game.TradeState = params.TradeState

	gameStateChanged := game.State != params.State
	game.State = params.State
	if gameStateChanged && game.State == models.GameStateStarted {
		game.CurrentRound = 1
	}

	if err = s.repo.Update(ctx, game); err != nil {
		return fmt.Errorf("s.repo.Update: %w", err)
	}

	if gameStateChanged {
		s.onGameStateChange(game.State)
	}

	if tradeStateChanged && game.TradeState == models.TradeStateStarted {
		go func() {
			s.log.Trace().Msg("start trade period")
			s.tradeController.StartTradePeriod()
			s.log.Trace().Msg("stopped trade period")

			game.TradeState = models.TradeStateNotStarted
			if err = s.repo.Update(ctx, game); err != nil {
				s.log.Error().Err(err).Msg("s.repo.Update")
			}
		}()
	}

	return nil
}

func (s *Service) CreateNewGame(ctx context.Context) error {
	s.log.Trace().Msg("create new game")

	game, err := s.repo.Get(ctx)
	if err != nil {
		return fmt.Errorf("s.repo.Get: %w", err)
	}
	game.CurrentGame++
	// Registration Closed (-1)
	game.State = models.GameStateClosed
	s.onGameStateChange(models.GameStateClosed)

	// default
	game.CurrentRound = 0
	game.TradeState = models.TradeStateNotStarted
	s.tradeController.StopTradePeriod()

	if err = s.repo.Update(ctx, game); err != nil {
		return fmt.Errorf("s.repo.Update: %w", err)
	}
	return nil
}

func (s *Service) StartGame(ctx context.Context) error {
	s.log.Trace().Msg("start game")

	game, err := s.repo.Get(ctx)
	if err != nil {
		return fmt.Errorf("s.repo.Get: %w", err)
	}

	game.State = models.GameStateStarted
	s.onGameStateChange(models.GameStateStarted)
	game.CurrentRound = 0
	game.TradeState = models.TradeStateNotStarted
	s.tradeController.StopTradePeriod()

	if err = s.repo.Update(ctx, game); err != nil {
		return fmt.Errorf("s.repo.Update: %w", err)
	}
	return nil
}

func (s *Service) StartRegistration(ctx context.Context) error {
	s.log.Trace().Msg("start registration")

	game, err := s.repo.Get(ctx)
	if err != nil {
		return fmt.Errorf("s.repo.Get: %w", err)
	}

	game.State = models.GameStateOpened
	s.onGameStateChange(models.GameStateOpened)

	if err = s.repo.Update(ctx, game); err != nil {
		return fmt.Errorf("s.repo.Update: %w", err)
	}
	return nil
}

func (s *Service) StopRegistration(ctx context.Context) error {
	s.log.Trace().Msg("stop registration")

	game, err := s.repo.Get(ctx)
	if err != nil {
		return fmt.Errorf("s.repo.Get: %w", err)
	}

	game.State = models.GameStateClosed
	s.onGameStateChange(models.GameStateClosed)

	if err = s.repo.Update(ctx, game); err != nil {
		return fmt.Errorf("s.repo.Update: %w", err)
	}
	return nil
}

func (s *Service) StartRound(ctx context.Context) error {
	s.log.Trace().Msg("start round")

	game, err := s.repo.Get(ctx)
	if err != nil {
		return fmt.Errorf("s.repo.Get: %w", err)
	}
	game.CurrentRound++
	if err = s.repo.Update(ctx, game); err != nil {
		return fmt.Errorf("s.repo.Update: %w", err)
	}
	return nil
}

func (s *Service) StartTrade(ctx context.Context) error {
	game, err := s.repo.Get(ctx)
	if err != nil {
		return fmt.Errorf("s.repo.Get: %w", err)
	}
	game.TradeState = models.TradeStateStarted

	if err = s.repo.Update(ctx, game); err != nil {
		return fmt.Errorf("s.repo.Update: %w", err)
	}

	go func() {
		s.log.Trace().Msg("start trade period")
		s.tradeController.StartTradePeriod()
		s.log.Trace().Msg("stop trade period")
		game.TradeState = models.TradeStateNotStarted
		if err = s.repo.Update(context.Background(), game); err != nil {
			s.log.Error().Err(err).Msg("s.repo.Update")
		}
	}()

	return nil
}

func (s *Service) StopTrade(_ context.Context) {
	s.log.Trace().Msg("called stop trade")
	s.tradeController.StopTradePeriod()
}

func (s *Service) UpdateTradePeriod(period time.Duration) {
	s.log.Trace().Msg("trade period updated")
	s.tradeController.SetPeriod(period)
}

func (s *Service) onGameStateChange(state models.GameState) {
	switch state {
	case models.GameStateOpened:
		s.log.Trace().Msg("start registration period")
		s.gameController.StartRegistrationPeriod()
	case models.GameStateClosed:
		s.log.Trace().Msg("stop registration period")
		s.gameController.StopRegistrationPeriod()
	case models.GameStateStarted:
		s.log.Trace().Msg("stop registration period")
		s.gameController.StopRegistrationPeriod()
	default:
	}
}
