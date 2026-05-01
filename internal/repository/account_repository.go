package repository

import (
	"context"
	"database/sql"
	"errors"
	"go-banking-service/internal/database"
	"go-banking-service/internal/logger"
	"go-banking-service/internal/model"
)

type AccountRepository interface {
	Create(ctx context.Context, account *model.Account) error
	GetByID(ctx context.Context, id string) (*model.Account, error)
	GetByIDAndUserID(ctx context.Context, id string, userID string) (*model.Account, error)
	GetByUserID(ctx context.Context, userID string) ([]*model.Account, error)
	UpdateBalance(ctx context.Context, id string, balance float64) error
	UpdateBalanceTx(ctx context.Context, tx *sql.Tx, id string, balance float64) error
}

type AccountRepositoryImpl struct {
	Storage *database.Storage
}

func NewAccountRepository(storage *database.Storage) AccountRepository {
	return &AccountRepositoryImpl{Storage: storage}
}

func (r *AccountRepositoryImpl) Create(ctx context.Context, account *model.Account) error {
	query := `INSERT INTO accounts (id, user_id, balance, currency, created_at) VALUES ($1, $2, $3, $4, $5)`
	_, err := r.Storage.DB.ExecContext(ctx, query, account.ID, account.UserID, account.Balance, account.Currency, account.CreatedAt)
	if err != nil {
		logger.Error("Failed to create account in DB", "error", err, "account_id", account.ID)
		return err
	}
	return nil
}

func (r *AccountRepositoryImpl) GetByID(ctx context.Context, id string) (*model.Account, error) {
	var account model.Account
	query := `SELECT id, user_id, balance, currency, created_at FROM accounts WHERE id = $1`
	err := r.Storage.DB.QueryRowContext(ctx, query, id).Scan(&account.ID, &account.UserID, &account.Balance, &account.Currency, &account.CreatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		logger.Error("Failed to get account by ID from DB", "error", err, "account_id", id)
		return nil, err
	}
	return &account, nil
}

func (r *AccountRepositoryImpl) GetByIDAndUserID(ctx context.Context, id string, userID string) (*model.Account, error) {
	var account model.Account
	query := `SELECT id, user_id, balance, currency, created_at FROM accounts WHERE id = $1 AND user_id = $2`
	err := r.Storage.DB.QueryRowContext(ctx, query, id, userID).Scan(&account.ID, &account.UserID, &account.Balance, &account.Currency, &account.CreatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("access denied or account not found")
		}
		logger.Error("Failed to get account by ID and user ID from DB", "error", err, "account_id", id, "user_id", userID)
		return nil, err
	}
	return &account, nil
}

func (r *AccountRepositoryImpl) GetByUserID(ctx context.Context, userID string) ([]*model.Account, error) {
	rows, err := r.Storage.DB.QueryContext(ctx, `SELECT id, user_id, balance, currency, created_at FROM accounts WHERE user_id = $1`, userID)
	if err != nil {
		logger.Error("Failed to get accounts by user ID from DB", "error", err, "user_id", userID)
		return nil, err
	}
	defer rows.Close()

	var accounts []*model.Account
	for rows.Next() {
		var account model.Account
		if err := rows.Scan(&account.ID, &account.UserID, &account.Balance, &account.Currency, &account.CreatedAt); err != nil {
			logger.Error("Failed to scan account row", "error", err)
			return nil, err
		}
		accounts = append(accounts, &account)
	}
	if err := rows.Err(); err != nil {
		logger.Error("Error iterating over account rows", "error", err)
		return nil, err
	}
	return accounts, nil
}

func (r *AccountRepositoryImpl) UpdateBalance(ctx context.Context, id string, balance float64) error {
	query := `UPDATE accounts SET balance = $1 WHERE id = $2`
	_, err := r.Storage.DB.ExecContext(ctx, query, balance, id)
	if err != nil {
		logger.Error("Failed to update account balance in DB", "error", err, "account_id", id, "balance", balance)
		return err
	}
	return nil
}

func (r *AccountRepositoryImpl) UpdateBalanceTx(ctx context.Context, tx *sql.Tx, id string, balance float64) error {
	query := `UPDATE accounts SET balance = $1 WHERE id = $2`
	_, err := tx.ExecContext(ctx, query, balance, id)
	if err != nil {
		logger.Error("Failed to update account balance in DB (tx)", "error", err, "account_id", id, "balance", balance)
		return err
	}
	return nil
}
