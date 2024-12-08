package companies

import (
	"context"
	"errors"
	"fmt"
	"github.com/rs/zerolog"
	"github.com/samber/lo"
	"investment-game-backend/internal/models"
	"investment-game-backend/internal/repo"
)

type Service struct {
	repo       repo.CompaniesRepo
	sharesRepo repo.CompanySharesRepo
	gameRepo   repo.GamesRepo
	log        *zerolog.Logger
}

func New(
	repo repo.CompaniesRepo,
	sharesRepo repo.CompanySharesRepo,
	gameRepo repo.GamesRepo,
	log *zerolog.Logger,
) *Service {
	return &Service{
		repo:       repo,
		sharesRepo: sharesRepo,
		gameRepo:   gameRepo,
		log:        log,
	}
}

type CreateWithSharesParams struct {
	Name   string
	Shares map[int]int64
}

func (s *Service) CreateWithShares(ctx context.Context, params CreateWithSharesParams) (int64, error) {
	company := &models.Company{
		Name: params.Name,
	}
	createdID, err := s.repo.Create(ctx, company)
	if err != nil {
		return 0, fmt.Errorf("s.repo.Create: %w", err)
	}

	for round, sharePrice := range params.Shares {
		if _, err = s.sharesRepo.Create(
			ctx,
			&models.CompanyShare{
				CompanyID: createdID,
				Round:     round,
				Price:     sharePrice,
			},
		); err != nil {
			return 0, fmt.Errorf("s.sharesRepo.Create: %w", err)
		}
	}

	return createdID, nil
}

type UpdateParams struct {
	ID     int64
	Name   string
	Shares map[int]int64
}

func (s *Service) Update(ctx context.Context, params UpdateParams) error {
	company, err := s.repo.GetByID(ctx, params.ID)
	if err != nil {
		return fmt.Errorf("s.repo.GetByID: %w", err)
	}

	if company.IsArchived() {
		return errors.New("cannot update archived company")
	}

	company.Name = params.Name

	if err = s.repo.Update(ctx, company); err != nil {
		return fmt.Errorf("s.repo.Update: %w", err)
	}

	for round, price := range params.Shares {
		if err = s.sharesRepo.Update(
			ctx,
			&models.CompanyShare{
				CompanyID: company.ID,
				Round:     round,
				Price:     price,
			},
		); err != nil {
			return fmt.Errorf("s.sharesRepo.Update: %w", err)
		}
	}

	return nil
}

func (s *Service) Archive(ctx context.Context, id int64) error {
	company, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("s.repo.GetByID: %w", err)
	}

	company.Archived = lo.ToPtr(true)

	if err = s.repo.Update(ctx, company); err != nil {
		return fmt.Errorf("s.repo.Update: %w", err)
	}
	return nil
}

func (s *Service) GetAllWithShares(ctx context.Context, onlyCurrentRound bool) ([]models.CompanyWithShares, error) {
	companies, err := s.repo.GetAllNotArchived(ctx)
	if err != nil {
		return nil, fmt.Errorf("s.repo.GetAllNotArchived: %w", err)
	}

	game, err := s.gameRepo.Get(ctx)
	if err != nil {
		return nil, fmt.Errorf("s.gameRepo.Get: %w", err)
	}

	shares, err := s.sharesRepo.GetAllActual(ctx)
	if err != nil {
		return nil, fmt.Errorf("s.sharesRepo.GetAllActual: %w", err)
	}

	sharesByCompanyID := make(map[int64]map[int]int64, len(companies))

	for _, share := range shares {
		if onlyCurrentRound && share.Round != game.CurrentRound {
			continue
		}

		priceByRound, ok := sharesByCompanyID[share.CompanyID]
		if !ok {
			priceByRound = map[int]int64{share.Round: share.Price}
			sharesByCompanyID[share.CompanyID] = priceByRound
			continue
		}
		priceByRound[share.Round] = share.Price
		sharesByCompanyID[share.CompanyID] = priceByRound
	}

	result := make([]models.CompanyWithShares, 0, len(companies))
	for _, company := range companies {
		companyShares := sharesByCompanyID[company.ID]
		result = append(
			result,
			models.CompanyWithShares{
				Company: company,
				Shares:  companyShares,
			},
		)
	}

	return result, nil
}
