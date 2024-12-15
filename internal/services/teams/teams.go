package teams

import (
	"context"
	"errors"
	"fmt"
	"github.com/rs/zerolog"
	"github.com/samber/lo"
	"investment-game-backend/internal/models"
	"investment-game-backend/internal/repo"
	"math/rand"
	"slices"
)

type Service struct {
	teamsRepo               repo.TeamsRepo
	balancesRepo            repo.BalancesRepo
	balanceTransactionsRepo repo.BalanceTransactionsRepo
	settingsRepo            repo.SettingsRepo
	additionalInfosRepo     repo.AdditionalInfosRepo
	sharesRepo              repo.CompanySharesRepo
	gamesRepo               repo.GamesRepo
	companiesRepo           repo.CompaniesRepo
	log                     *zerolog.Logger
	isTradePeriod           bool
	isRegistrationPeriod    bool
}

func New(
	teamsRepo repo.TeamsRepo,
	balancesRepo repo.BalancesRepo,
	settingsRepo repo.SettingsRepo,
	additionalInfosRepo repo.AdditionalInfosRepo,
	sharesRepo repo.CompanySharesRepo,
	balanceTransactionsRepo repo.BalanceTransactionsRepo,
	gamesRepo repo.GamesRepo,
	companiesRepo repo.CompaniesRepo,
	log *zerolog.Logger,
) *Service {
	return &Service{
		teamsRepo:               teamsRepo,
		balancesRepo:            balancesRepo,
		settingsRepo:            settingsRepo,
		additionalInfosRepo:     additionalInfosRepo,
		sharesRepo:              sharesRepo,
		balanceTransactionsRepo: balanceTransactionsRepo,
		gamesRepo:               gamesRepo,
		companiesRepo:           companiesRepo,
		log:                     log,
	}
}

type CreateParams struct {
	Name        string
	Credentials string
}

var ErrNoRegistrationPeriod = errors.New("cannot create team because is not registration period")

func (s *Service) Create(ctx context.Context, params CreateParams) (int64, error) {
	if !s.isRegistrationPeriod {
		s.log.Debug().Msg("cannot create team because is not registration period")
		return 0, ErrNoRegistrationPeriod
	}

	settings, err := s.settingsRepo.Get(ctx)
	if err != nil {
		return 0, fmt.Errorf("s.settingsRepo.Get: %w", err)
	}

	balanceID, err := s.balancesRepo.Create(
		ctx,
		&models.Balance{Amount: settings.DefaultBalanceAmount},
	)
	if err != nil {
		return 0, fmt.Errorf("s.balancesRepo.Create: %w", err)
	}

	game, err := s.gamesRepo.Get(ctx)
	if err != nil {
		return 0, fmt.Errorf("s.gamesRepo.Get: %w", err)
	}

	teamID, err := s.teamsRepo.Create(
		ctx,
		&models.Team{
			Name:        params.Name,
			Credentials: params.Credentials,
			BalanceID:   balanceID,
			GameID:      game.CurrentGame,
		},
	)
	if err != nil {
		return 0, fmt.Errorf("s.teamsRepo.Create: %w", err)
	}

	return teamID, nil
}

type UpdateParams struct {
	ID      int64
	Name    string
	Members []string
}

func (s *Service) Update(ctx context.Context, params UpdateParams) error {
	team, err := s.teamsRepo.GetByID(ctx, params.ID)
	if err != nil {
		return fmt.Errorf("s.teamsRepo.GetByID: %w", err)
	}

	team.Name = params.Name
	team.Members = params.Members

	if err = s.teamsRepo.Update(ctx, team); err != nil {
		return fmt.Errorf("s.repo.Update: %w", err)
	}
	return nil
}

type PurchaseParams struct {
	TeamID           int64
	SharesChanges    map[int64]int64
	AdditionalInfoID *int64
}

func (params PurchaseParams) Validate() error {
	if len(params.SharesChanges) == 0 && params.AdditionalInfoID == nil {
		return errors.New("purchase details are empty")
	}
	return nil
}

