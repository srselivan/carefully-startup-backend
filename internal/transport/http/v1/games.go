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
	})
}

type (
	getGameResp struct {
		State        models.GameState  `json:"state"`
		CurrentRound int               `json:"currentRound"`
		RoundState   models.RoundState `json:"roundState"`
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
			RoundState:   game.RoundState,
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
		RoundState   models.RoundState `json:"roundState"`
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
			RoundState:   request.RoundState,
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
