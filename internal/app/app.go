package app

import (
	"errors"
	"investment-game-backend/internal/config"
	pgrepo "investment-game-backend/internal/repo/pg"
	additionalinfos "investment-game-backend/internal/services/additional_infos"
	"investment-game-backend/internal/services/auth"
	"investment-game-backend/internal/services/companies"
	"investment-game-backend/internal/services/games"
	"investment-game-backend/internal/services/settings"
	"investment-game-backend/internal/services/teams"
	v1 "investment-game-backend/internal/transport/http/v1"
	"investment-game-backend/pkg/http/server"
	"investment-game-backend/pkg/logger"
	"investment-game-backend/pkg/postgres"
	"net/http"
	"os"
	"os/signal"
	"syscall"
)

func Run(cfg *config.Config) {
	log := logger.New(logger.Config{
		Level:         "debug",
		FilePath:      "",
		NeedLogToFile: false,
	})

	pg, err := postgres.New(cfg.Postgres.URL)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to connect to database")
	}
	log.Info().Msg("successfully connected to database")

	if err = postgres.RunMigrations(pg.DB, cfg.Postgres.MigrationsPath); err != nil {
		log.Fatal().Err(err).Msg("failed to run migrations")
	}
	log.Info().Msg("successfully applied migrations")

	additionalInfosRepo := pgrepo.NewAdditionalInfosRepo(pg)
	authRepo := pgrepo.NewAuthRepo(pg)
	balanceTransactionsRepo := pgrepo.NewBalanceTransactionsRepo(pg)
	balancesRepo := pgrepo.NewBalancesRepo(pg)
	companiesRepo := pgrepo.NewCompaniesRepo(pg)
	companySharesRepo := pgrepo.NewCompanySharesRepo(pg)
	gamesRepo := pgrepo.NewGamesRepo(pg)
	settingsRepo := pgrepo.NewSettingsRepo(pg)
	teamsRepo := pgrepo.NewTeamsRepo(pg)

	authService := auth.New(
		teamsRepo,
		authRepo,
		auth.JWTConfig{
			JWTAccessExpirationTime:  cfg.JWT.JWTAccessExpirationTime,
			JWTRefreshExpirationTime: cfg.JWT.JWTRefreshExpirationTime,
			JWTAccessSecretKey:       cfg.JWT.JWTAccessSecretKey,
			JWTRefreshSecretKey:      cfg.JWT.JWTRefreshSecretKey,
		},
		auth.AdminCredentials{
			Username: cfg.Admin.Username,
			Password: cfg.Admin.Password,
		},
		log,
	)
	companiesService := companies.New(
		companiesRepo,
		companySharesRepo,
		gamesRepo,
		log,
	)
	gamesService := games.New(gamesRepo, log)
	settingsService := settings.New(settingsRepo, log)
	teamsService := teams.New(
		teamsRepo,
		balancesRepo,
		settingsRepo,
		additionalInfosRepo,
		companySharesRepo,
		balanceTransactionsRepo,
		gamesRepo,
		log,
	)
	additionalInfosService := additionalinfos.New(additionalInfosRepo, log)

	router := v1.NewRouter(v1.Config{
		SettingsService:       settingsService,
		CompaniesService:      companiesService,
		GamesService:          gamesService,
		TeamsService:          teamsService,
		AuthService:           authService,
		AdditionalInfoService: additionalInfosService,
	})

	httpServer := server.New(server.Config{
		Addr:    cfg.HTTP.Addr,
		Handler: router.GetHTTPHandler(),
	})

	go func() {
		if err = httpServer.Run(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Error().Err(err).Msg("error on server run")
		}
	}()
	log.Info().Msg("http server started on " + cfg.HTTP.Addr)

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGINT, syscall.SIGKILL)
	<-quit

	if err = httpServer.Stop(); err != nil {
		log.Error().Err(err).Msg("failed to stop http server")
	}
	log.Info().Msg("successfully stopped http server")

	if err = pg.Close(); err != nil {
		log.Error().Err(err).Msg("failed to close database")
	}
	log.Info().Msg("successfully closed database connect")
}