var (
	ErrIsNoTradePeriod        = errors.New("cannot do purchase because is not trade period")
	ErrIncorrectCountOfShares = errors.New("incorrect count of shares")
	ErrNoMoneyForOperation    = errors.New("insufficient balance to complete the transaction")
)

func (s *Service) Purchase(ctx context.Context, params PurchaseParams) (int64, error) {
	if !s.isTradePeriod {
		s.log.Debug().Msg("cannot do purchase because is not trade period")
		return 0, ErrIsNoTradePeriod
	}

	if err := params.Validate(); err != nil {
		return 0, fmt.Errorf("params.Validate: %w", err)
	}

	game, err := s.gamesRepo.Get(ctx)
	if err != nil {
		return 0, fmt.Errorf("s.gamesRepo.Get: %w", err)
	}
	team, err := s.teamsRepo.GetByID(ctx, params.TeamID)
	if err != nil {
		return 0, fmt.Errorf("s.teamsRepo.GetByID: %w", err)
	}
	balance, err := s.balancesRepo.GetByID(ctx, team.BalanceID)
	if err != nil {
		return 0, fmt.Errorf("s.balancesRepo.GetByID: %w", err)
	}

	if params.SharesChanges != nil {
		if team.Shares == nil {
			team.Shares = make(models.TeamSharesState)
		}
		if err = team.Shares.MergeChanges(params.SharesChanges); err != nil {
			if errors.Is(err, models.ErrSharesCountCannotBeNegative) {
				return 0, ErrIncorrectCountOfShares
			}
			return 0, fmt.Errorf("team.Shares.MergeChanges: %w", err)
		}
	}

	purchaseAmount, err := s.getPurchaseAmount(
		ctx,
		getPurchaseAmountParams{
			round:            game.CurrentRound,
			sharesChanges:    params.SharesChanges,
			additionalInfoID: params.AdditionalInfoID,
		},
	)
	if err != nil {
		return 0, fmt.Errorf("s.getPurchaseAmount: %w", err)
	}

	s.log.Trace().
		Int64("purchase_amount", purchaseAmount).
		Int64("team_id", team.ID).
		Str("team_name", team.Name).
		Int64("balance", balance.Amount).
		Msg("do purchase")

	if params.AdditionalInfoID != nil {
		if err = s.purchaseAdditionalInfo(
			ctx,
			purchaseAdditionalInfo{
				game:             game,
				balance:          balance,
				additionalInfoID: params.AdditionalInfoID,
				amount:           purchaseAmount,
			},
		); err != nil {
			return 0, fmt.Errorf("s.purchaseAdditionalInfo: %w", err)
		}
	} else {
		if err = s.purchaseShares(
			ctx,
			purchaseSharesParams{
				game:          game,
				balance:       balance,
				sharesChanges: params.SharesChanges,
				amount:        purchaseAmount,
				team:          team,
			},
		); err != nil {
			return 0, fmt.Errorf("s.purchaseShares: %w", err)
		}
	}

	if params.AdditionalInfoID != nil {
		team.AdditionalInfos = append(team.AdditionalInfos, *params.AdditionalInfoID)
	}
	if err = s.teamsRepo.Update(ctx, team); err != nil {
		return 0, fmt.Errorf("s.teamsRepo.Update: %w", err)
	}

	return balance.Amount, nil
}

type getPurchaseAmountParams struct {
	round            int
	sharesChanges    map[int64]int64
	additionalInfoID *int64
}

func (s *Service) getPurchaseAmount(ctx context.Context, params getPurchaseAmountParams) (int64, error) {
	if params.additionalInfoID != nil {
		additionalInfo, err := s.additionalInfosRepo.GetByID(ctx, *params.additionalInfoID)
		if err != nil {
			return 0, fmt.Errorf("s.additionalInfosRepo.GetByID: %w", err)
		}
		return additionalInfo.Cost, nil
	}

	companyIDs := lo.Keys(params.sharesChanges)
	shares, err := s.sharesRepo.GetListByCompanyIDsAndRound(ctx, companyIDs, params.round)
	if err != nil {
		return 0, fmt.Errorf("s.sharesRepo.GetListByCompanyIDsAndRound: %w", err)
	}
	priceByCompanyID := lo.SliceToMap(
		shares,
		func(share models.CompanyShare) (int64, int64) {
			return share.CompanyID, share.Price
		},
	)

	var amount int64
	for companyID, count := range params.sharesChanges {
		price := priceByCompanyID[companyID]
		amount += price * count
	}

	return amount, nil
}

