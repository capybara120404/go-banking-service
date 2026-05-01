package repository

import (
	"context"
	"database/sql"
	"errors"
	"go-banking-service/internal/database"
	"go-banking-service/internal/logger"
	"go-banking-service/internal/model"

	"github.com/lib/pq"
)

var ErrUniqueViolation = errors.New("unique constraint violation")

type UserRepository interface {
	Create(ctx context.Context, user *model.User) error
	GetByEmail(ctx context.Context, email string) (*model.User, error)
	GetByUsername(ctx context.Context, username string) (*model.User, error)
	GetByID(ctx context.Context, id string) (*model.User, error)
}

type userRepository struct {
	storage *database.Storage
}

func NewUserRepository(storage *database.Storage) UserRepository {
	return &userRepository{storage: storage}
}

func (r *userRepository) Create(ctx context.Context, user *model.User) error {
	query := `INSERT INTO users (id, username, email, password_hash, created_at) VALUES ($1, $2, $3, $4, $5)`
	_, err := r.storage.DB.ExecContext(ctx, query, user.ID, user.Username, user.Email, user.PasswordHash, user.CreatedAt)
	if err != nil {
		if pqErr, ok := err.(*pq.Error); ok && pqErr.Code == "23505" {
			logger.Warn("Unique constraint violation in DB", "error", err, "user_id", user.ID)
			return ErrUniqueViolation
		}
		logger.Error("Failed to create user in DB", "error", err, "user_id", user.ID)
		return err
	}
	return nil
}

func (r *userRepository) GetByEmail(ctx context.Context, email string) (*model.User, error) {
	var user model.User
	query := `SELECT id, username, email, password_hash, created_at FROM users WHERE email = $1`
	err := r.storage.DB.QueryRowContext(ctx, query, email).Scan(&user.ID, &user.Username, &user.Email, &user.PasswordHash, &user.CreatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		logger.Error("Failed to get user by email from DB", "error", err, "email", email)
		return nil, err
	}
	return &user, nil
}

func (r *userRepository) GetByUsername(ctx context.Context, username string) (*model.User, error) {
	var user model.User
	query := `SELECT id, username, email, password_hash, created_at FROM users WHERE username = $1`
	err := r.storage.DB.QueryRowContext(ctx, query, username).Scan(&user.ID, &user.Username, &user.Email, &user.PasswordHash, &user.CreatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		logger.Error("Failed to get user by username from DB", "error", err, "username", username)
		return nil, err
	}
	return &user, nil
}

func (r *userRepository) GetByID(ctx context.Context, id string) (*model.User, error) {
	var user model.User
	query := `SELECT id, username, email, password_hash, created_at FROM users WHERE id = $1`
	err := r.storage.DB.QueryRowContext(ctx, query, id).Scan(&user.ID, &user.Username, &user.Email, &user.PasswordHash, &user.CreatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		logger.Error("Failed to get user by ID from DB", "error", err, "user_id", id)
		return nil, err
	}
	return &user, nil
}
