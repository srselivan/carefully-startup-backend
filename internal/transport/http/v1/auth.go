package v1

import (
	"errors"
	"github.com/go-chi/chi/v5"
	jsoniter "github.com/json-iterator/go"
	"investment-game-backend/internal/services/teams"
	"io"
	"net/http"
	"time"
)

func (r *Router) initAuthRoutes(router chi.Router) {
	router.Route("/auth", func(settingsRouter chi.Router) {
		settingsRouter.Post("/registration", r.registration)
		settingsRouter.Post("/login", r.login)
		settingsRouter.Post("/refresh", r.refresh)
	})
}

type (
	registrationReq struct {
		TeamName string `json:"teamName"`
		Password string `json:"password"`
	}
)

func (r *Router) registration(resp http.ResponseWriter, req *http.Request) {
	body, err := io.ReadAll(req.Body)
	if err != nil {
		r.log.Error().Err(err).Msg("error on request body read")
		resp.WriteHeader(http.StatusInternalServerError)
		_, _ = resp.Write([]byte(http.StatusText(http.StatusInternalServerError)))
		return
	}

	var request registrationReq
	if err = jsoniter.Unmarshal(body, &request); err != nil {
		r.log.Error().Err(err).Msg("json unmarshal error")
		resp.WriteHeader(http.StatusInternalServerError)
		_, _ = resp.Write([]byte(http.StatusText(http.StatusInternalServerError)))
		return
	}

	_, err = r.teamService.Create(req.Context(), teams.CreateParams{
		Name:        request.TeamName,
		Credentials: request.TeamName + ":" + request.Password,
	})
	if err != nil {
		r.log.Error().Err(err).Msg("create team error")
		if errors.Is(err, teams.ErrNoRegistrationPeriod) {
			resp.WriteHeader(http.StatusBadRequest)
			return
		}
		resp.WriteHeader(http.StatusInternalServerError)
		_, _ = resp.Write([]byte(err.Error()))
		return
	}

	resp.WriteHeader(http.StatusCreated)
	return
}

type (
	loginReq struct {
		TeamName string `json:"teamName"`
		Password string `json:"password"`
		IsAdmin  bool   `json:"isAdmin"`
	}
	loginResp struct {
		AccessToken  string `json:"accessToken"`
		RefreshToken string `json:"refreshToken"`
	}
)

func (r *Router) login(resp http.ResponseWriter, req *http.Request) {
	body, err := io.ReadAll(req.Body)
	if err != nil {
		r.log.Error().Err(err).Msg("error on request body read")
		resp.WriteHeader(http.StatusInternalServerError)
		_, _ = resp.Write([]byte(http.StatusText(http.StatusInternalServerError)))
		return
	}

	var request loginReq
	if err = jsoniter.Unmarshal(body, &request); err != nil {
		r.log.Error().Err(err).Msg("json unmarshal error")
		resp.WriteHeader(http.StatusInternalServerError)
		_, _ = resp.Write([]byte(http.StatusText(http.StatusInternalServerError)))
		return
	}

	jwtPair, err := r.authService.Login(req.Context(), request.TeamName+":"+request.Password, request.IsAdmin)
	if err != nil {
		r.log.Error().Err(err).Msg("login error")
		resp.WriteHeader(http.StatusUnauthorized)
		_, _ = resp.Write([]byte(http.StatusText(http.StatusUnauthorized)))
		return
	}

	cookie := &http.Cookie{
		Name:     "refreshToken",
		Value:    jwtPair.RefreshToken,
		Expires:  time.Now().Add(r.authService.RefreshTokenExpTime()),
		Secure:   false,
		HttpOnly: true,
	}

	http.SetCookie(resp, cookie)

	response, err := jsoniter.Marshal(
		loginResp{
			AccessToken:  jwtPair.AccessToken,
			RefreshToken: jwtPair.RefreshToken,
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
	refreshResp struct {
		AccessToken  string `json:"accessToken"`
		RefreshToken string `json:"refreshToken"`
	}
)

func (r *Router) refresh(resp http.ResponseWriter, req *http.Request) {
	var refreshToken string

	cookie, err := req.Cookie("refreshToken")
	if errors.Is(err, http.ErrNoCookie) {
		body, err := io.ReadAll(req.Body)
		if err != nil {
			r.log.Error().Err(err).Msg("error on request body read")
			resp.WriteHeader(http.StatusInternalServerError)
			_, _ = resp.Write([]byte(http.StatusText(http.StatusInternalServerError)))
			return
		}
		refreshToken = string(body)
	} else {
		refreshToken = cookie.Value
	}

	jwtPair, err := r.authService.Refresh(req.Context(), refreshToken)
	if err != nil {
		r.log.Error().Err(err).Str("token", refreshToken).Msg("refresh error")
		resp.WriteHeader(http.StatusInternalServerError)
		_, _ = resp.Write([]byte(http.StatusText(http.StatusInternalServerError)))
		return
	}

	cookie = &http.Cookie{
		Name:     "refreshToken",
		Value:    jwtPair.RefreshToken,
		Expires:  time.Now().Add(r.authService.RefreshTokenExpTime()),
		Secure:   false,
		HttpOnly: true,
	}

	http.SetCookie(resp, cookie)

	response, err := jsoniter.Marshal(
		refreshResp{
			AccessToken:  jwtPair.AccessToken,
			RefreshToken: jwtPair.RefreshToken,
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