type purchaseAdditionalInfo struct {
	game             *models.Game
	balance          *models.Balance
	additionalInfoID *int64
	amount           int64
}

func (s *Service) purchaseAdditionalInfo(ctx context.Context, params purchaseAdditionalInfo) error {
	if params.balance.Amount-params.amount < 0 {
		return ErrNoMoneyForOperation
	}

	_, err := s.balanceTransactionsRepo.Create(
		ctx,
		&models.BalanceTransaction{
			BalanceID:        params.balance.ID,
			Round:            params.game.CurrentRound,
			Amount:           params.amount,
			Details:          nil,
			AdditionalInfoID: params.additionalInfoID,
			RandomEventID:    nil,
		},
	)
	if err != nil {
		return fmt.Errorf("s.balanceTransactionsRepo.Create: %w", err)
	}

	params.balance.Amount -= params.amount
	if err = s.balancesRepo.Update(ctx, params.balance); err != nil {
		return fmt.Errorf("s.balancesRepo.Update: %w", err)
	}

	return nil
}

type purchaseSharesParams struct {
	game          *models.Game
	balance       *models.Balance
	sharesChanges map[int64]int64
	amount        int64
	team          *models.Team
}

func (s *Service) purchaseShares(ctx context.Context, params purchaseSharesParams) error {
	transaction, err := s.balanceTransactionsRepo.Get(ctx, params.balance.ID, params.game.CurrentRound)
	if err != nil && !errors.Is(err, repo.ErrNotFound) {
		return fmt.Errorf("s.balanceTransactionsRepo.Get: %w", err)
	}

	if errors.Is(err, repo.ErrNotFound) {
		if params.balance.Amount-params.amount < 0 {
			return ErrNoMoneyForOperation
		}

		if err = s.createNewBalanceTransaction(
			ctx,
			createNewBalanceTransactionParams{
				balance:       params.balance,
				game:          params.game,
				sharesChanges: params.sharesChanges,
				amount:        params.amount,
			},
		); err != nil {
			return fmt.Errorf("s.createNewBalanceTransaction: %w", err)
		}

		return nil
	} else {
		balanceAfterTransactionUpdate := params.balance.Amount + transaction.Amount - params.amount
		if balanceAfterTransactionUpdate < 0 {
			return ErrNoMoneyForOperation
		}

		for key, value := range transaction.Details {
			transaction.Details[key] = -value
		}
		if err = params.team.Shares.MergeChanges(transaction.Details); err != nil {
			return fmt.Errorf("params.team.Shares.MergeChanges: %w", err)
		}

		if err = s.updateExistedBalanceTransaction(
			ctx,
			updateExistedBalanceTransactionParams{
				balance:       params.balance,
				transaction:   transaction,
				sharesChanges: params.sharesChanges,
				amount:        params.amount,
			},
		); err != nil {
			return fmt.Errorf("s.updateExistedBalanceTransaction: %w", err)
		}

		return nil
	}
}

type createNewBalanceTransactionParams struct {
	game          *models.Game
	balance       *models.Balance
	sharesChanges map[int64]int64
	amount        int64
}

func (s *Service) createNewBalanceTransaction(
	ctx context.Context,
	params createNewBalanceTransactionParams,
) error {
	_, err := s.balanceTransactionsRepo.Create(
		ctx,
		&models.BalanceTransaction{
			BalanceID:        params.balance.ID,
			Round:            params.game.CurrentRound,
			Amount:           params.amount,
			Details:          params.sharesChanges,
			AdditionalInfoID: nil,
			RandomEventID:    nil,
		},
	)
	if err != nil {
		return fmt.Errorf("s.balanceTransactionsRepo.Create: %w", err)
	}

	params.balance.Amount -= params.amount
	if err = s.balancesRepo.Update(ctx, params.balance); err != nil {
		return fmt.Errorf("s.balancesRepo.Update: %w", err)
	}

	return nil
}

