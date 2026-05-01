package service

import (
	"context"
	"errors"
	"fmt"
	"go-banking-service/internal/dto"
	"go-banking-service/internal/logger"
	"go-banking-service/internal/model"
	"go-banking-service/internal/repository"
	"math"
	"time"

	"go-banking-service/internal/database"
)

type CreditService interface {
	CreateCredit(ctx context.Context, userID string, req *dto.CreateCreditRequest) (*dto.CreditResponse, error)
	GetCreditSchedule(ctx context.Context, creditID string, userID string) ([]*dto.PaymentScheduleResponse, error)
	ProcessScheduledPayments(ctx context.Context) error
}

type creditService struct {
	creditRepo          repository.CreditRepository
	paymentScheduleRepo repository.PaymentScheduleRepository
	accountRepo         repository.AccountRepository
	transactionRepo     repository.TransactionRepository
	userRepo            repository.UserRepository
	keyRateProvider     KeyRateProvider
	emailService        EmailService
	storage             *database.Storage
}

type KeyRateProvider interface {
	GetKeyRate() (float64, error)
}

func NewCreditService(
	creditRepo repository.CreditRepository,
	paymentScheduleRepo repository.PaymentScheduleRepository,
	accountRepo repository.AccountRepository,
	transactionRepo repository.TransactionRepository,
	userRepo repository.UserRepository,
	keyRateProvider KeyRateProvider,
	emailService EmailService,
	storage *database.Storage,
) CreditService {
	return &creditService{
		creditRepo:          creditRepo,
		paymentScheduleRepo: paymentScheduleRepo,
		accountRepo:         accountRepo,
		transactionRepo:     transactionRepo,
		userRepo:            userRepo,
		keyRateProvider:     keyRateProvider,
		emailService:        emailService,
		storage:             storage,
	}
}

func (s *creditService) CreateCredit(ctx context.Context, userID string, req *dto.CreateCreditRequest) (*dto.CreditResponse, error) {
	account, err := s.accountRepo.GetByIDAndUserID(ctx, req.AccountID, userID)
	if err != nil {
		logger.Error("Failed to get account for credit creation", "error", err, "account_id", req.AccountID, "user_id", userID)
		return nil, err
	}

	keyRate, err := s.keyRateProvider.GetKeyRate()
	if err != nil {
		logger.Error("Failed to get key rate for credit creation", "error", err)
		return nil, fmt.Errorf("failed to get key rate: %w", err)
	}

	interestRate := keyRate + 5.0

	monthlyRate := interestRate / 12.0 / 100.0
	numPayments := float64(req.TermMonths)

	monthlyPayment := req.Principal * (monthlyRate * math.Pow(1+monthlyRate, numPayments)) / (math.Pow(1+monthlyRate, numPayments) - 1)

	credit := &model.Credit{
		ID:             generateUUID(),
		UserID:         userID,
		AccountID:      account.ID,
		Principal:      req.Principal,
		InterestRate:   interestRate,
		TermMonths:     req.TermMonths,
		MonthlyPayment: monthlyPayment,
		RemainingDebt:  req.Principal,
		Status:         "active",
		CreatedAt:      time.Now(),
	}

	err = s.creditRepo.Create(ctx, credit)
	if err != nil {
		logger.Error("Failed to create credit in database", "error", err, "credit_id", credit.ID)
		return nil, fmt.Errorf("failed to create credit: %w", err)
	}

	schedules := make([]*model.PaymentSchedule, 0, req.TermMonths)
	currentDate := time.Now().AddDate(0, 1, 0)
	for i := 0; i < req.TermMonths; i++ {
		schedule := &model.PaymentSchedule{
			ID:             generateUUID(),
			CreditID:       credit.ID,
			PaymentDate:    currentDate,
			Amount:         monthlyPayment,
			IsPaid:         false,
			LateFeeApplied: false,
		}
		schedules = append(schedules, schedule)
		currentDate = currentDate.AddDate(0, 1, 0)
	}

	err = s.paymentScheduleRepo.CreateBatch(ctx, schedules)
	if err != nil {
		logger.Error("Failed to create payment schedules for credit", "error", err, "credit_id", credit.ID)
		return nil, fmt.Errorf("failed to create payment schedules: %w", err)
	}

	logger.Info("Credit created successfully", "credit_id", credit.ID, "user_id", userID, "principal", req.Principal)

	return &dto.CreditResponse{
		ID:             credit.ID,
		UserID:         credit.UserID,
		AccountID:      credit.AccountID,
		Principal:      credit.Principal,
		InterestRate:   credit.InterestRate,
		TermMonths:     credit.TermMonths,
		MonthlyPayment: credit.MonthlyPayment,
		RemainingDebt:  credit.RemainingDebt,
		Status:         credit.Status,
		CreatedAt:      credit.CreatedAt,
	}, nil
}

