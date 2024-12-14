package v1

import (
	"github.com/go-chi/chi/v5"
	jsoniter "github.com/json-iterator/go"
	"github.com/samber/lo"
	"investment-game-backend/internal/models"
	"investment-game-backend/internal/services/teams"
	"io"
	"net/http"
	"strconv"
)

func (r *Router) initTeamsRoutes(router chi.Router) {
	router.Route("/team", func(subRouter chi.Router) {
		subRouter.Use(r.AuthMiddleware)
		subRouter.Patch("/", r.updateTeam)
		subRouter.Post("/purchase", r.teamPurchase)
		subRouter.Post("/purchase/additional-info/{team_id}", r.teamPurchaseAdditionalInfo)
		subRouter.Get("/{team_id}", r.getTeamByID)
		subRouter.Get("/", r.getAllTeams)
	})
}

type (
	updateTeamReq struct {
		ID      int64    `json:"id"`
		Name    string   `json:"name"`
		Members []string `json:"members"`
	}
)

func (r *Router) updateTeam(resp http.ResponseWriter, req *http.Request) {
	body, err := io.ReadAll(req.Body)
	if err != nil {
		r.log.Error().Err(err).Msg("error on request body read")
		resp.WriteHeader(http.StatusInternalServerError)
		_, _ = resp.Write([]byte(http.StatusText(http.StatusInternalServerError)))
		return
	}

	var request updateTeamReq
	if err = jsoniter.Unmarshal(body, &request); err != nil {
		r.log.Error().Err(err).Msg("json unmarshal error")
		resp.WriteHeader(http.StatusInternalServerError)
		_, _ = resp.Write([]byte(http.StatusText(http.StatusInternalServerError)))
		return
	}

	if err = r.teamService.Update(
		req.Context(),
		teams.UpdateParams{
			ID:      request.ID,
			Name:    request.Name,
			Members: request.Members,
		},
	); err != nil {
		r.log.Error().Err(err).Msg("update team error")
		resp.WriteHeader(http.StatusInternalServerError)
		_, _ = resp.Write([]byte(http.StatusText(http.StatusInternalServerError)))
		return
	}

	resp.WriteHeader(http.StatusOK)
	return
}

type (
	teamPurchaseReq struct {
		TeamID           int64           `json:"id"`
		SharesChanges    map[int64]int64 `json:"sharesChanges"`
		AdditionalInfoID *int64          `json:"additionalInfoId"`
	}
	teamPurchaseResp struct {
		BalanceAmount int64 `json:"balanceAmount"`
	}
)

