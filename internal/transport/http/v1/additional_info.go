package v1

import (
	"github.com/go-chi/chi/v5"
	jsoniter "github.com/json-iterator/go"
	"github.com/samber/lo"
	"investment-game-backend/internal/models"
	additionalinfos "investment-game-backend/internal/services/additional_infos"
	"io"
	"net/http"
	"strconv"
)

func (r *Router) initAdditionalInfosRoutes(router chi.Router) {
	router.Route("/additional-info", func(subRouter chi.Router) {
		subRouter.Post("/", r.createAdditionalInfo)
		subRouter.Put("/{additional_info_id}", r.updateAdditionalInfo)
		subRouter.Get("/", r.getAllActualAdditionalInfos)
	})
}

type (
	createAdditionalInfoReq struct {
		Name        string                    `json:"name"`
		Description string                    `json:"description"`
		Type        models.AdditionalInfoType `json:"type"`
		Cost        int64                     `json:"cost"`
		CompanyID   *int64                    `json:"companyId"`
	}
	createAdditionalInfoResp struct {
		ID          int64                     `json:"id"`
		Name        string                    `json:"name"`
		Description string                    `json:"description"`
		Type        models.AdditionalInfoType `json:"type"`
		Cost        int64                     `json:"cost"`
		CompanyID   *int64                    `json:"companyId"`
	}
)

func (r *Router) createAdditionalInfo(resp http.ResponseWriter, req *http.Request) {
	body, err := io.ReadAll(req.Body)
	if err != nil {
		r.log.Error().Err(err).Msg("error on request body read")
		resp.WriteHeader(http.StatusInternalServerError)
		_, _ = resp.Write([]byte(http.StatusText(http.StatusInternalServerError)))
		return
	}

	var request createAdditionalInfoReq
	if err = jsoniter.Unmarshal(body, &request); err != nil {
		r.log.Error().Err(err).Msg("json unmarshal error")
		resp.WriteHeader(http.StatusInternalServerError)
		_, _ = resp.Write([]byte(http.StatusText(http.StatusInternalServerError)))
		return
	}

	additionalInfo, err := r.additionalInfoService.Create(
		req.Context(),
		additionalinfos.CreateParams{
			Name:        request.Name,
			Description: request.Description,
			Type:        request.Type,
			Cost:        request.Cost,
			CompanyID:   request.CompanyID,
		},
	)
	if err != nil {
		r.log.Error().Err(err).Msg("create error")
		resp.WriteHeader(http.StatusInternalServerError)
		_, _ = resp.Write([]byte(http.StatusText(http.StatusInternalServerError)))
		return
	}

	response, err := jsoniter.Marshal(
		createAdditionalInfoResp{
			ID:          additionalInfo.ID,
			Name:        additionalInfo.Name,
			Description: additionalInfo.Description,
			Type:        additionalInfo.Type,
			Cost:        additionalInfo.Cost,
			CompanyID:   additionalInfo.CompanyID,
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
	updateAdditionalInfoReq struct {
		Name        string                    `json:"name"`
		Description string                    `json:"description"`
		Type        models.AdditionalInfoType `json:"type"`
		Cost        int64                     `json:"cost"`
		CompanyID   *int64                    `json:"companyId"`
	}
)

func (r *Router) updateAdditionalInfo(resp http.ResponseWriter, req *http.Request) {
	additionalInfoIDParam := chi.URLParam(req, "additional_info_id")
	additionalInfoID, err := strconv.Atoi(additionalInfoIDParam)
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

	var request updateAdditionalInfoReq
	if err = jsoniter.Unmarshal(body, &request); err != nil {
		r.log.Error().Err(err).Msg("json unmarshal error")
		resp.WriteHeader(http.StatusInternalServerError)
		_, _ = resp.Write([]byte(http.StatusText(http.StatusInternalServerError)))
		return
	}

	err = r.additionalInfoService.Update(
		req.Context(),
		additionalinfos.UpdateParams{
			ID:          int64(additionalInfoID),
			Name:        request.Name,
			Description: request.Description,
			Cost:        request.Cost,
			CompanyID:   request.CompanyID,
		},
	)
	if err != nil {
		r.log.Error().Err(err).Msg("update error")
		resp.WriteHeader(http.StatusInternalServerError)
		_, _ = resp.Write([]byte(http.StatusText(http.StatusInternalServerError)))
		return
	}

	resp.WriteHeader(http.StatusOK)
	return
}

type (
	getAllActualAdditionalInfosResp struct {
		Data []getAllActualAdditionalInfosRespItem `json:"data"`
	}
	getAllActualAdditionalInfosRespItem struct {
		ID          int64  `json:"id"`
		Name        string `json:"name"`
		Description string `json:"description"`
		Type        int    `json:"type"`
		Cost        int64  `json:"cost"`
		CompanyID   *int64 `json:"companyId"`
	}
)

func (r *Router) getAllActualAdditionalInfos(resp http.ResponseWriter, req *http.Request) {
	infoTypeParam := req.URL.Query().Get("type")
	if infoTypeParam == "" {
		resp.WriteHeader(http.StatusBadRequest)
		return
	}

	infoType, err := strconv.Atoi(infoTypeParam)
	if err != nil {
		r.log.Error().Err(err).Msg("get query param")
		resp.WriteHeader(http.StatusBadRequest)
		_, _ = resp.Write([]byte(http.StatusText(http.StatusBadRequest)))
		return
	}

	infos, err := r.additionalInfoService.GetActualListByType(req.Context(), models.AdditionalInfoType(infoType))
	if err != nil {
		r.log.Error().Err(err).Msg("GetActualListByType error")
		resp.WriteHeader(http.StatusInternalServerError)
		_, _ = resp.Write([]byte(http.StatusText(http.StatusInternalServerError)))
		return
	}

	response, err := jsoniter.Marshal(
		getAllActualAdditionalInfosResp{
			Data: lo.Map(
				infos,
				func(item models.AdditionalInfo, _ int) getAllActualAdditionalInfosRespItem {
					return getAllActualAdditionalInfosRespItem{
						ID:          item.ID,
						Name:        item.Name,
						Description: item.Description,
						Type:        int(item.Type),
						Cost:        item.Cost,
						CompanyID:   item.CompanyID,
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

	resp.WriteHeader(http.StatusOK)
	_, _ = resp.Write(response)
	return
}