type updateExistedBalanceTransactionParams struct {
	balance       *models.Balance
	transaction   *models.BalanceTransaction
	sharesChanges map[int64]int64
	amount        int64
}

func (s *Service) updateExistedBalanceTransaction(
	ctx context.Context,
	params updateExistedBalanceTransactionParams,
) error {
	if err := s.balanceTransactionsRepo.Update(
		ctx,
		&models.BalanceTransaction{
			ID:               params.transaction.ID,
			BalanceID:        params.transaction.BalanceID,
			Round:            params.transaction.Round,
			Amount:           params.amount,
			Details:          params.sharesChanges,
			AdditionalInfoID: nil,
			RandomEventID:    nil,
		},
	); err != nil {
		return fmt.Errorf("s.balanceTransactionsRepo.Update: %w", err)
	}

	params.balance.Amount = params.balance.Amount + params.transaction.Amount - params.amount
	if err := s.balancesRepo.Update(ctx, params.balance); err != nil {
		return fmt.Errorf("s.balancesRepo.Update: %w", err)
	}

	return nil
}

func (s *Service) GetListForGame(ctx context.Context) ([]models.Team, error) {
	teams, err := s.teamsRepo.GetAllByGameID(ctx, 1)
	if err != nil {
		return nil, fmt.Errorf("s.teamsRepo.GetAllByGameID: %w", err)
	}
	return teams, nil
}

type DetailedTeam struct {
	Team                      *models.Team
	AdditionalInfos           []models.AdditionalInfo
	Balance                   int64
	HasTransactionInThisRound bool
}

func (s *Service) GetDetailedByID(ctx context.Context, id int64) (DetailedTeam, error) {
	team, err := s.teamsRepo.GetByID(ctx, id)
	if err != nil {
		return DetailedTeam{}, fmt.Errorf("s.teamsRepo.GetByID: %w", err)
	}

	balance, err := s.balancesRepo.GetByID(ctx, team.BalanceID)
	if err != nil {
		return DetailedTeam{}, fmt.Errorf("s.balancesRepo.GetByID: %w", err)
	}

	if err = s.fillTeamSharesByZeroValuesIfNeeded(ctx, team); err != nil {
		return DetailedTeam{}, fmt.Errorf("s.fillTeamSharesByZeroValuesIfNeeded: %w", err)
	}

	game, err := s.gamesRepo.Get(ctx)
	if err != nil {
		return DetailedTeam{}, fmt.Errorf("s.gamesRepo.Get: %w", err)
	}

	var hasTransactionInThisRound bool
	_, err = s.balanceTransactionsRepo.Get(ctx, team.BalanceID, game.CurrentRound)
	if err != nil && !errors.Is(err, repo.ErrNotFound) {
		return DetailedTeam{}, fmt.Errorf("s.balanceTransactionsRepo.Get: %w", err)
	}
	if err == nil {
		hasTransactionInThisRound = true
	}

	if len(team.AdditionalInfos) == 0 {
		return DetailedTeam{
			Team:                      team,
			AdditionalInfos:           nil,
			Balance:                   balance.Amount,
			HasTransactionInThisRound: hasTransactionInThisRound,
		}, nil
	}

	additionalInfos, err := s.additionalInfosRepo.GetByIDs(ctx, team.AdditionalInfos)
	if err != nil {
		return DetailedTeam{}, fmt.Errorf("s.additionalInfosRepo.GetByIDs: %w", err)
	}

	return DetailedTeam{
		Team:                      team,
		AdditionalInfos:           additionalInfos,
		Balance:                   balance.Amount,
		HasTransactionInThisRound: hasTransactionInThisRound,
	}, nil
}

func (s *Service) fillTeamSharesByZeroValuesIfNeeded(ctx context.Context, team *models.Team) error {
	companies, err := s.companiesRepo.GetAllNotArchived(ctx)
	if err != nil {
		return fmt.Errorf("s.companiesRepo.GetAllNotArchived: %w", err)
	}

	if len(team.Shares) == len(companies) {
		return nil
	}

	if team.Shares == nil {
		team.Shares = make(models.TeamSharesState)
	}

	sharesWithZeroValues := make(models.TeamSharesState)
	for _, company := range companies {
		sharesWithZeroValues[company.ID] = team.Shares[company.ID]
	}
	team.Shares = sharesWithZeroValues
	return nil
}