func (r *Router) teamPurchase(resp http.ResponseWriter, req *http.Request) {
	body, err := io.ReadAll(req.Body)
	if err != nil {
		r.log.Error().Err(err).Msg("error on request body read")
		resp.WriteHeader(http.StatusInternalServerError)
		_, _ = resp.Write([]byte(http.StatusText(http.StatusInternalServerError)))
		return
	}

	var request teamPurchaseReq
	if err = jsoniter.Unmarshal(body, &request); err != nil {
		r.log.Error().Err(err).Msg("json unmarshal error")
		resp.WriteHeader(http.StatusInternalServerError)
		_, _ = resp.Write([]byte(http.StatusText(http.StatusInternalServerError)))
		return
	}

	amount, err := r.teamService.Purchase(
		req.Context(),
		teams.PurchaseParams{
			TeamID:           request.TeamID,
			SharesChanges:    request.SharesChanges,
			AdditionalInfoID: request.AdditionalInfoID,
		},
	)
	if err != nil {
		r.log.Error().Err(err).Msg("purchase error")
		resp.WriteHeader(http.StatusInternalServerError)
		_, _ = resp.Write([]byte(err.Error()))
		return
	}

	response, err := jsoniter.Marshal(
		teamPurchaseResp{
			BalanceAmount: amount,
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
	getTeamByIDResp struct {
		TeamID                    int64                           `json:"id"`
		Name                      string                          `json:"name"`
		Members                   []string                        `json:"members"`
		Shares                    map[int64]int64                 `json:"shares"`
		AdditionalInfoIds         []int64                         `json:"additionalInfoIds"`
		RandomEventID             *int64                          `json:"randomEventId"`
		AdditionalInfos           []getTeamByIDRespAdditionalInfo `json:"additionalInfos"`
		BalanceAmount             int64                           `json:"balanceAmount"`
		HasTransactionInThisRound bool                            `json:"hasTransactionInThisRound"`
	}
	getTeamByIDRespAdditionalInfo struct {
		ID          int64  `json:"id"`
		Name        string `json:"name"`
		Description string `json:"description"`
		Type        int    `json:"type"`
		Cost        int64  `json:"cost"`
		CompanyID   *int64 `json:"companyId"`
	}
)

func (r *Router) getTeamByID(resp http.ResponseWriter, req *http.Request) {
	teamIDParam := chi.URLParam(req, "team_id")
	teamID, err := strconv.Atoi(teamIDParam)
	if err != nil {
		r.log.Error().Err(err).Msg("get path param")
		resp.WriteHeader(http.StatusInternalServerError)
		_, _ = resp.Write([]byte(http.StatusText(http.StatusInternalServerError)))
		return
	}

	detailedTeam, err := r.teamService.GetDetailedByID(req.Context(), int64(teamID))
	if err != nil {
		r.log.Error().Err(err).Msg("get by id error")
		resp.WriteHeader(http.StatusInternalServerError)
		_, _ = resp.Write([]byte(http.StatusText(http.StatusInternalServerError)))
		return
	}

	response, err := jsoniter.Marshal(
		getTeamByIDResp{
			TeamID:            detailedTeam.Team.ID,
			Name:              detailedTeam.Team.Name,
			Members:           detailedTeam.Team.Members,
			Shares:            detailedTeam.Team.Shares,
			AdditionalInfoIds: detailedTeam.Team.AdditionalInfos,
			RandomEventID:     detailedTeam.Team.RandomEventID,
			AdditionalInfos: lo.Map(
				detailedTeam.AdditionalInfos,
				func(item models.AdditionalInfo, _ int) getTeamByIDRespAdditionalInfo {
					return getTeamByIDRespAdditionalInfo{
						ID:          item.ID,
						Name:        item.Name,
						Description: item.Description,
						Type:        int(item.Type),
						Cost:        item.Cost,
						CompanyID:   item.CompanyID,
					}
				},
			),
			BalanceAmount:             detailedTeam.Balance,
			HasTransactionInThisRound: detailedTeam.HasTransactionInThisRound,
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
	getAllTeamsResp struct {
		Data []getAllTeamsRespItem `json:"data"`
	}
	getAllTeamsRespItem struct {
		ID   int64  `json:"id"`
		Name string `json:"name"`
	}
)

func (r *Router) getAllTeams(resp http.ResponseWriter, req *http.Request) {
	ts, err := r.teamService.GetAllForCurrentGame(req.Context())
	if err != nil {
		r.log.Error().Err(err).Msg("GetAllForCurrentGame error")
		resp.WriteHeader(http.StatusInternalServerError)
		_, _ = resp.Write([]byte(err.Error()))
		return
	}

	response, err := jsoniter.Marshal(
		getAllTeamsResp{
			Data: lo.Map(ts, func(item models.Team, _ int) getAllTeamsRespItem {
				return getAllTeamsRespItem{
					ID:   item.ID,
					Name: item.Name,
				}
			}),
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
	teamPurchaseAdditionalInfoResp struct {
		ID            int64                     `json:"id"`
		Name          string                    `json:"name"`
		Description   string                    `json:"description"`
		Type          models.AdditionalInfoType `json:"type"`
		Cost          int64                     `json:"cost"`
		CompanyID     *int64                    `json:"companyId"`
		BalanceAmount int64                     `json:"balanceAmount"`
	}
)

func (r *Router) teamPurchaseAdditionalInfo(resp http.ResponseWriter, req *http.Request) {
	teamIDParam := chi.URLParam(req, "team_id")
	teamID, err := strconv.Atoi(teamIDParam)
	if err != nil {
		r.log.Error().Err(err).Msg("get path param")
		resp.WriteHeader(http.StatusInternalServerError)
		_, _ = resp.Write([]byte(http.StatusText(http.StatusInternalServerError)))
		return
	}

	addInfo, amount, err := r.teamService.PurchaseAdditionalInfoCompanyInfo(req.Context(), int64(teamID))
	if err != nil {
		r.log.Error().Err(err).Msg("purchase error")
		resp.WriteHeader(http.StatusInternalServerError)
		_, _ = resp.Write([]byte(err.Error()))
		return
	}

	response, err := jsoniter.Marshal(
		teamPurchaseAdditionalInfoResp{
			ID:            addInfo.ID,
			Name:          addInfo.Name,
			Description:   addInfo.Description,
			Type:          addInfo.Type,
			Cost:          addInfo.Cost,
			CompanyID:     addInfo.CompanyID,
			BalanceAmount: amount,
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
