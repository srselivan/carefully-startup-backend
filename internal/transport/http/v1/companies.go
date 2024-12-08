package v1

import (
	"github.com/go-chi/chi/v5"
	jsoniter "github.com/json-iterator/go"
	"github.com/samber/lo"
	"investment-game-backend/internal/models"
	"investment-game-backend/internal/services/companies"
	"io"
	"net/http"
	"strconv"
)

func (r *Router) initCompanyRoutes(router chi.Router) {
	router.Route("/company", func(companyRouter chi.Router) {
		companyRouter.Post("/", r.createCompanyWithShares)
		companyRouter.Get("/", r.getCompaniesWithShares)
		companyRouter.Put("/{company_id}", r.updateCompanyWithShares)
		companyRouter.Patch("/{company_id}", r.archiveCompanyWithShares)
	})
}

type (
	createCompanyReq struct {
		Name   string             `json:"name"`
		Shares map[string]float64 `json:"shares"`
	}
	createCompanyResp struct {
		ID     int64              `json:"id"`
		Name   string             `json:"name"`
		Shares map[string]float64 `json:"shares"`
	}
)

func (r *Router) createCompanyWithShares(resp http.ResponseWriter, req *http.Request) {
	body, err := io.ReadAll(req.Body)
	if err != nil {
		r.log.Error().Err(err).Msg("error on request body read")
		resp.WriteHeader(http.StatusInternalServerError)
		_, _ = resp.Write([]byte(http.StatusText(http.StatusInternalServerError)))
		return
	}

	var request createCompanyReq
	if err = jsoniter.Unmarshal(body, &request); err != nil {
		r.log.Error().Err(err).Msg("json unmarshal error")
		resp.WriteHeader(http.StatusInternalServerError)
		_, _ = resp.Write([]byte(http.StatusText(http.StatusInternalServerError)))
		return
	}

	createdCompanyID, err := r.companiesService.CreateWithShares(
		req.Context(),
		companies.CreateWithSharesParams{
			Name: request.Name,
			Shares: lo.MapEntries(
				request.Shares,
				func(round string, price float64) (int, int64) {
					roundItn, _ := strconv.Atoi(round)
					return roundItn, int64(price)
				},
			),
		},
	)
	if err != nil {
		r.log.Error().Err(err).Msg("CreateWithShares error")
		resp.WriteHeader(http.StatusInternalServerError)
		_, _ = resp.Write([]byte(http.StatusText(http.StatusInternalServerError)))
		return
	}

	response, err := jsoniter.Marshal(
		createCompanyResp{
			ID:     createdCompanyID,
			Name:   request.Name,
			Shares: request.Shares,
		},
	)
	if err != nil {
		r.log.Error().Err(err).Msg("marshal to json error")
		resp.WriteHeader(http.StatusInternalServerError)
		_, _ = resp.Write([]byte(http.StatusText(http.StatusInternalServerError)))
		return
	}

	resp.WriteHeader(http.StatusCreated)
	_, _ = resp.Write(response)
	return
}

type (
	getCompaniesWithSharesResp struct {
		Data []getCompaniesWithSharesRespItem `json:"data"`
	}
	getCompaniesWithSharesRespItem struct {
		ID     int64         `json:"id"`
		Name   string        `json:"name"`
		Shares map[int]int64 `json:"shares"`
	}
)

func (r *Router) getCompaniesWithShares(resp http.ResponseWriter, req *http.Request) {
	companyWithShares, err := r.companiesService.GetAllWithShares(req.Context(), false)
	if err != nil {
		r.log.Error().Err(err).Msg("GetAllWithShares error")
		resp.WriteHeader(http.StatusInternalServerError)
		_, _ = resp.Write([]byte(http.StatusText(http.StatusInternalServerError)))
		return
	}

	response, err := jsoniter.Marshal(
		getCompaniesWithSharesResp{
			Data: lo.Map(
				companyWithShares,
				func(item models.CompanyWithShares, _ int) getCompaniesWithSharesRespItem {
					return getCompaniesWithSharesRespItem{
						ID:     item.ID,
						Name:   item.Name,
						Shares: item.Shares,
					}
				},
			),
		},
	)
	if err != nil {
		r.log.Error().Err(err).Msg("marshal to json error")
		resp.WriteHeader(http.StatusInternalServerError)
		_, _ = resp.Write([]byte(http.StatusText(http.StatusInternalServerError)))
		return
	}

	resp.WriteHeader(http.StatusCreated)
	_, _ = resp.Write(response)
	return
}

type (
	updateCompanyReq struct {
		Name   string             `json:"name"`
		Shares map[string]float64 `json:"shares"`
	}
)

func (r *Router) updateCompanyWithShares(resp http.ResponseWriter, req *http.Request) {
	companyIDParam := chi.URLParam(req, "company_id")
	companyID, err := strconv.Atoi(companyIDParam)
	if err != nil {
		r.log.Error().Err(err).Msg("get path param")
		resp.WriteHeader(http.StatusInternalServerError)
		_, _ = resp.Write([]byte(http.StatusText(http.StatusInternalServerError)))
		return
	}

	body, err := io.ReadAll(req.Body)
	if err != nil {
		r.log.Error().Err(err).Msg("error on request body read")
		resp.WriteHeader(http.StatusInternalServerError)
		_, _ = resp.Write([]byte(http.StatusText(http.StatusInternalServerError)))
		return
	}

	var request updateCompanyReq
	if err = jsoniter.Unmarshal(body, &request); err != nil {
		r.log.Error().Err(err).Msg("json unmarshal error")
		resp.WriteHeader(http.StatusInternalServerError)
		_, _ = resp.Write([]byte(http.StatusText(http.StatusInternalServerError)))
		return
	}

	err = r.companiesService.Update(
		req.Context(),
		companies.UpdateParams{
			ID:   int64(companyID),
			Name: request.Name,
			Shares: lo.MapEntries(
				request.Shares,
				func(round string, price float64) (int, int64) {
					roundItn, _ := strconv.Atoi(round)
					return roundItn, int64(price)
				},
			),
		},
	)
	if err != nil {
		r.log.Error().Err(err).Msg("Update error")
		resp.WriteHeader(http.StatusInternalServerError)
		_, _ = resp.Write([]byte(http.StatusText(http.StatusInternalServerError)))
		return
	}

	resp.WriteHeader(http.StatusOK)
	return
}

func (r *Router) archiveCompanyWithShares(resp http.ResponseWriter, req *http.Request) {
	companyIDParam := chi.URLParam(req, "company_id")
	companyID, err := strconv.Atoi(companyIDParam)
	if err != nil {
		r.log.Error().Err(err).Msg("get path param")
		resp.WriteHeader(http.StatusInternalServerError)
		_, _ = resp.Write([]byte(http.StatusText(http.StatusInternalServerError)))
		return
	}

	if err = r.companiesService.Archive(req.Context(), int64(companyID)); err != nil {
		r.log.Error().Err(err).Msg("Update error")
		resp.WriteHeader(http.StatusInternalServerError)
		_, _ = resp.Write([]byte(http.StatusText(http.StatusInternalServerError)))
		return
	}

	resp.WriteHeader(http.StatusOK)
	return
}
