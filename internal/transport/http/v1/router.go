package v1

import (
	"github.com/go-chi/chi/v5"
	"github.com/rs/zerolog"
	"investment-game-backend/internal/services"
	"net/http"
)

type Router struct {
	router chi.Router
	log    *zerolog.Logger

	settingsService       services.Settings
	companiesService      services.Companies
	gamesService          services.Games
	teamService           services.Teams
	authService           services.Auth
	additionalInfoService services.AdditionalInfos
}

type Config struct {
	SettingsService       services.Settings
	CompaniesService      services.Companies
	GamesService          services.Games
	TeamsService          services.Teams
	AuthService           services.Auth
	AdditionalInfoService services.AdditionalInfos
}

func NewRouter(cfg Config) *Router {
	r := Router{
		router:                chi.NewRouter(),
		settingsService:       cfg.SettingsService,
		companiesService:      cfg.CompaniesService,
		gamesService:          cfg.GamesService,
		teamService:           cfg.TeamsService,
		authService:           cfg.AuthService,
		additionalInfoService: cfg.AdditionalInfoService,
	}

	r.initRouter()

	return &r
}

func (r *Router) initRouter() {
	apiRouter := chi.NewRouter()
	r.initGamesRoutes(apiRouter)
	r.initSettingsRoutes(apiRouter)
	r.initAuthRoutes(apiRouter)
	r.initCompanyRoutes(apiRouter)
	r.initAdditionalInfosRoutes(apiRouter)

	r.router.Mount("/api", apiRouter)
}

func (r *Router) GetHTTPHandler() http.Handler {
	return r.router
}
