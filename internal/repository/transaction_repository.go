package repository

import (
	"context"
	"database/sql"
	"go-banking-service/internal/database"
	"go-banking-service/internal/logger"
	"go-banking-service/internal/model"
)

type TransactionRepository interface {
	Create(ctx context.Context, transaction *model.Transaction) error
	GetBySenderID(ctx context.Context, senderID string) ([]*model.Transaction, error)
	GetByReceiverID(ctx context.Context, receiverID string) ([]*model.Transaction, error)
	CreateTx(ctx context.Context, tx *sql.Tx, transaction *model.Transaction) error
}

type TransactionRepositoryImpl struct {
	Storage *database.Storage
}

func NewTransactionRepository(storage *database.Storage) TransactionRepository {
	return &TransactionRepositoryImpl{Storage: storage}
}

func (r *TransactionRepositoryImpl) Create(ctx context.Context, transaction *model.Transaction) error {
	query := `INSERT INTO transactions (id, sender_id, receiver_id, amount, type, description, created_at) VALUES ($1, $2, $3, $4, $5, $6, $7)`
	_, err := r.Storage.DB.ExecContext(ctx, query, transaction.ID, transaction.SenderID, transaction.ReceiverID, transaction.Amount, transaction.Type, transaction.Description, transaction.CreatedAt)
	if err != nil {
		logger.Error("Failed to create transaction in DB", "error", err, "transaction_id", transaction.ID)
		return err
	}
	return nil
}

func (r *TransactionRepositoryImpl) GetBySenderID(ctx context.Context, senderID string) ([]*model.Transaction, error) {
	rows, err := r.Storage.DB.QueryContext(ctx, `SELECT id, sender_id, receiver_id, amount, type, description, created_at FROM transactions WHERE sender_id = $1`, senderID)
	if err != nil {
		logger.Error("Failed to get transactions by sender ID from DB", "error", err, "sender_id", senderID)
		return nil, err
	}
	defer rows.Close()

	var transactions []*model.Transaction
	for rows.Next() {
		var transaction model.Transaction
		if err := rows.Scan(&transaction.ID, &transaction.SenderID, &transaction.ReceiverID, &transaction.Amount, &transaction.Type, &transaction.Description, &transaction.CreatedAt); err != nil {
			logger.Error("Failed to scan transaction row", "error", err)
			return nil, err
		}
		transactions = append(transactions, &transaction)
	}
	if err := rows.Err(); err != nil {
		logger.Error("Error iterating over transaction rows", "error", err)
		return nil, err
	}
	return transactions, nil
}

func (r *TransactionRepositoryImpl) GetByReceiverID(ctx context.Context, receiverID string) ([]*model.Transaction, error) {
	rows, err := r.Storage.DB.QueryContext(ctx, `SELECT id, sender_id, receiver_id, amount, type, description, created_at FROM transactions WHERE receiver_id = $1`, receiverID)
	if err != nil {
		logger.Error("Failed to get transactions by receiver ID from DB", "error", err, "receiver_id", receiverID)
		return nil, err
	}
	defer rows.Close()

	var transactions []*model.Transaction
	for rows.Next() {
		var transaction model.Transaction
		if err := rows.Scan(&transaction.ID, &transaction.SenderID, &transaction.ReceiverID, &transaction.Amount, &transaction.Type, &transaction.Description, &transaction.CreatedAt); err != nil {
			logger.Error("Failed to scan transaction row", "error", err)
			return nil, err
		}
		transactions = append(transactions, &transaction)
	}
	if err := rows.Err(); err != nil {
		logger.Error("Error iterating over transaction rows", "error", err)
		return nil, err
	}
	return transactions, nil
}

func (r *TransactionRepositoryImpl) CreateTx(ctx context.Context, tx *sql.Tx, transaction *model.Transaction) error {
	query := `INSERT INTO transactions (id, sender_id, receiver_id, amount, type, description, created_at) VALUES ($1, $2, $3, $4, $5, $6, $7)`
	_, err := tx.ExecContext(ctx, query, transaction.ID, transaction.SenderID, transaction.ReceiverID, transaction.Amount, transaction.Type, transaction.Description, transaction.CreatedAt)
	if err != nil {
		logger.Error("Failed to create transaction in DB (tx)", "error", err, "transaction_id", transaction.ID)
		return err
	}
	return nil
}
