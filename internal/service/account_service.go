package service

import (
	"context"
	"errors"
	"fmt"
	"go-banking-service/internal/dto"
	"go-banking-service/internal/logger"
	"go-banking-service/internal/model"
	"go-banking-service/internal/repository"
	"time"
)

type AccountService interface {
	CreateAccount(ctx context.Context, userID string, req *dto.CreateAccountRequest) (*dto.AccountResponse, error)
	GetAccountsByUserID(ctx context.Context, userID string) ([]*dto.AccountResponse, error)
	GetAccountByID(ctx context.Context, accountID string, userID string) (*dto.AccountResponse, error)
	Deposit(ctx context.Context, accountID string, amount float64, userID string) error
	Withdraw(ctx context.Context, accountID string, amount float64, userID string) error
	PredictBalance(ctx context.Context, accountID string, userID string, days int) (*dto.PredictBalanceResponse, error)
}

type accountService struct {
	accountRepo         repository.AccountRepository
	creditRepo          repository.CreditRepository
	paymentScheduleRepo repository.PaymentScheduleRepository
}

func NewAccountService(accountRepo repository.AccountRepository, creditRepo repository.CreditRepository, paymentScheduleRepo repository.PaymentScheduleRepository) AccountService {
	return &accountService{
		accountRepo:         accountRepo,
		creditRepo:          creditRepo,
		paymentScheduleRepo: paymentScheduleRepo,
	}
}

func (s *accountService) CreateAccount(ctx context.Context, userID string, req *dto.CreateAccountRequest) (*dto.AccountResponse, error) {
	account := &model.Account{
		ID:        generateUUID(),
		UserID:    userID,
		Balance:   0.0,
		Currency:  "RUB",
		CreatedAt: time.Now(),
	}

	err := s.accountRepo.Create(ctx, account)
	if err != nil {
		logger.Error("Failed to create account", "error", err, "user_id", userID)
		return nil, fmt.Errorf("failed to create account: %w", err)
	}

	logger.Info("Account created successfully", "account_id", account.ID, "user_id", userID)

	return &dto.AccountResponse{
		ID:        account.ID,
		UserID:    account.UserID,
		Balance:   account.Balance,
		Currency:  account.Currency,
		CreatedAt: account.CreatedAt,
	}, nil
}

func (s *accountService) GetAccountsByUserID(ctx context.Context, userID string) ([]*dto.AccountResponse, error) {
	accounts, err := s.accountRepo.GetByUserID(ctx, userID)
	if err != nil {
		logger.Error("Failed to get accounts for user", "error", err, "user_id", userID)
		return nil, fmt.Errorf("failed to get accounts: %w", err)
	}

	var responses []*dto.AccountResponse
	for _, acc := range accounts {
		responses = append(responses, &dto.AccountResponse{
			ID:        acc.ID,
			UserID:    acc.UserID,
			Balance:   acc.Balance,
			Currency:  acc.Currency,
			CreatedAt: acc.CreatedAt,
		})
	}
	return responses, nil
}

func (s *accountService) GetAccountByID(ctx context.Context, accountID string, userID string) (*dto.AccountResponse, error) {
	account, err := s.accountRepo.GetByIDAndUserID(ctx, accountID, userID)
	if err != nil {
		logger.Error("Failed to get account by ID and user ID", "error", err, "account_id", accountID, "user_id", userID)
		return nil, err
	}

	return &dto.AccountResponse{
		ID:        account.ID,
		UserID:    account.UserID,
		Balance:   account.Balance,
		Currency:  account.Currency,
		CreatedAt: account.CreatedAt,
	}, nil
}

func (s *accountService) Deposit(ctx context.Context, accountID string, amount float64, userID string) error {
	account, err := s.accountRepo.GetByIDAndUserID(ctx, accountID, userID)
	if err != nil {
		logger.Error("Failed to get account for deposit", "error", err, "account_id", accountID, "user_id", userID)
		return err
	}

	newBalance := account.Balance + amount
	err = s.accountRepo.UpdateBalance(ctx, accountID, newBalance)
	if err != nil {
		logger.Error("Failed to update balance for deposit", "error", err, "account_id", accountID, "amount", amount)
		return fmt.Errorf("failed to update balance: %w", err)
	}

	logger.Info("Deposit successful", "account_id", accountID, "amount", amount, "new_balance", newBalance)
	return nil
}

func (s *accountService) Withdraw(ctx context.Context, accountID string, amount float64, userID string) error {
	account, err := s.accountRepo.GetByIDAndUserID(ctx, accountID, userID)
	if err != nil {
		logger.Error("Failed to get account for withdrawal", "error", err, "account_id", accountID, "user_id", userID)
		return err
	}

	if account.Balance < amount {
		logger.Warn("Insufficient funds for withdrawal", "account_id", accountID, "balance", account.Balance, "amount", amount)
		return errors.New("insufficient funds")
	}

	newBalance := account.Balance - amount
	err = s.accountRepo.UpdateBalance(ctx, accountID, newBalance)
	if err != nil {
		logger.Error("Failed to update balance for withdrawal", "error", err, "account_id", accountID, "amount", amount)
		return fmt.Errorf("failed to update balance: %w", err)
	}

	logger.Info("Withdrawal successful", "account_id", accountID, "amount", amount, "new_balance", newBalance)
	return nil
}

func (s *accountService) PredictBalance(ctx context.Context, accountID string, userID string, days int) (*dto.PredictBalanceResponse, error) {
	if days > 365 {
		days = 365
	} else if days <= 0 {
		days = 30
	}

	account, err := s.accountRepo.GetByIDAndUserID(ctx, accountID, userID)
	if err != nil {
		logger.Error("Failed to get account for prediction", "error", err, "account_id", accountID, "user_id", userID)
		return nil, err
	}

	credits, err := s.creditRepo.GetByUserID(ctx, userID)
	if err != nil {
		logger.Error("Failed to get credits", "error", err, "user_id", userID)
	}

	totalScheduledPayments := 0.0
	endDate := time.Now().AddDate(0, 0, days)

	for _, credit := range credits {
		if credit.AccountID == accountID && credit.Status == "active" {
			schedules, err := s.paymentScheduleRepo.GetByCreditID(ctx, credit.ID)
			if err == nil {
				for _, schedule := range schedules {
					if !schedule.IsPaid && schedule.PaymentDate.Before(endDate) {
						totalScheduledPayments += schedule.Amount
					}
				}
			}
		}
	}

	predictedBalance := account.Balance - totalScheduledPayments

	return &dto.PredictBalanceResponse{
		AccountID:        accountID,
		CurrentBalance:   account.Balance,
		PredictedBalance: predictedBalance,
		Days:             days,
	}, nil
}
