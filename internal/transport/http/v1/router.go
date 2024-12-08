package v1

import (
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/cors"
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
	Log                   *zerolog.Logger
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
		log:                   cfg.Log,
	}

	r.initRouter()

	return &r
}

func (r *Router) initRouter() {
	apiRouter := chi.NewRouter()
	apiRouter.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"https://*", "http://*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type"},
		AllowCredentials: true,
	}))
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
