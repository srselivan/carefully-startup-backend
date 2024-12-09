package v1

import (
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/cors"
	"github.com/gorilla/websocket"
	"github.com/rs/zerolog"
	"investment-game-backend/internal/services"
	"investment-game-backend/internal/services/games"
	"net/http"
)

type Router struct {
	router                chi.Router
	log                   *zerolog.Logger
	secretJWT             string
	settingsService       services.Settings
	companiesService      services.Companies
	gamesService          services.Games
	teamService           services.Teams
	authService           services.Auth
	additionalInfoService services.AdditionalInfos
	upgrader              websocket.Upgrader
	teamsNotifier         *games.TeamsNotifier
}

type Config struct {
	SettingsService       services.Settings
	CompaniesService      services.Companies
	GamesService          services.Games
	TeamsService          services.Teams
	AuthService           services.Auth
	AdditionalInfoService services.AdditionalInfos
	SecretJWT             string
	Log                   *zerolog.Logger
	TeamsNotifier         *games.TeamsNotifier
}

func NewRouter(cfg Config) *Router {
	r := Router{
		router:                chi.NewRouter(),
		secretJWT:             cfg.SecretJWT,
		settingsService:       cfg.SettingsService,
		companiesService:      cfg.CompaniesService,
		gamesService:          cfg.GamesService,
		teamService:           cfg.TeamsService,
		authService:           cfg.AuthService,
		additionalInfoService: cfg.AdditionalInfoService,
		log:                   cfg.Log,
		upgrader:              websocket.Upgrader{},
		teamsNotifier:         cfg.TeamsNotifier,
	}

	r.initRouter()

	return &r
}

func (r *Router) initRouter() {
	apiRouter := chi.NewRouter()
	apiRouter.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"https://*", "http://*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "PATCH", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type"},
		AllowCredentials: true,
	}))
	r.initGamesRoutes(apiRouter)
	r.initSettingsRoutes(apiRouter)
	r.initAuthRoutes(apiRouter)
	r.initCompanyRoutes(apiRouter)
	r.initAdditionalInfosRoutes(apiRouter)
	r.initTeamsRoutes(apiRouter)

	r.router.Mount("/api", apiRouter)
}

func (r *Router) GetHTTPHandler() http.Handler {
	return r.router
}
