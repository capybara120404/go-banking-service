package repository

import (
	"context"
	"database/sql"
	"errors"
	"go-banking-service/internal/database"
	"go-banking-service/internal/logger"
	"go-banking-service/internal/model"
)

type CreditRepository interface {
	Create(ctx context.Context, credit *model.Credit) error
	GetByID(ctx context.Context, id string) (*model.Credit, error)
	GetByUserID(ctx context.Context, userID string) ([]*model.Credit, error)
	UpdateRemainingDebt(ctx context.Context, id string, remainingDebt float64) error
	UpdateStatus(ctx context.Context, id string, status string) error
	UpdateRemainingDebtTx(ctx context.Context, tx *sql.Tx, id string, remainingDebt float64) error
	UpdateStatusTx(ctx context.Context, tx *sql.Tx, id string, status string) error
}

type creditRepository struct {
	storage *database.Storage
}

func NewCreditRepository(storage *database.Storage) CreditRepository {
	return &creditRepository{storage: storage}
}

func (r *creditRepository) Create(ctx context.Context, credit *model.Credit) error {
	query := `INSERT INTO credits (id, user_id, account_id, principal, interest_rate, term_months, monthly_payment, remaining_debt, status, created_at) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)`
	_, err := r.storage.DB.ExecContext(ctx, query, credit.ID, credit.UserID, credit.AccountID, credit.Principal, credit.InterestRate, credit.TermMonths, credit.MonthlyPayment, credit.RemainingDebt, credit.Status, credit.CreatedAt)
	if err != nil {
		logger.Error("Failed to create credit in DB", "error", err, "credit_id", credit.ID)
		return err
	}
	return nil
}

func (r *creditRepository) GetByID(ctx context.Context, id string) (*model.Credit, error) {
	var credit model.Credit
	query := `SELECT id, user_id, account_id, principal, interest_rate, term_months, monthly_payment, remaining_debt, status, created_at FROM credits WHERE id = $1`
	err := r.storage.DB.QueryRowContext(ctx, query, id).Scan(&credit.ID, &credit.UserID, &credit.AccountID, &credit.Principal, &credit.InterestRate, &credit.TermMonths, &credit.MonthlyPayment, &credit.RemainingDebt, &credit.Status, &credit.CreatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		logger.Error("Failed to get credit by ID from DB", "error", err, "credit_id", id)
		return nil, err
	}
	return &credit, nil
}

func (r *creditRepository) GetByUserID(ctx context.Context, userID string) ([]*model.Credit, error) {
	rows, err := r.storage.DB.QueryContext(ctx, `SELECT id, user_id, account_id, principal, interest_rate, term_months, monthly_payment, remaining_debt, status, created_at FROM credits WHERE user_id = $1`, userID)
	if err != nil {
		logger.Error("Failed to get credits by user ID from DB", "error", err, "user_id", userID)
		return nil, err
	}
	defer rows.Close()

	var credits []*model.Credit
	for rows.Next() {
		var credit model.Credit
		if err := rows.Scan(&credit.ID, &credit.UserID, &credit.AccountID, &credit.Principal, &credit.InterestRate, &credit.TermMonths, &credit.MonthlyPayment, &credit.RemainingDebt, &credit.Status, &credit.CreatedAt); err != nil {
			logger.Error("Failed to scan credit row", "error", err)
			return nil, err
		}
		credits = append(credits, &credit)
	}
	if err := rows.Err(); err != nil {
		logger.Error("Error iterating over credit rows", "error", err)
		return nil, err
	}
	return credits, nil
}

func (r *creditRepository) UpdateRemainingDebt(ctx context.Context, id string, remainingDebt float64) error {
	query := `UPDATE credits SET remaining_debt = $1 WHERE id = $2`
	_, err := r.storage.DB.ExecContext(ctx, query, remainingDebt, id)
	if err != nil {
		logger.Error("Failed to update remaining debt in DB", "error", err, "credit_id", id, "remaining_debt", remainingDebt)
		return err
	}
	return nil
}

func (r *creditRepository) UpdateStatus(ctx context.Context, id string, status string) error {
	query := `UPDATE credits SET status = $1 WHERE id = $2`
	_, err := r.storage.DB.ExecContext(ctx, query, status, id)
	if err != nil {
		logger.Error("Failed to update credit status in DB", "error", err, "credit_id", id, "status", status)
		return err
	}
	return nil
}

func (r *creditRepository) UpdateRemainingDebtTx(ctx context.Context, tx *sql.Tx, id string, remainingDebt float64) error {
	query := `UPDATE credits SET remaining_debt = $1 WHERE id = $2`
	_, err := tx.ExecContext(ctx, query, remainingDebt, id)
	if err != nil {
		logger.Error("Failed to update remaining debt in DB (tx)", "error", err, "credit_id", id, "remaining_debt", remainingDebt)
		return err
	}
	return nil
}

func (r *creditRepository) UpdateStatusTx(ctx context.Context, tx *sql.Tx, id string, status string) error {
	query := `UPDATE credits SET status = $1 WHERE id = $2`
	_, err := tx.ExecContext(ctx, query, status, id)
	if err != nil {
		logger.Error("Failed to update credit status in DB (tx)", "error", err, "credit_id", id, "status", status)
		return err
	}
	return nil
}
