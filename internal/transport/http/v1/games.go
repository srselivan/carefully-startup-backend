package v1

import (
	"github.com/go-chi/chi/v5"
	jsoniter "github.com/json-iterator/go"
	"investment-game-backend/internal/models"
	gamesservice "investment-game-backend/internal/services/games"
	"io"
	"net/http"
)

func (r *Router) initGamesRoutes(router chi.Router) {
	router.Route("/game", func(gameRouter chi.Router) {
		gameRouter.Get("/", r.getGame)
		gameRouter.Put("/", r.updateGame)
		gameRouter.Patch("/create", r.createNewGame)
		gameRouter.Patch("/start", r.startGame)
		gameRouter.Patch("/registration/start", r.startRegistration)
		gameRouter.Patch("/registration/stop", r.stopRegistration)
		gameRouter.Patch("/round/start", r.startRound)
		gameRouter.Patch("/trade/start", r.startTrade)
		gameRouter.Patch("/trade/stop", r.stopTrade)
	})
}

type (
	getGameResp struct {
		State        models.GameState  `json:"state"`
		CurrentRound int               `json:"currentRound"`
		TradeState   models.TradeState `json:"tradeState"`
	}
)

func (r *Router) getGame(resp http.ResponseWriter, req *http.Request) {
	game, err := r.gamesService.Get(req.Context())
	if err != nil {
		r.log.Error().Err(err).Msg("games service: get error")
		resp.WriteHeader(http.StatusInternalServerError)
		_, _ = resp.Write([]byte(http.StatusText(http.StatusInternalServerError)))
		return
	}

	response, err := jsoniter.Marshal(
		getGameResp{
			State:        game.State,
			CurrentRound: game.CurrentRound,
			TradeState:   game.TradeState,
		},
	)
	if err != nil {
		r.log.Error().Err(err).Msg("marshal to json error")
		resp.WriteHeader(http.StatusInternalServerError)
		_, _ = resp.Write([]byte(http.StatusText(http.StatusInternalServerError)))
		return
	}

	resp.WriteHeader(http.StatusOK)
	_, _ = resp.Write(response)
	return
}

type (
	updateGameReq struct {
		State        models.GameState  `json:"state"`
		CurrentRound int               `json:"currentRound"`
		TradeState   models.TradeState `json:"tradeState"`
	}
)

func (r *Router) updateGame(resp http.ResponseWriter, req *http.Request) {
	body, err := io.ReadAll(req.Body)
	if err != nil {
		r.log.Error().Err(err).Msg("error on request body read")
		resp.WriteHeader(http.StatusInternalServerError)
		_, _ = resp.Write([]byte(http.StatusText(http.StatusInternalServerError)))
		return
	}

	var request updateGameReq
	if err = jsoniter.Unmarshal(body, &request); err != nil {
		r.log.Error().Err(err).Msg("json unmarshal error")
		resp.WriteHeader(http.StatusInternalServerError)
		_, _ = resp.Write([]byte(http.StatusText(http.StatusInternalServerError)))
		return
	}

	if err = r.gamesService.Update(
		req.Context(),
		gamesservice.UpdateParams{
			State:        request.State,
			CurrentRound: request.CurrentRound,
			TradeState:   request.TradeState,
		},
	); err != nil {
		r.log.Error().Err(err).Msg("games service: update error")
		resp.WriteHeader(http.StatusInternalServerError)
		_, _ = resp.Write([]byte(http.StatusText(http.StatusInternalServerError)))
		return
	}

	resp.WriteHeader(http.StatusAccepted)
	return
}

func (r *Router) createNewGame(resp http.ResponseWriter, req *http.Request) {
	if err := r.gamesService.CreateNewGame(req.Context()); err != nil {
		r.log.Error().Err(err).Msg("StartNewGame error")
		resp.WriteHeader(http.StatusInternalServerError)
		_, _ = resp.Write([]byte(err.Error()))
		return
	}
	resp.WriteHeader(http.StatusOK)
	return
}

func (r *Router) startGame(resp http.ResponseWriter, req *http.Request) {
	if err := r.gamesService.StartGame(req.Context()); err != nil {
		r.log.Error().Err(err).Msg("StartGame error")
		resp.WriteHeader(http.StatusInternalServerError)
		_, _ = resp.Write([]byte(err.Error()))
		return
	}
	resp.WriteHeader(http.StatusOK)
	return
}

func (r *Router) startRegistration(resp http.ResponseWriter, req *http.Request) {
	if err := r.gamesService.StartRegistration(req.Context()); err != nil {
		r.log.Error().Err(err).Msg("StartRegistration error")
		resp.WriteHeader(http.StatusInternalServerError)
		_, _ = resp.Write([]byte(err.Error()))
		return
	}
	resp.WriteHeader(http.StatusOK)
	return
}

func (r *Router) stopRegistration(resp http.ResponseWriter, req *http.Request) {
	if err := r.gamesService.StopRegistration(req.Context()); err != nil {
		r.log.Error().Err(err).Msg("StopRegistration error")
		resp.WriteHeader(http.StatusInternalServerError)
		_, _ = resp.Write([]byte(err.Error()))
		return
	}
	resp.WriteHeader(http.StatusOK)
	return
}

func (r *Router) startRound(resp http.ResponseWriter, req *http.Request) {
	if err := r.gamesService.StartRound(req.Context()); err != nil {
		r.log.Error().Err(err).Msg("StartRound error")
		resp.WriteHeader(http.StatusInternalServerError)
		_, _ = resp.Write([]byte(err.Error()))
		return
	}
	resp.WriteHeader(http.StatusOK)
	return
}

func (r *Router) startTrade(resp http.ResponseWriter, req *http.Request) {
	if err := r.gamesService.StartTrade(req.Context()); err != nil {
		r.log.Error().Err(err).Msg("StartTrade error")
		resp.WriteHeader(http.StatusInternalServerError)
		_, _ = resp.Write([]byte(err.Error()))
		return
	}
	resp.WriteHeader(http.StatusOK)
	return
}

func (r *Router) stopTrade(resp http.ResponseWriter, req *http.Request) {
	r.gamesService.StopTrade(req.Context())
	resp.WriteHeader(http.StatusOK)
	return
}