func (s *creditService) GetCreditSchedule(ctx context.Context, creditID string, userID string) ([]*dto.PaymentScheduleResponse, error) {
	credit, err := s.creditRepo.GetByID(ctx, creditID)
	if err != nil {
		logger.Error("Failed to get credit for schedule", "error", err, "credit_id", creditID)
		return nil, fmt.Errorf("failed to get credit: %w", err)
	}
	if credit == nil {
		return nil, errors.New("credit not found")
	}
	if credit.UserID != userID {
		logger.Warn("Access denied to credit schedule", "credit_id", creditID, "user_id", userID)
		return nil, errors.New("access denied")
	}

	schedules, err := s.paymentScheduleRepo.GetByCreditID(ctx, creditID)
	if err != nil {
		logger.Error("Failed to get payment schedules", "error", err, "credit_id", creditID)
		return nil, fmt.Errorf("failed to get payment schedules: %w", err)
	}

	var responses []*dto.PaymentScheduleResponse
	for _, schedule := range schedules {
		responses = append(responses, &dto.PaymentScheduleResponse{
			ID:             schedule.ID,
			CreditID:       schedule.CreditID,
			PaymentDate:    schedule.PaymentDate,
			Amount:         schedule.Amount,
			IsPaid:         schedule.IsPaid,
			LateFeeApplied: schedule.LateFeeApplied,
		})
	}
	return responses, nil
}