func (s *Service) NotifyTradePeriodUpdated(isTrade bool) {
	s.log.Trace().Bool("is_trade", isTrade).Msg("team service: NotifyTradePeriodUpdated")
	s.isTradePeriod = isTrade
}

func (s *Service) NotifyGameRegistrationPeriodUpdated(idRegistration bool) {
	s.log.Trace().Bool("is_registration", idRegistration).Msg("team service: NotifyGameRegistrationPeriodUpdated")
	s.isRegistrationPeriod = idRegistration
}

func (s *Service) GetAllForCurrentGame(ctx context.Context) ([]models.Team, error) {
	game, err := s.gamesRepo.Get(ctx)
	if err != nil {
		return nil, fmt.Errorf("s.gamesRepo.Get: %w", err)
	}

	teams, err := s.teamsRepo.GetAllByGameID(ctx, game.CurrentGame)
	if err != nil {
		return nil, fmt.Errorf("s.teamsRepo.GetAllByGameID: %w", err)
	}

	return teams, nil
}

var ErrNoAdditionalInfos = errors.New("no additional infos")

func (s *Service) PurchaseAdditionalInfoCompanyInfo(ctx context.Context, teamId int64) (models.AdditionalInfo, int64, error) {
	game, err := s.gamesRepo.Get(ctx)
	if err != nil {
		return models.AdditionalInfo{}, 0, fmt.Errorf("s.gamesRepo.Get: %w", err)
	}
	team, err := s.teamsRepo.GetByID(ctx, teamId)
	if err != nil {
		return models.AdditionalInfo{}, 0, fmt.Errorf("s.teamsRepo.GetByID: %w", err)
	}
	balance, err := s.balancesRepo.GetByID(ctx, team.BalanceID)
	if err != nil {
		return models.AdditionalInfo{}, 0, fmt.Errorf("s.balancesRepo.GetByID: %w", err)
	}

	additionalInfos, err := s.additionalInfosRepo.GetAllActualWithType(ctx, models.AdditionalInfoTypeCompanyInfo)
	if err != nil {
		return models.AdditionalInfo{}, 0, fmt.Errorf("s.additionalInfosRepo.GetAllActualWithType: %w", err)
	}

	additionalInfos = lo.Filter(additionalInfos, func(item models.AdditionalInfo, _ int) bool {
		return item.Round == game.CurrentRound
	})

	infosMap := lo.SliceToMap(
		additionalInfos,
		func(item models.AdditionalInfo) (int64, models.AdditionalInfo) {
			return item.ID, item
		},
	)
	for _, id := range team.AdditionalInfos {
		_, ok := infosMap[id]
		if ok {
			delete(infosMap, id)
		}
	}
	if len(infosMap) == 0 {
		return models.AdditionalInfo{}, 0, ErrNoAdditionalInfos
	}

	infoIds := lo.Keys(infosMap)
	additionalInfoToBuyID := rand.Intn(len(infoIds))
	additionalInfoToBuy := infosMap[infoIds[additionalInfoToBuyID]]

	if err = s.purchaseAdditionalInfo(
		ctx,
		purchaseAdditionalInfo{
			game:             game,
			balance:          balance,
			additionalInfoID: &additionalInfoToBuy.ID,
			amount:           additionalInfoToBuy.Cost,
		},
	); err != nil {
		return models.AdditionalInfo{}, 0, fmt.Errorf("s.purchaseAdditionalInfo: %w", err)
	}

	team.AdditionalInfos = append(team.AdditionalInfos, additionalInfoToBuy.ID)
	if err = s.teamsRepo.Update(ctx, team); err != nil {
		return models.AdditionalInfo{}, 0, fmt.Errorf("s.teamsRepo.Update: %w", err)
	}

	return additionalInfoToBuy, balance.Amount, nil
}

