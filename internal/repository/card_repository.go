package repository

import (
	"context"
	"database/sql"
	"errors"
	"go-banking-service/internal/database"
	"go-banking-service/internal/logger"
	"go-banking-service/internal/model"
)

type CardRepository interface {
	Create(ctx context.Context, card *model.Card) error
	GetByUserID(ctx context.Context, userID string) ([]*model.Card, error)
	GetByID(ctx context.Context, id string) (*model.Card, error)
}

type cardRepository struct {
	storage *database.Storage
}

func NewCardRepository(storage *database.Storage) CardRepository {
	return &cardRepository{storage: storage}
}

func (r *cardRepository) Create(ctx context.Context, card *model.Card) error {
	query := `INSERT INTO cards (id, user_id, account_id, number_encrypted, expiry_encrypted, cvv_hash, hmac, created_at) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`
	_, err := r.storage.DB.ExecContext(ctx, query, card.ID, card.UserID, card.AccountID, card.NumberEncrypted, card.ExpiryEncrypted, card.CVVHash, card.HMAC, card.CreatedAt)
	if err != nil {
		logger.Error("Failed to create card in DB", "error", err, "card_id", card.ID)
		return err
	}
	return nil
}

func (r *cardRepository) GetByUserID(ctx context.Context, userID string) ([]*model.Card, error) {
	rows, err := r.storage.DB.QueryContext(ctx, `SELECT id, user_id, account_id, number_encrypted, expiry_encrypted, cvv_hash, hmac, created_at FROM cards WHERE user_id = $1`, userID)
	if err != nil {
		logger.Error("Failed to get cards by user ID from DB", "error", err, "user_id", userID)
		return nil, err
	}
	defer rows.Close()

	var cards []*model.Card
	for rows.Next() {
		var card model.Card
		if err := rows.Scan(&card.ID, &card.UserID, &card.AccountID, &card.NumberEncrypted, &card.ExpiryEncrypted, &card.CVVHash, &card.HMAC, &card.CreatedAt); err != nil {
			logger.Error("Failed to scan card row", "error", err)
			return nil, err
		}
		cards = append(cards, &card)
	}
	if err := rows.Err(); err != nil {
		logger.Error("Error iterating over card rows", "error", err)
		return nil, err
	}
	return cards, nil
}

func (r *cardRepository) GetByID(ctx context.Context, id string) (*model.Card, error) {
	var card model.Card
	query := `SELECT id, user_id, account_id, number_encrypted, expiry_encrypted, cvv_hash, hmac, created_at FROM cards WHERE id = $1`
	err := r.storage.DB.QueryRowContext(ctx, query, id).Scan(&card.ID, &card.UserID, &card.AccountID, &card.NumberEncrypted, &card.ExpiryEncrypted, &card.CVVHash, &card.HMAC, &card.CreatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		logger.Error("Failed to get card by ID from DB", "error", err, "card_id", id)
		return nil, err
	}
	return &card, nil
}
