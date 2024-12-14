package auth

import (
	"context"
	"errors"
	"fmt"
	"github.com/golang-jwt/jwt/v5"
	"github.com/rs/zerolog"
	"investment-game-backend/internal/models"
	"investment-game-backend/internal/repo"
	"time"
)

const administratorTeamID int64 = -1

type Service struct {
	teamsRepo        repo.TeamsRepo
	authRepo         repo.AuthRepo
	gamesRepo        repo.GamesRepo
	jwtConfig        JWTConfig
	adminCredentials string
	log              *zerolog.Logger
}

func New(
	teamsRepo repo.TeamsRepo,
	authRepo repo.AuthRepo,
	gamesRepo repo.GamesRepo,
	jwtConfig JWTConfig,
	adminCredentials AdminCredentials,
	log *zerolog.Logger,
) *Service {
	return &Service{
		teamsRepo:        teamsRepo,
		authRepo:         authRepo,
		gamesRepo:        gamesRepo,
		jwtConfig:        jwtConfig,
		adminCredentials: adminCredentials.String(),
		log:              log,
	}
}

func (s *Service) Login(ctx context.Context, credentials string, isAdmin bool) (models.JWTPair, error) {
	if isAdmin {
		if s.adminCredentials != credentials {
			return models.JWTPair{}, errors.New("unsuccessful login")
		}
	}

	game, err := s.gamesRepo.Get(ctx)
	if err != nil {
		return models.JWTPair{}, fmt.Errorf("failed to get game: %w", err)
	}
	var (
		additionalClaims map[string]any
		teamID           int64
	)
	if s.adminCredentials != credentials {
		team, err := s.teamsRepo.GetByCredentials(ctx, credentials, game.CurrentGame)
		if err != nil {
			if errors.Is(err, repo.ErrNotFound) {
				return models.JWTPair{}, errors.New("unsuccessful login")
			}
			return models.JWTPair{}, fmt.Errorf("s.teamsRepo.GetByCredentials: %w", err)
		}
		additionalClaims = s.getClaimsForTeam(ctx, team)
		teamID = team.ID
	} else {
		additionalClaims = s.getClaimsForAdmin(ctx)
		teamID = administratorTeamID
	}

	jwtPair, err := s.getJWTPair(ctx, additionalClaims)
	if err != nil {
		return models.JWTPair{}, fmt.Errorf("s.getJWTPair: %w", err)
	}

	if err = s.authRepo.SetRefreshToken(ctx, teamID, jwtPair.RefreshToken); err != nil {
		return models.JWTPair{}, fmt.Errorf("s.repo.SetRefreshToken: %w", err)
	}

	return jwtPair, nil
}

func (s *Service) Refresh(ctx context.Context, refreshToken string) (models.JWTPair, error) {
	token, err := jwt.Parse(refreshToken, func(token *jwt.Token) (interface{}, error) {
		return []byte(s.jwtConfig.JWTRefreshSecretKey), nil
	})
	if err != nil {
		return models.JWTPair{}, fmt.Errorf("jwt.Parse: %w", err)
	}
	claims := token.Claims.(jwt.MapClaims)

	teamIDFromClaims, ok := claims["sub"]
	if !ok {
		return models.JWTPair{}, errors.New("not found sub claim in jwt")
	}
	teamID, ok := teamIDFromClaims.(float64)
	if !ok {
		return models.JWTPair{}, errors.New("sub is not integer")
	}

	ok, err = s.authRepo.VerifyRefreshToken(ctx, int64(teamID), refreshToken)
	if err != nil {
		return models.JWTPair{}, fmt.Errorf("s.repo.VerifyRefreshToken: %w", err)
	}
	if !ok {
		return models.JWTPair{}, errors.New("such refresh token not exist")
	}

	var additionalClaims map[string]any
	if int64(teamID) == administratorTeamID {
		additionalClaims = s.getClaimsForAdmin(ctx)
	} else {
		additionalClaims = s.getClaimsForTeam(
			ctx,
			&models.Team{
				ID:   int64(teamID),
				Name: claims["name"].(string),
			},
		)
	}

	jwtPair, err := s.getJWTPair(ctx, additionalClaims)
	if err != nil {
		return models.JWTPair{}, fmt.Errorf("s.getJWTPair: %w", err)
	}
	if err != nil {
		return models.JWTPair{}, fmt.Errorf("s.getJWTPair: %w", err)
	}

	if err = s.authRepo.SetRefreshToken(ctx, int64(teamID), jwtPair.RefreshToken); err != nil {
		return models.JWTPair{}, fmt.Errorf("s.repo.SetRefreshToken: %w", err)
	}

	return jwtPair, nil
}

func (s *Service) RefreshTokenExpTime() time.Duration {
	return s.jwtConfig.JWTRefreshExpirationTime
}

func (s *Service) getClaimsForTeam(_ context.Context, team *models.Team) map[string]any {
	claims := map[string]any{
		"sub":  team.ID,
		"name": team.Name,
		"role": "team",
	}
	return claims
}

func (s *Service) getClaimsForAdmin(_ context.Context) map[string]any {
	claims := map[string]any{
		"role": "admin",
	}
	return claims
}

func (s *Service) getJWTPair(_ context.Context, additionalClaims map[string]any) (models.JWTPair, error) {
	atClaims := make(jwt.MapClaims, len(additionalClaims)+1)
	atClaims["exp"] = time.Now().Add(s.jwtConfig.JWTAccessExpirationTime).Unix()
	for key, value := range additionalClaims {
		atClaims[key] = value
	}
	at := jwt.NewWithClaims(jwt.SigningMethodHS256, atClaims)
	accessToken, err := at.SignedString([]byte(s.jwtConfig.JWTAccessSecretKey))
	if err != nil {
		return models.JWTPair{}, fmt.Errorf("token.SignedString: %w", err)
	}

	rtClaims := make(jwt.MapClaims, len(additionalClaims)+1)
	rtClaims["exp"] = time.Now().Add(s.jwtConfig.JWTRefreshExpirationTime).Unix()
	for key, value := range additionalClaims {
		rtClaims[key] = value
	}
	rt := jwt.NewWithClaims(jwt.SigningMethodHS256, rtClaims)
	refreshToken, err := rt.SignedString([]byte(s.jwtConfig.JWTRefreshSecretKey))
	if err != nil {
		return models.JWTPair{}, fmt.Errorf("token.SignedString: %w", err)
	}

	return models.JWTPair{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}
