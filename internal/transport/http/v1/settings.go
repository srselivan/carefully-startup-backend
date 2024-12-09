package v1

import (
	"github.com/go-chi/chi/v5"
	jsoniter "github.com/json-iterator/go"
	settingsservice "investment-game-backend/internal/services/settings"
	"io"
	"net/http"
	"time"
)

func (r *Router) initSettingsRoutes(router chi.Router) {
	router.Route("/settings", func(settingsRouter chi.Router) {
		settingsRouter.Use(r.AuthMiddleware)
		settingsRouter.Get("/", r.getSettings)
		settingsRouter.Put("/", r.updateSettings)
	})
}

type (
	getSettingsResp struct {
		RoundsCount        int           `json:"roundsCount"`
		RoundsDuration     time.Duration `json:"roundsDuration"`
		LinkToPDF          string        `json:"linkToPdf"`
		EnableRandomEvents bool          `json:"enableRandomEvents"`
	}
)

func (r *Router) getSettings(resp http.ResponseWriter, req *http.Request) {
	settings, err := r.settingsService.Get(req.Context())
	if err != nil {
		r.log.Error().Err(err).Msg("settings service: get error")
		resp.WriteHeader(http.StatusInternalServerError)
		_, _ = resp.Write([]byte(http.StatusText(http.StatusInternalServerError)))
		return
	}

	response, err := jsoniter.Marshal(
		getSettingsResp{
			RoundsCount:        settings.RoundsCount,
			RoundsDuration:     settings.RoundsDuration,
			LinkToPDF:          settings.LinkToPDF,
			EnableRandomEvents: settings.EnableRandomEvents,
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
	updateSettingsReq struct {
		RoundsCount        int           `json:"roundsCount"`
		RoundsDuration     time.Duration `json:"roundsDuration"`
		LinkToPDF          string        `json:"linkToPdf"`
		EnableRandomEvents bool          `json:"enableRandomEvents"`
	}
)

func (r *Router) updateSettings(resp http.ResponseWriter, req *http.Request) {
	body, err := io.ReadAll(req.Body)
	if err != nil {
		r.log.Error().Err(err).Msg("error on request body read")
		resp.WriteHeader(http.StatusInternalServerError)
		_, _ = resp.Write([]byte(http.StatusText(http.StatusInternalServerError)))
		return
	}

	var request updateSettingsReq
	if err = jsoniter.Unmarshal(body, &request); err != nil {
		r.log.Error().Err(err).Msg("json unmarshal error")
		resp.WriteHeader(http.StatusInternalServerError)
		_, _ = resp.Write([]byte(http.StatusText(http.StatusInternalServerError)))
		return
	}

	if err = r.settingsService.Update(
		req.Context(),
		settingsservice.UpdateParams{
			RoundsCount:        request.RoundsCount,
			RoundsDuration:     request.RoundsDuration,
			LinkToPDF:          request.LinkToPDF,
			EnableRandomEvents: request.EnableRandomEvents,
		},
	); err != nil {
		r.log.Error().Err(err).Msg("settings service: update error")
		resp.WriteHeader(http.StatusInternalServerError)
		_, _ = resp.Write([]byte(http.StatusText(http.StatusInternalServerError)))
		return
	}

	resp.WriteHeader(http.StatusAccepted)
	return
}
