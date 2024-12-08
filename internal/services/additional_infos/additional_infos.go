package additional_infos

import (
	"context"
	"fmt"
	"github.com/rs/zerolog"
	"investment-game-backend/internal/models"
	"investment-game-backend/internal/repo"
)

type Service struct {
	repo repo.AdditionalInfosRepo
	log  *zerolog.Logger
}

func New(
	repo repo.AdditionalInfosRepo,
	log *zerolog.Logger,
) *Service {
	return &Service{
		repo: repo,
		log:  log,
	}
}

type CreateParams struct {
	Name        string
	Description string
	Type        models.AdditionalInfoType
	Cost        int64
	CompanyID   *int64
}

func (s *Service) Create(ctx context.Context, params CreateParams) (*models.AdditionalInfo, error) {
	info := &models.AdditionalInfo{
		Name:        params.Name,
		Description: params.Description,
		Type:        params.Type,
		Cost:        params.Cost,
		CompanyID:   params.CompanyID,
	}
	id, err := s.repo.Create(ctx, info)
	if err != nil {
		return nil, fmt.Errorf("s.repo.Create: %w", err)
	}
	info.ID = id
	return info, nil
}

type UpdateParams struct {
	ID          int64
	Name        string
	Description string
	Cost        int64
	CompanyID   *int64
}

func (s *Service) Update(ctx context.Context, params UpdateParams) error {
	info, err := s.repo.GetByID(ctx, params.ID)
	if err != nil {
		return fmt.Errorf("s.repo.GetByID: %w", err)
	}

	info.Name = params.Name
	info.Description = params.Description
	info.CompanyID = params.CompanyID
	info.Cost = params.Cost

	if err = s.repo.Update(ctx, info); err != nil {
		return fmt.Errorf("s.repo.Update: %w", err)
	}
	return nil
}

func (s *Service) GetActualListByType(
	ctx context.Context,
	infoType models.AdditionalInfoType,
) ([]models.AdditionalInfo, error) {
	infos, err := s.repo.GetAllActualWithType(ctx, infoType)
	if err != nil {
		return nil, fmt.Errorf("s.repo.GetAllActualWithType: %w", err)
	}
	return infos, nil
}
