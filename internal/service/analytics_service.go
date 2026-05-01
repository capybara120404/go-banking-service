package service

import (
	"context"
	"fmt"
	"go-banking-service/internal/dto"
	"go-banking-service/internal/logger"
	"go-banking-service/internal/repository"
	"time"
)

type AnalyticsService interface {
	GetAnalytics(ctx context.Context, userID string, days int) (*dto.AnalyticsResponse, error)
}

type analyticsService struct {
	transactionRepo     repository.TransactionRepository
	creditRepo          repository.CreditRepository
	accountRepo         repository.AccountRepository
	paymentScheduleRepo repository.PaymentScheduleRepository
}

func NewAnalyticsService(
	transactionRepo repository.TransactionRepository,
	creditRepo repository.CreditRepository,
	accountRepo repository.AccountRepository,
	paymentScheduleRepo repository.PaymentScheduleRepository,
) AnalyticsService {
	return &analyticsService{
		transactionRepo:     transactionRepo,
		creditRepo:          creditRepo,
		accountRepo:         accountRepo,
		paymentScheduleRepo: paymentScheduleRepo,
	}
}

func (s *analyticsService) GetAnalytics(ctx context.Context, userID string, days int) (*dto.AnalyticsResponse, error) {
	if days > 365 {
		days = 365
	} else if days <= 0 {
		days = 30
	}

	accounts, err := s.accountRepo.GetByUserID(ctx, userID)
	if err != nil {
		logger.Error("Failed to get accounts for analytics", "error", err, "user_id", userID)
		return nil, fmt.Errorf("failed to get accounts: %w", err)
	}

	var totalIncome float64
	var totalExpense float64

	startDate := time.Now().AddDate(0, 0, -days)

	for _, account := range accounts {
		transactions, err := s.transactionRepo.GetBySenderID(ctx, account.ID)
		if err != nil {
			logger.Error("Failed to get sender transactions for analytics", "error", err, "account_id", account.ID)
			continue
		}
		for _, t := range transactions {
			if t.CreatedAt.After(startDate) {
				totalExpense += t.Amount
			}
		}

		receivedTransactions, err := s.transactionRepo.GetByReceiverID(ctx, account.ID)
		if err != nil {
			logger.Error("Failed to get receiver transactions for analytics", "error", err, "account_id", account.ID)
			continue
		}
		for _, t := range receivedTransactions {
			if t.CreatedAt.After(startDate) {
				totalIncome += t.Amount
			}
		}
	}

	credits, err := s.creditRepo.GetByUserID(ctx, userID)
	if err != nil {
		logger.Error("Failed to get credits for analytics", "error", err, "user_id", userID)
		return nil, fmt.Errorf("failed to get credits: %w", err)
	}

	var totalCreditLoad float64
	for _, credit := range credits {
		if credit.Status == "active" {
			totalCreditLoad += credit.RemainingDebt
		}
	}

	var currentBalance float64
	for _, account := range accounts {
		currentBalance += account.Balance
	}

	endDate := time.Now().AddDate(0, 0, days)
	var totalScheduledPayments float64

	for _, credit := range credits {
		if credit.Status == "active" {
			schedules, err := s.paymentScheduleRepo.GetByCreditID(ctx, credit.ID)
			if err == nil {
				for _, schedule := range schedules {
					if !schedule.IsPaid && schedule.PaymentDate.Before(endDate) && schedule.PaymentDate.After(time.Now()) {
						totalScheduledPayments += schedule.Amount
					}
				}
			}
		}
	}

	predictedBalance := currentBalance - totalScheduledPayments

	logger.Info("Analytics generated", "user_id", userID, "days", days, "income", totalIncome, "expense", totalExpense)

	return &dto.AnalyticsResponse{
		Income:           totalIncome,
		Expense:          totalExpense,
		CreditLoad:       totalCreditLoad,
		PredictedBalance: predictedBalance,
	}, nil
}