func (s *creditService) ProcessScheduledPayments(ctx context.Context) error {
	logger.Info("Starting scheduled payments processing")
	today := time.Now().Format("2006-01-02")
	unpaidSchedules, err := s.paymentScheduleRepo.GetUnpaidDue(ctx, today)
	if err != nil {
		logger.Error("Failed to get unpaid schedules", "error", err)
		return fmt.Errorf("failed to get unpaid schedules: %w", err)
	}

	processedCount := 0
	errorCount := 0

	for _, schedule := range unpaidSchedules {
		credit, err := s.creditRepo.GetByID(ctx, schedule.CreditID)
		if err != nil || credit == nil {
			logger.Error("Failed to get credit for scheduled payment", "error", err, "schedule_id", schedule.ID)
			errorCount++
			continue
		}

		account, err := s.accountRepo.GetByID(ctx, credit.AccountID)
		if err != nil || account == nil {
			logger.Error("Failed to get account for scheduled payment", "error", err, "schedule_id", schedule.ID)
			errorCount++
			continue
		}

		user, err := s.userRepo.GetByID(ctx, account.UserID)
		if err != nil || user == nil {
			logger.Error("Failed to get user for scheduled payment notification", "error", err, "account_id", account.ID)
			errorCount++
			continue
		}

		amountToPay := schedule.Amount
		lateFeeApplied := false
		penaltyAmount := 0.0

		if account.Balance < amountToPay {
			penaltyAmount = amountToPay * 0.10
			lateFeeApplied = true

			logger.Warn("Insufficient funds for scheduled payment, applying penalty",
				"schedule_id", schedule.ID,
				"account_id", credit.AccountID,
				"balance", account.Balance,
				"required", amountToPay,
				"penalty", penaltyAmount)

			newRemainingDebt := credit.RemainingDebt + penaltyAmount

			tx, err := s.storage.DB.BeginTx(ctx, nil)
			if err != nil {
				logger.Error("Failed to begin transaction for penalty application", "error", err, "schedule_id", schedule.ID)
				errorCount++
				continue
			}

			err = s.creditRepo.UpdateRemainingDebtTx(ctx, tx, credit.ID, newRemainingDebt)
			if err != nil {
				logger.Error("Failed to update remaining debt with penalty", "error", err, "schedule_id", schedule.ID)
				tx.Rollback()
				errorCount++
				continue
			}

			err = s.paymentScheduleRepo.UpdatePaidStatusTx(ctx, tx, schedule.ID, false, true)
			if err != nil {
				logger.Error("Failed to update paid status with penalty flag", "error", err, "schedule_id", schedule.ID)
				tx.Rollback()
				errorCount++
				continue
			}

			if err := tx.Commit(); err != nil {
				logger.Error("Failed to commit penalty transaction", "error", err, "schedule_id", schedule.ID)
				errorCount++
				continue
			}

			if err := s.emailService.SendPaymentNotification(user.Email, penaltyAmount); err != nil {
				logger.Error("Failed to send penalty notification email", "error", err, "user_email", user.Email)
			}

			errorCount++
			continue
		}

		tx, err := s.storage.DB.BeginTx(ctx, nil)
		if err != nil {
			logger.Error("Failed to begin transaction for scheduled payment", "error", err, "schedule_id", schedule.ID)
			errorCount++
			continue
		}

		err = s.accountRepo.UpdateBalanceTx(ctx, tx, credit.AccountID, account.Balance-amountToPay)
		if err != nil {
			logger.Error("Failed to update balance for scheduled payment", "error", err, "schedule_id", schedule.ID)
			tx.Rollback()
			errorCount++
			continue
		}

		newRemainingDebt := credit.RemainingDebt - schedule.Amount
		if newRemainingDebt < 0 {
			newRemainingDebt = 0
		}

		err = s.creditRepo.UpdateRemainingDebtTx(ctx, tx, credit.ID, newRemainingDebt)
		if err != nil {
			logger.Error("Failed to update remaining debt for scheduled payment", "error", err, "schedule_id", schedule.ID)
			tx.Rollback()
			errorCount++
			continue
		}

		if newRemainingDebt == 0 {
			err = s.creditRepo.UpdateStatusTx(ctx, tx, credit.ID, "closed")
			if err != nil {
				logger.Error("Failed to update credit status to closed", "error", err, "credit_id", credit.ID)
				tx.Rollback()
				errorCount++
				continue
			}
		}

		transaction := &model.Transaction{
			ID:          generateUUID(),
			SenderID:    credit.AccountID,
			ReceiverID:  "",
			Amount:      amountToPay,
			Type:        "credit_payment",
			Description: fmt.Sprintf("Credit payment for credit %s", credit.ID),
			CreatedAt:   time.Now(),
		}

		err = s.transactionRepo.CreateTx(ctx, tx, transaction)
		if err != nil {
			logger.Error("Failed to create transaction for scheduled payment", "error", err, "schedule_id", schedule.ID)
			tx.Rollback()
			errorCount++
			continue
		}

		err = s.paymentScheduleRepo.UpdatePaidStatusTx(ctx, tx, schedule.ID, true, lateFeeApplied)
		if err != nil {
			logger.Error("Failed to update paid status for scheduled payment", "error", err, "schedule_id", schedule.ID)
			tx.Rollback()
			errorCount++
			continue
		}

		if err := tx.Commit(); err != nil {
			logger.Error("Failed to commit transaction for scheduled payment", "error", err, "schedule_id", schedule.ID)
			errorCount++
			continue
		}

		if err := s.emailService.SendPaymentNotification(user.Email, amountToPay); err != nil {
			logger.Error("Failed to send payment notification email", "error", err, "user_email", user.Email)
		}

		processedCount++
		logger.Info("Scheduled payment processed successfully", "schedule_id", schedule.ID, "credit_id", credit.ID, "amount", amountToPay, "late_fee_applied", lateFeeApplied)
	}

	logger.Info("Finished scheduled payments processing", "processed", processedCount, "errors", errorCount)
	return nil
}