func (s *Service) ResetTransaction(ctx context.Context, teamID int64) (DetailedTeam, error) {
	game, err := s.gamesRepo.Get(ctx)
	if err != nil {
		return DetailedTeam{}, fmt.Errorf("s.gamesRepo.Get: %w", err)
	}
	team, err := s.teamsRepo.GetByID(ctx, teamID)
	if err != nil {
		return DetailedTeam{}, fmt.Errorf("s.teamsRepo.GetByID: %w", err)
	}
	balance, err := s.balancesRepo.GetByID(ctx, team.BalanceID)
	if err != nil {
		return DetailedTeam{}, fmt.Errorf("s.balancesRepo.GetByID: %w", err)
	}

	transaction, err := s.balanceTransactionsRepo.Get(ctx, balance.ID, game.CurrentRound)
	if err != nil {
		return DetailedTeam{}, fmt.Errorf("s.balanceTransactionsRepo.Get: %w", err)
	}
	balance.Amount = balance.Amount + transaction.Amount
	if err = s.balancesRepo.Update(ctx, balance); err != nil {
		return DetailedTeam{}, fmt.Errorf("s.balancesRepo.Update: %w", err)
	}

	for key, value := range transaction.Details {
		transaction.Details[key] = -value
	}
	if err = team.Shares.MergeChanges(transaction.Details); err != nil {
		return DetailedTeam{}, fmt.Errorf("params.team.Shares.MergeChanges: %w", err)
	}

	if err = s.teamsRepo.Update(ctx, team); err != nil {
		return DetailedTeam{}, fmt.Errorf("s.teamsRepo.Update: %w", err)
	}
	if err = s.balanceTransactionsRepo.Delete(ctx, balance.ID, game.CurrentRound); err != nil {
		return DetailedTeam{}, fmt.Errorf("s.balanceTransactionsRepo.Delete: %w", err)
	}

	detailedTeam, err := s.GetDetailedByID(ctx, teamID)
	if err != nil {
		return DetailedTeam{}, fmt.Errorf("s.GetDetailedByID: %w", err)
	}
	return detailedTeam, nil
}

type StatisticsByGame struct {
	Results []TeamResult `json:"results"`
}

type TeamResult struct {
	ID       int64  `json:"id"`
	TeamName string `json:"teamName"`
	Score    int64  `json:"score"`
}

func (s *Service) GetStatisticsByGame(ctx context.Context, round int) (StatisticsByGame, error) {
	game, err := s.gamesRepo.Get(ctx)
	if err != nil {
		return StatisticsByGame{}, fmt.Errorf("s.gamesRepo.Get: %w", err)
	}
	teams, err := s.teamsRepo.GetAllByGameID(ctx, game.CurrentGame)
	if err != nil {
		return StatisticsByGame{}, fmt.Errorf("s.teamsRepo.GetAllByGameID: %w", err)
	}
	if len(teams) == 0 {
		return StatisticsByGame{}, errors.New("no teams for current game")
	}

	companiesShares, err := s.sharesRepo.GetAllActual(ctx)
	if err != nil {
		return StatisticsByGame{}, fmt.Errorf("s.sharesRepo.GetAllActual: %w", err)
	}
	companiesSharesOnlyLastRound := lo.Filter(
		companiesShares,
		func(item models.CompanyShare, _ int) bool {
			return item.Round == round
		},
	)
	shareCostByCompanyID := lo.SliceToMap(
		companiesSharesOnlyLastRound,
		func(item models.CompanyShare) (int64, int64) {
			return item.CompanyID, item.Price
		},
	)

	statistics := StatisticsByGame{
		Results: make([]TeamResult, 0, len(teams)),
	}

	for _, team := range teams {
		score := int64(0)
		for companyId, count := range team.Shares {
			cost := shareCostByCompanyID[companyId]
			score += cost * count
		}
		statistics.Results = append(statistics.Results, TeamResult{
			ID:       team.ID,
			TeamName: team.Name,
			Score:    score,
		})
	}

	slices.SortFunc(statistics.Results, func(a, b TeamResult) int {
		if a.ID < b.ID {
			return -1
		}
		if a.ID > b.ID {
			return 1
		}
		return 0
	})

	slices.Reverse(statistics.Results)

	return statistics, nil
}
